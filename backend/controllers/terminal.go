package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"xgate-backend/config"
	"xgate-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsMsg struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func TerminalWS(c *gin.Context) {
	id := c.Param("id")
	var connInfo models.Connection
	if err := config.DB.First(&connInfo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "连接信息未找到"})
		return
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer ws.Close()

	sshConfig := &ssh.ClientConfig{
		User:            connInfo.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(connInfo.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	addr := connInfo.Host
	if connInfo.Port != 0 {
		addr = addr + ":" + fmt.Sprint(connInfo.Port)
	}
	sshConn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("SSH连接失败: "+err.Error()))
		return
	}
	defer sshConn.Close()

	sess, err := sshConn.NewSession()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("创建SSH会话失败: "+err.Error()))
		return
	}
	defer sess.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sess.RequestPty("xterm", 32, 120, modes); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("请求Pty失败: "+err.Error()))
		return
	}
	stdin, _ := sess.StdinPipe()
	stdout, _ := sess.StdoutPipe()
	stderr, _ := sess.StderrPipe()

	if err := sess.Shell(); err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("启动Shell失败: "+err.Error()))
		return
	}

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}
			stdin.Write(msg)
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			break
		}
		ws.WriteMessage(websocket.TextMessage, buf[:n])
	}
	io.Copy(ws.UnderlyingConn(), stderr)
}
