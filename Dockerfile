# 使用 Go 官方镜像作为基础镜像
FROM golang:1.24

# 设置工作目录
WORKDIR /app

# 复制项目文件到容器内
COPY . .

# 下载依赖
RUN go mod tidy

# 编译 Go 程序
RUN go build -o telegram-deepseek-bot main.go

# 设置运行环境变量（可选）
ENV TELEGRAM_BOT_TOKEN=""
ENV DEEPSEEK_TOKEN=""
ENV CUSTOM_URL=""
ENV DEEPSEEK_TYPE=""
ENV VOLC_AK=""
ENV VOLC_SK=""
ENV DB_TYPE=""
ENV DB_CONF=""

# 运行程序
CMD ["./telegram-deepseek-bot"]
