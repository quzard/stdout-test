# stdout-test

## 简介

这是一个多线程日志输出测试程序，支持将日志同时输出到标准输出(stdout)和文件。程序具有速率限制功能，可以控制日志输出速度，并支持日志文件的自动轮转。

## 开始使用

### 先决条件

- Go 1.x

### 环境变量配置

程序通过以下环境变量来控制行为：

#### 日志输出控制
- `SHOULD_PRINT`: 是否输出到stdout，设置为"on"开启
- `SHOULD_APPEND_TO_FILE`: 是否输出到文件，设置为"on"开启
- `PRINT_DIRECTLY`: 是否直接打印而不经过channel，设置为"on"开启
- `LOG`: 要输出的日志内容
- `LOG_WRITE_RATE_MB`: 日志写入速率限制(MB/s)，默认500MB/s

#### 运行参数
- `MINUTE`: 程序运行时间(分钟)，默认10分钟，设置为-1表示持续运行
- `THREAD`: 并发线程数，默认5

#### 日志文件配置
- `LOG_DIR`: 日志文件目录，默认为`/logs`
- `LOG_MAX_SIZE`: 单个日志文件最大大小(MB)，默认10240MB
- `LOG_MAX_BACKUPS`: 保留的最大日志文件数，默认10个
- `LOG_MAX_AGE`: 日志文件保留天数，默认1天

### 安装

1. 克隆仓库
```sh
git clone <仓库URL>
```

2. 进入项目目录
```sh
cd stdout-test
```

3. 安装依赖
```sh
go mod download
```

### 使用

基本运行命令：
```sh
go run server_stdout.go
```

使用自定义参数运行：
```sh
export SHOULD_PRINT=on
export SHOULD_APPEND_TO_FILE=on
export LOG="测试日志"
export THREAD=10
export MINUTE=5
go run server_stdout.go
```

## Docker支持

### 构建Docker镜像

```sh
docker build -t quzard/echo-server --no-cache .
```

### 运行Docker容器

```sh
docker run -d --name stdout-test \
    -e SHOULD_PRINT=on \
    -e SHOULD_APPEND_TO_FILE=on \
    -e LOG="测试日志" \
    -e THREAD=10 \
    -e MINUTE=5 \
    -e LOG_DIR=/logs \
    -e LOG_MAX_SIZE=10240 \
    -e LOG_MAX_BACKUPS=10 \
    -e LOG_MAX_AGE=1 \
    -e LOG_WRITE_RATE_MB=500 \
    -v ./logs:/logs \
    quzard/echo-server
```