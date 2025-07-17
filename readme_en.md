<!-- xGate Logo -->
<p align="center">
  <img src="frontend/public/xgate.svg" alt="xGate Logo" width="96" height="96">
</p>

<p align="center">
  <a href="https://github.com/ipowerink/xgate-allinone/stargazers">
    <img src="https://img.shields.io/github/stars/ipowerink/xgate-allinone?style=social" alt="GitHub stars" />
  </a>
</p>

# xGate - A Modern Bastion Host

xGate is a lightweight, modern, and easy-to-use bastion host system built with Golang and Vue.js. It provides secure server access through a web interface, integrating web terminal and file management capabilities.

For detailed information about features, architecture, and database structure, please refer to the [Project Description Document](./doc/description_en.md).
![Example](./doc/img/feature.png)

## Features

- **Web SSH Terminal**: Securely access server terminals through your browser.
- **File Manager**: Support for uploading, downloading, browsing, and deleting server files.
- **User Management**: Multi-user system with administrator and regular user support, permission isolation.
- **Connection Management**: Support for adding, deleting, modifying, and querying SSH connections with group management.
- **Multi-tenant Isolation**: Regular users can only access their own connections, while administrators can manage all resources.
- **Group Display**: Connections can be grouped and displayed as collapsible panels in the frontend.
- **Flexible Configuration**: Support for SQLite (default) / MySQL, environment variables take priority, suitable for containerized deployment.
- **High-performance Caching**: Backend connection information memory cache, isolated by user, automatic expiration.

## Deployment Methods

You can deploy xGate using Docker, Docker Compose, or run it locally for development.

### 0. Single Docker Deployment (Recommended for Beginners)
```bash
docker run -d -p 8088:80 ipowerink/xgate-allinone
```
Access: http://localhost:8088

### 1. Docker Compose

This is the simplest way to start. This method will start the frontend, backend, and a MySQL database simultaneously.

**Requirements**: `docker` and `docker-compose` installed.

**Steps**:
1. Clone this repository.
2. Navigate to the project root directory.
3. Run the following command:
   ```bash
   docker-compose up -d
   ```
4. Access `http://localhost:8088` in your browser.

The `docker-compose.yml` file is pre-configured to use MySQL. All data will be persisted in Docker volumes.

### 2. Single Docker Container

You can run the entire application (frontend + backend) in a single Docker container. This setup uses the built-in SQLite database by default, with data persisted inside the container.

**Requirements**: `docker` installed.

**Steps**:
1. Clone this repository.
2. Build the image using the root directory's `Dockerfile`:
   ```bash
   docker build -t xgate-server .
   ```
3. Run the container:
   ```bash
   # Using SQLite (default)
   docker run -d -p 8088:8080 --name xgate-instance xgate-server

   # Or, connect to an external MySQL database
   docker run -d -p 8088:8080 --name xgate-instance \
     -e XGATE_DATABASE_TYPE=mysql \
     -e XGATE_DATABASE_HOST=<your-mysql-host> \
     -e XGATE_DATABASE_PORT=<your-mysql-port> \
     -e XGATE_DATABASE_USER=<your-mysql-username> \
     -e XGATE_DATABASE_PASSWORD=<your-mysql-password> \
     -e XGATE_DATABASE_NAME=<your-mysql-database> \
     xgate-server
   ```
4. Access `http://localhost:8088` in your browser.

### 3. Local Development

Suitable for developers who want to contribute code or run the service locally.

**Requirements**: `Go` (1.18+), `Node.js` (16+), and `pnpm` installed.

**Steps**:
1. **Start the Backend Service**:
   - Navigate to the `backend` directory.
   - Run the service. It will use SQLite (`xgate.db`) by default.
     ```bash
     cd backend
     go run main.go
     ```
   - The backend will run on `http://localhost:8080`.

2. **Start the Frontend Service**:
   - Navigate to the `frontend` directory.
   - Install dependencies and run the development server.
     ```bash
     cd frontend
     pnpm install
     pnpm dev
     ```
   - The frontend will run on `http://localhost:5173`. The development server is configured to proxy API requests to the backend.

## Environment Variables

Backend service configuration can be managed through environment variables.

| Variable Name              | Description                                                           | Default Value                        |
| -------------------------- | ---------------------------------------------------------------------- | ------------------------------------ |
| `XGATE_SERVER_PORT`        | Port for the backend service to listen on.                           | `8080`                               |
| `XGATE_JWT_SECRET`         | **(Required for production)** Secret key for signing JWT tokens. Please change to a long and random string. | `a_very_secret_key_for_local_dev`    |
| `XGATE_DATABASE_TYPE`      | Database type to use. Set to `mysql` to enable MySQL.                | `sqlite`                             |
| `XGATE_DATABASE_HOST`      | Hostname or IP address of the database server.                       | `localhost`                          |
| `XGATE_DATABASE_PORT`      | Port of the database server.                                          | `3306`                               |
| `XGATE_DATABASE_USER`      | Username for database connection.                                     | `root`                               |
| `XGATE_DATABASE_PASSWORD`  | Password for database connection.                                     | (empty string)                       |
| `XGATE_DATABASE_NAME`      | Name of the database to connect to.                                   | `xgate`                              |
| `XGATE_DATABASE_PATH`      | (SQLite only) Path to the SQLite database file.                      | `xgate.db`                           |

## Contributing

Welcome to contribute! You can participate by submitting Pull Requests or opening Issues.
