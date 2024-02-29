# 使用官方Go镜像作为构建环境
FROM golang:1.18 as builder

# 设置工作目录
WORKDIR /app

# 设置Go mod代理
ENV GOPROXY=https://goproxy.io,direct

# 复制源代码
COPY . .

# 编译Go程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用scratch作为基础镜像创建一个更小的镜像
FROM scratch

WORKDIR /app

# 从构建器中复制编译的程序
COPY --from=builder /app/main .

# 设置容器启动时运行的命令
CMD ["./main"]

# 暴露8000端口
EXPOSE 8000
