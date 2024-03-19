# stdout-test

## 简介

这是一个简单的Go服务器，它在8000端口监听并将请求体回显给客户端。它还将请求体附加到文件。

## 开始使用

### 先决条件

- Go 1.x

### 环境变量配置

这个应用程序使用以下环境变量来配置日志文件的行为：

- `LOG_DIR`: 日志文件的目录。默认值是`/logs`。
- `LOG_MAX_SIZE`: 日志文件的最大大小（以兆字节为单位）。默认值是10。
- `LOG_MAX_BACKUPS`: 保留的最大备份日志文件数量。默认值是5。
- `LOG_MAX_AGE`: 日志文件的最大存活时间（以天为单位）。默认值是28。
- `LOG_COMPRESS`: 是否压缩旧的日志文件。默认值是`false`。

### 安装

1. 克隆仓库
```sh
git clone <你的仓库URL>
```
2. 导航到项目目录
```sh
cd <你的项目名称>
```
3. 安装依赖
```sh
go mod download
```

### 使用

使用以下命令运行服务器：
```sh
go run server.go
```

## Docker支持

应用程序附带了一个Dockerfile，用于将应用程序容器化。Dockerfile使用多阶段构建过程来创建一个小的Docker镜像。

### 构建Docker镜像

要构建Docker镜像，导航到项目目录并运行：

```sh
docker build -t quzard/echo-server --no-cache .
```

### 运行Docker镜像

要运行Docker镜像，使用以下命令：

```sh
docker run -d -p 8000:8000 --name echo-server \
    -e LOG_DIR=/logs \
    -e LOG_MAX_SIZE=10 \
    -e LOG_MAX_BACKUPS=5 \
    -e LOG_MAX_AGE=28 \
    -e LOG_COMPRESS=false \
    -v ./logs:/logs \
    quzard/echo-server
```

这将启动应用程序并在8000端口上公开它。

## 使用

```bash
#!/bin/bash
if [ -z "$1" ]; then
    echo "Usage: $0 <message>"
    exit 1
fi

postData=$(printf '%s,%s,Exception in thread "main" java.lang.NullPointerException\n     at com.example.myproject.Book.getTitle\n     at com.example.myproject.Book.getTitle\n     at com.example.myproject.Book.getTitle\n    ...23 more' "$1" "$2")

curl -X POST \
     -d "$postData" \
     -H "Content-Type: text/plain" \
     http://127.0.0.1:8000
```