FROM golang:1.23-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

# 安装基本构建工具
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 确保生成静态链接的二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-extldflags "-static"' -o main .

FROM alpine:latest

# 安装基本运行时依赖
RUN apk add --no-cache ca-certificates libc6-compat

WORKDIR /root

# 复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件目录
COPY --from=builder /app/config ./config

RUN chmod +x ./main

EXPOSE 8888

CMD ["./main"]