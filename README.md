# stdout-test

## 简介

这是一个多线程日志输出测试程序，支持将日志同时输出到标准输出(stdout)和文件。程序具有速率限制功能，可以控制日志输出速度，并支持日志文件的自动轮转。

## 开始使用

### 先决条件

- Go

### 环境变量配置

程序通过以下环境变量来控制行为：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| MINUTE | 程序运行时间(分钟),设为-1表示一直运行 | 10 |
| THREAD | 并发线程数 | 5 |
| LOG | 日志内容 | - |
| SHOULD_PRINT | 是否输出到stdout(on/off) | off |
| SHOULD_APPEND_TO_FILE | 是否写入文件(on/off) | off |
| PRINT_DIRECTLY | 是否直接打印不带时间戳(on/off) | off |
| LOG_DIR | 日志文件目录 | /logs |
| LOG_MAX_SIZE | 单个日志文件最大大小(MB) | 10240 |
| LOG_MAX_BACKUPS | 日志文件备份数量 | 10 |
| LOG_MAX_AGE | 日志文件保留天数 | 1 |
| LOG_WRITE_RATE_MB | 写入速率限制(MB/s) | 500 |


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
# 设置环境变量
export MINUTE=10                   # 运行10分钟
export THREAD=5                    # 使用5个线程
export LOG="test log"              # 日志内容
export SHOULD_PRINT=on             # 输出到stdout
export SHOULD_APPEND_TO_FILE=on    # 写入文件
export LOG_DIR=/logs               # 日志目录
export LOG_WRITE_RATE_MB=500       # 写入速率限制(MB/s)
# 运行程序
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