FROM cr.registry.res.cloud.zhejianglab.com/infrahi-installer/nvidiadriver:460.73.01-ubuntu20.04
#修改源地址
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo '$TZ' > /etc/timezone
RUN \
    sed -i "s@http://.*archive.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list && \
    sed -i "s@http://.*security.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list

#为了安装linux-tools-5.4.0-196-generi (这是bpftool的依赖)
RUN echo "deb http://security.ubuntu.com/ubuntu focal-security main" >> /etc/apt/sources.list
#依赖的基础软件

RUN apt-get update && apt-get install -y wget git make
RUN apt-get update && apt-get install -y ca-certificates bash expat net-tools iotop tcpdump tzdata curl telnet

# 1. 安装golang
RUN wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
RUN tar -C /usr/local -zxvf  go1.24.0.linux-amd64.tar.gz
RUN rm go1.24.0.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOROOT="/usr/local/go"

# 清理api缓存
RUN apt-get clean



