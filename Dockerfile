# 阶段 1：编译 Go 程序
FROM golang:1.21-alpine AS go-builder
WORKDIR /app
COPY main.go .
RUN go mod init fnshow || true && go mod tidy && go build -o fnshow main.go

# 阶段 2：运行原版 Python Web + 新版 Go 测试面板
FROM python:3.9-slim
WORKDIR /app

# 把 GitHub 上的所有原始文件拷进来
COPY . .
# 把刚才编译好的 Go 程序拷进来
COPY --from=go-builder /app/fnshow .

# 安装原项目的 Python 依赖 (如果出错也跳过，防止原项目依赖冲突导致容器起不来)
RUN pip install --no-cache-dir -r requirements.txt || true

# 创建一个启动脚本，让后台的 Go 和前台的 Python 同时运行
RUN echo '#!/bin/sh' > start.sh && \
    echo './fnshow &' >> start.sh && \
    echo 'python main.py' >> start.sh && \
    chmod +x start.sh

CMD ["./start.sh"]
