#!/bin/bash

# 记忆增强对话系统 - Linux安装脚本
# 适用于Ubuntu/CentOS等主流Linux发行版

set -e

echo "=== 记忆增强对话系统安装脚本 ==="
echo "将安装MongoDB、Redis并配置项目环境"

# 检测系统类型
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION=$VERSION_ID
else
    echo "无法检测操作系统类型"
    exit 1
fi

echo "检测到系统: $OS $VERSION"

# 安装MongoDB
install_mongodb() {
    echo "=== 安装MongoDB ==="
    
    case $OS in
        ubuntu|debian)
            # Ubuntu/Debian
            sudo apt update
            sudo apt install -y gnupg curl
            
            # 添加MongoDB官方仓库
            curl -fsSL https://pgp.mongodb.com/server-7.0.asc | \
               sudo gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg --dearmor
            
            echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | \
               sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
            
            sudo apt update
            sudo apt install -y mongodb-org
            
            # 启动MongoDB服务
            sudo systemctl start mongod
            sudo systemctl enable mongod
            ;;
            
        centos|rhel|fedora)
            # CentOS/RHEL/Fedora
            sudo yum install -y yum-utils
            
            # 添加MongoDB仓库
            sudo yum-config-manager --add-repo https://repo.mongodb.org/yum/redhat/mongodb-org-7.0.repo
            sudo yum install -y mongodb-org
            
            # 启动MongoDB服务
            sudo systemctl start mongod
            sudo systemctl enable mongod
            ;;
        *)
            echo "不支持的Linux发行版: $OS"
            exit 1
            ;;
    esac
    
    echo "MongoDB安装完成"
    mongod --version
}

# 安装Redis
install_redis() {
    echo "=== 安装Redis ==="
    
    case $OS in
        ubuntu|debian)
            sudo apt update
            sudo apt install -y redis-server
            ;;
            
        centos|rhel|fedora)
            sudo yum install -y epel-release
            sudo yum install -y redis
            ;;
        *)
            echo "不支持的Linux发行版: $OS"
            exit 1
            ;;
    esac
    
    # 启动Redis服务
    sudo systemctl start redis
    sudo systemctl enable redis
    
    echo "Redis安装完成"
    redis-server --version
}

# 安装Go（如果需要）
install_go() {
    echo "=== 安装Go ==="
    
    # 检查是否已安装Go
    if command -v go &> /dev/null; then
        echo "Go已安装，版本: $(go version)"
        return
    fi
    
    GO_VERSION="1.24.2"
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    
    # 下载并安装Go
    cd /tmp
    wget https://golang.org/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz
    
    # 设置环境变量
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    
    source ~/.bashrc
    
    echo "Go安装完成"
    go version
}

# 安装Node.js（用于前端）
install_nodejs() {
    echo "=== 安装Node.js ==="
    
    # 检查是否已安装Node.js
    if command -v node &> /dev/null; then
        echo "Node.js已安装，版本: $(node --version)"
        return
    fi
    
    case $OS in
        ubuntu|debian)
            curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
            sudo apt-get install -y nodejs
            ;;
            
        centos|rhel|fedora)
            curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
            sudo yum install -y nodejs
            ;;
        *)
            echo "不支持的Linux发行版: $OS"
            exit 1
            ;;
    esac
    
    echo "Node.js安装完成"
    node --version
    npm --version
}

# 配置防火墙
configure_firewall() {
    echo "=== 配置防火墙 ==="
    
    # 检查防火墙状态
    if command -v ufw &> /dev/null; then
        # Ubuntu ufw
        sudo ufw allow 27017/tcp  # MongoDB
        sudo ufw allow 6379/tcp   # Redis
        sudo ufw allow 8080/tcp   # 后端API
        sudo ufw allow 5173/tcp   # 前端开发服务器
        echo "防火墙规则已添加"
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS firewalld
        sudo firewall-cmd --permanent --add-port=27017/tcp
        sudo firewall-cmd --permanent --add-port=6379/tcp
        sudo firewall-cmd --permanent --add-port=8080/tcp
        sudo firewall-cmd --permanent --add-port=5173/tcp
        sudo firewall-cmd --reload
        echo "防火墙规则已添加"
    else
        echo "未检测到防火墙工具，请手动配置端口"
    fi
}

# 创建项目目录结构
setup_project() {
    echo "=== 设置项目目录 ==="
    
    # 创建数据目录
    sudo mkdir -p /data/db
    sudo chown -R mongodb:mongodb /data/db
    
    # 创建日志目录
    sudo mkdir -p /var/log/remember
    sudo chown -R $USER:$USER /var/log/remember
    
    echo "项目目录设置完成"
}

# 显示安装完成信息
show_completion() {
    echo ""
    echo "=== 安装完成 ==="
    echo ""
    echo "服务状态:"
    echo "MongoDB: $(sudo systemctl is-active mongod)"
    echo "Redis: $(sudo systemctl is-active redis)"
    echo ""
    echo "下一步操作:"
    echo "1. 进入项目目录: cd /path/to/remember-main"
    echo "2. 启动后端服务: cd remember && go run server_main.go"
    echo "3. 启动前端服务: cd remember-web && npm install && npm run dev"
    echo ""
    echo "默认端口:"
    echo "- 后端API: 8080"
    echo "- 前端开发服务器: 5173"
    echo "- MongoDB: 27017"
    echo "- Redis: 6379"
    echo ""
    echo "查看详细启动说明请阅读 README_START.md"
}

# 主安装流程
main() {
    echo "开始安装记忆增强对话系统..."
    
    # 更新系统包管理器
    case $OS in
        ubuntu|debian)
            sudo apt update
            ;;
        centos|rhel|fedora)
            sudo yum update -y
            ;;
    esac
    
    install_mongodb
    install_redis
    install_go
    install_nodejs
    configure_firewall
    setup_project
    show_completion
    
    echo "安装完成！"
}

# 执行主函数
main "$@"
