# 使用官方 Go 镜像作为基础镜像
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /opt

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

# 下载依赖
RUN go mod download

# 复制所有源码到容器中
COPY . .

# 编译 Beego 项目，静态编译二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inspection-manager

# 使用精简的 Alpine 镜像作为运行环境
FROM cr.registry.res.cloud.zhejianglab.com/infrahi-installer/alpine:3.17 as prod

RUN apk update && apk add --no-cache ca-certificates bash expat net-tools iotop tcpdump tzdata curl
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /opt/

# 复制编译后的二进制文件
COPY --from=builder /opt/inspection-manager .

# 复制配置文件
COPY conf/app.conf /opt/conf/app.conf
COPY kube/ /opt/kube/

# 为二进制文件添加执行权限
RUN chmod +x /opt/inspection-manager

# 暴露应用的端口
EXPOSE 10086

# 启动应用
CMD ["./inspection-manager"]