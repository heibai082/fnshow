# 使用轻量级 Go 镜像作为基础
FROM golang:1.21-alpine

# 设置工作目录
WORKDIR /app

# 设置国内代理，加速编译（如果是在国外服务器构建可以删掉这一行）
ENV GOPROXY=https://goproxy.cn,direct

# 将当前目录下的所有文件（包括 main.go）复制到容器中
COPY . .

# 直接编译 main.go，生成名为 fnshow 的可执行文件
# 我们不使用 go mod，避免因为缺少 go.mod 文件导致的构建失败
RUN go build -o fnshow main.go

# 赋予执行权限
RUN chmod +x fnshow

# 暴露端口：
# 5000: 接收飞牛 Webhook (播放/停止通知)
# 5001: 网页测试面板 (点击按钮发送通知)
EXPOSE 5000 5001

# 启动程序
CMD ["./fnshow"]
