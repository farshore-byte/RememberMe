#!/bin/bash

# build_all.sh - 编译所有 Go 服务
# 必须在项目根目录运行

set -e

echo "=== 编译所有微服务 ==="

SERVICES=(
    "server_main.go"
    "messages_main.go"
    "user_main.go"
    "topic_main.go"
    "event_main.go"
    "openai_main.go"
)

for service in "${SERVICES[@]}"; do
    dir="remember"
    base=$(basename "$service" .go)
    echo "进入目录 $dir 编译 $service ..."
    cd "$dir"
    go build -o "$base" "$service"
    echo "生成可执行文件：$dir/$base"
    cd ..
done

echo "=== 编译完成 ==="
