# 使用多阶段构建来减小最终镜像大小
# 第一阶段：构建阶段
FROM golang:latest AS builder

# 设置工作目录
WORKDIR /app

# 将源代码复制到容器中
COPY . .

# 下载依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 创建 log 目录
RUN mkdir -p /app/log
RUN mkdir -p /app/images

# 复制config文件
COPY config /app/config
COPY sdconfig /app/sdconfig

# 赋予 app 目录下的所有文件和文件夹所有权限
RUN chmod -R 777 /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 设置环境变量
# ENV JUGG_TOOL_BOX_SERVICE_ENV=production
# ENV MY_DOMAIN=ec2-15-168-7-66.ap-northeast-3.compute.amazonaws.com
# ENV TZ=Asia/Shanghai

# 暴露应用端口（根据你的应用需求修改）
EXPOSE 8080

# 运行应用
CMD ["./main"]
# CMD ["tail", "-f", "/dev/null"]