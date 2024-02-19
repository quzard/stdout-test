# stdout-test

## 简介

这是一个简单的Go服务器，它在8000端口监听并将请求体回显给客户端。它还将请求体附加到名为`output.txt`的文件。

## 开始使用

### 先决条件

- Go 1.x

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
docker build -t <你的镜像名称> .
```

这将创建一个带有标签`<你的镜像名称>`的Docker镜像。

### 运行Docker镜像

要运行Docker镜像，使用以下命令：

```sh
docker run -p 8000:8000 <你的镜像名称>
```

这将启动应用程序并在8000端口上公开它。