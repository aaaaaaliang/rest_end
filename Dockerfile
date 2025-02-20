FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN GOS=linux GOARCH=amd64 go build -o main .

# 使用较小的镜像作为运行时环境
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/
# 从 builder 镜像中复制编译好的程序
COPY --from=builder /app/main .

# 暴露容器的端口（根据实际需要修改）
EXPOSE 8080

# 设置容器启动时运行的命令
CMD ["./main"]