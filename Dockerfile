# 使用官方 Go 环境作为编译地
FROM golang:1.21-alpine

# 设置工作目录
WORKDIR /app

# 复制当前目录下所有文件
COPY . .

# 关键修正：手动初始化 Go 模块，确保编译不会报错
# 这样即使你仓库里没写 go.mod，它也能自己生成并编译
RUN go mod init fnshow && \
    go mod tidy && \
    go build -o fnshow main.go

# 赋予运行权限
RUN chmod +x fnshow

# 暴露端口：5000(Webhook), 5001(测试面板)
EXPOSE 5000 5001

# 启动
CMD ["./fnshow"]
