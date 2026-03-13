# 第一阶段：编译环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置代理（国内环境编译更快，如果不需要可以删掉）
ENV GOPROXY=https://goproxy.cn,direct

# 复制源代码并初始化项目（强制覆盖旧的 mod）
COPY main.go .
RUN go mod init fnshow || true && \
    go mod tidy && \
    go build -o fnshow main.go

# 第二阶段：运行环境
FROM alpine:latest

# 安装基础证书（确保能发送 https 请求到推送接口）
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从编译阶段拷贝生成的程序
COPY --from=builder /app/fnshow .

# 暴露端口：
# 5000: 接收飞牛 Webhook (播放/停止)
# 5001: 网页测试面板
EXPOSE 5000 5001

# 启动程序
CMD ["./fnshow"]
