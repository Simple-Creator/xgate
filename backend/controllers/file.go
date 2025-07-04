package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jump-backend/config"
	"jump-backend/models"

	"fmt"

	"encoding/base64"

	"io"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Helper function to get a connection with an ownership check
func getConnectionForUser(c *gin.Context) (*models.Connection, error) {
	userID, role := getUserInfo(c)
	id := c.Param("id")

	var conn models.Connection
	db := config.DB
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}

	if err := db.First(&conn, id).Error; err != nil {
		return nil, err
	}
	return &conn, nil
}

// Helper to create ssh.Client
func newSSHClient(conn *models.Connection) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: conn.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), config)
}

// 通过SFTP获取远程文件列表
func ListFiles(c *gin.Context) {
	connInfo, err := getConnectionForUser(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到或无权限"})
		return
	}
	path := c.Query("path")
	if path == "" {
		path = "."
	}
	offset := 0
	limit := 100
	if v := c.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}
	if v := c.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
		if limit <= 0 {
			limit = 100
		}
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	entries, err := sftpClient.ReadDir(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "读取目录失败: " + err.Error()})
		return
	}
	total := len(entries)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}
	files := []gin.H{}
	for _, entry := range entries[offset:end] {
		fileType := "file"
		if entry.IsDir() {
			fileType = "dir"
		} else if entry.Mode()&os.ModeSymlink != 0 {
			fileType = "link"
		}
		files = append(files, gin.H{
			"name":    entry.Name(),
			"path":    filepath.Join(path, entry.Name()),
			"type":    fileType,
			"size":    entry.Size(),
			"modTime": entry.ModTime().Format("2006-01-02 15:04"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "total": total, "offset": offset, "limit": limit})
}

// 通过SFTP获取远程主机家目录
func GetHomeDir(c *gin.Context) {
	connInfo, err := getConnectionForUser(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到或无权限"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	home, err := sftpClient.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取家目录失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"home": home})
}

// 路径转义
func escapePath(path string) string {
	return `'` + strings.ReplaceAll(path, `'`, `'\''`) + `'`
}

// 解析ls -l输出（size直接用parts[4]，无需正则）
func parseLsOutput(output, parentPath string) []gin.H {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := []gin.H{}
	if len(lines) <= 1 && strings.Contains(lines[0], "total 0") {
		return files
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "total") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 8 {
			continue
		}

		fileType := "file"
		if parts[0][0] == 'd' {
			fileType = "dir"
		} else if parts[0][0] == 'l' {
			fileType = "link"
		}

		nameIndex := 7
		for i := 7; i < len(parts); i++ {
			if strings.Contains(parts[i], ":") { // Look for time stamp
				nameIndex = i + 1
				break
			}
		}

		name := strings.Join(parts[nameIndex:], " ")

		files = append(files, gin.H{
			"name":    name,
			"path":    filepath.Join(parentPath, name),
			"type":    fileType,
			"size":    parts[4],
			"modTime": parts[5] + " " + parts[6],
		})
	}
	return files
}

// 上传文件
func UploadFile(c *gin.Context) {
	path := c.PostForm("path")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "未选择文件"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "文件打开失败"})
		return
	}
	defer file.Close()

	connInfo, err := getConnectionForUser(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到或无权限"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()

	remotePath := filepath.Join(path, fileHeader.Filename)
	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "远程文件创建失败: " + err.Error()})
		return
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "文件上传失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "上传成功"})
}

// 下载文件
func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "缺少文件路径"})
		return
	}
	// 查找连接信息
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	f, err := sftpClient.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "远程文件打开失败: " + err.Error()})
		return
	}
	defer f.Close()
	filename := filepath.Base(path)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, f)
}

// 删除文件
func DeleteFile(c *gin.Context) {
	id := c.Param("id")
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "缺少文件路径"})
		return
	}
	// 查找连接信息
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	if err := sftpClient.Remove(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "远程删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 通过SFTP重命名远程文件
func RenameFile(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	if err := sftpClient.Rename(req.OldPath, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "重命名失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "重命名成功"})
}

// 编辑文件内容（通过SFTP写入远程文件，内容用base64编码防止特殊字符丢失）
func EditFile(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	// 查找连接信息
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	f, err := sftpClient.Create(req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "远程文件创建失败: " + err.Error()})
		return
	}
	defer f.Close()
	data, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "内容解码失败: " + err.Error()})
		return
	}
	if _, err := f.Write(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "写入文件失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "保存成功"})
}

// 读取远程文件内容（通过SFTP，内容base64编码返回）
func ReadFile(c *gin.Context) {
	id := c.Param("id")
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "缺少文件路径"})
		return
	}
	// 查找连接信息
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", connInfo.Host, connInfo.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "SSH连接失败: " + err.Error()})
		return
	}
	defer client.Close()
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "SFTP创建失败: " + err.Error()})
		return
	}
	defer sftpClient.Close()
	f, err := sftpClient.Open(path)
	if err != nil {
		log.Printf("[ReadFile] 远程文件打开失败: %s, 路径: %s", err.Error(), path)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "远程文件打开失败: " + err.Error()})
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		log.Printf("[ReadFile] Stat 失败: %s, 路径: %s", err.Error(), path)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取文件信息失败: " + err.Error()})
		return
	}
	if stat.Size() > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "文件过大（>10MB），请下载后查看"})
		return
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Printf("[ReadFile] 读取文件失败: %s, 路径: %s", err.Error(), path)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "读取文件失败: " + err.Error()})
		return
	}
	b64 := base64.StdEncoding.EncodeToString(content)
	c.JSON(http.StatusOK, gin.H{"content": b64})
}
