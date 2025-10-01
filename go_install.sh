#!/bin/bash
set -e

INSTALL_DIR=/usr/local
PROFILE=/etc/profile

# 阿里云镜像
MIRROR=https://mirrors.aliyun.com/golang

# 自动获取最新版本号
LATEST=$(wget -qO- https://golang.google.cn/VERSION?m=text | head -n1)
GO_TARBALL=${LATEST}.linux-amd64.tar.gz

echo ">>> 正在下载 Go ${LATEST} ..."
wget -c ${MIRROR}/${GO_TARBALL} -O /tmp/${GO_TARBALL}

echo ">>> 解压 Go 到 ${INSTALL_DIR} ..."
rm -rf ${INSTALL_DIR}/go
tar -C ${INSTALL_DIR} -xzf /tmp/${GO_TARBALL}

# 配置环境变量
if ! grep -q "export GOROOT=" $PROFILE; then
    echo "export GOROOT=${INSTALL_DIR}/go" >> $PROFILE
    echo "export GOPATH=\$HOME/go" >> $PROFILE
    echo "export PATH=\$PATH:\$GOROOT/bin:\$GOPATH/bin" >> $PROFILE
fi

echo ">>> 安装完成！请运行以下命令使环境变量生效："
echo "source $PROFILE"
echo ">>> 然后验证版本："
echo "go version"
