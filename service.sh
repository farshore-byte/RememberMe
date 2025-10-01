#!/bin/bash

# Farshore AI - 所有微服务启动脚本
# 支持启动/停止/重启全部或单个服务
# 日志存放在 logs/ 目录

set -e

echo "=== Farshore AI 微服务启动脚本 ==="

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 日志目录
LOG_DIR="./logs"
mkdir -p "$LOG_DIR"

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 配置文件路径
CONFIG_FILE="./remember/config.yaml"

# 读取配置函数
read_config() {
    local key=$1
    awk -v k="$key" '
    $1 == "server:" { in_server=1; next }
    in_server && $1 == k":" { print $2; exit }
    ' "$CONFIG_FILE"
}

# 读取端口配置
read_ports() {
    # 从config.yaml读取端口
    MAIN_PORT=$(read_config "main")
    SESSION_PORT=$(read_config "session_messages")
    USER_PORT=$(read_config "user_poritrait")
    TOPIC_PORT=$(read_config "topic_summary")
    EVENT_PORT=$(read_config "chat_event")
    OPENAI_PORT=$(read_config "openai")
    WEB_PORT=$(read_config "web")   
    # 如果端口为空，使用默认值
    MAIN_PORT=${MAIN_PORT:-6006}
    SESSION_PORT=${SESSION_PORT:-9120}
    USER_PORT=${USER_PORT:-9121}
    TOPIC_PORT=${TOPIC_PORT:-9122}
    EVENT_PORT=${EVENT_PORT:-9123}
    OPENAI_PORT=${OPENAI_PORT:-8344}
    WEB_PORT=${WEB_PORT:-8120}

}

# 检查项目根目录
check_project_root() {
    if [ ! -d "remember" ]; then
        log_error "请在项目根目录运行此脚本"
        exit 1
    fi
}

# 检查端口
check_port() {
    local port=$1
    local service=$2
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        log_warning "$service (端口 $port) 已被占用，跳过启动"
        return 1
    fi
    return 0
}

# 启动函数
start_main_service() {
    read_ports
    log_info "启动主服务 (端口 $MAIN_PORT)..."
    if check_port $MAIN_PORT "主服务"; then
        cd remember
        if [ ! -f "./server_main" ]; then
            log_error "可执行文件 server_main 不存在，请先编译"
            cd ..
            return 1
        fi
        ./server_main > ../$LOG_DIR/main.log 2>&1 &
        local pid=$!
        echo $pid > /tmp/farshore_main.pid
        # 检查进程是否还在运行
        sleep 1
        if ps -p $pid > /dev/null 2>&1; then
            log_success "主服务已启动，日志: $LOG_DIR/main.log"
            cd ..
            return 0
        else
            log_error "主服务启动失败，请检查日志: $LOG_DIR/main.log"
            rm -f /tmp/farshore_main.pid
            cd ..
            return 1
        fi
    else
        return 1
    fi
}

start_session_service() {
    read_ports
    log_info "启动会话服务 (端口 $SESSION_PORT)..."
    if check_port $SESSION_PORT "会话服务"; then
        cd remember
        if [ ! -f "./messages_main" ]; then
            log_error "可执行文件 messages_main 不存在，请先编译"
            cd ..
            return 1
        fi
        ./messages_main > ../$LOG_DIR/session_messages.log 2>&1 &
        local pid=$!
        echo $pid > /tmp/farshore_session.pid
        sleep 1
        if ps -p $pid > /dev/null 2>&1; then
            log_success "会话服务已启动，日志: $LOG_DIR/session_messages.log"
            cd ..
            return 0
        else
            log_error "会话服务启动失败，请检查日志: $LOG_DIR/session_messages.log"
            rm -f /tmp/farshore_session.pid
            cd ..
            return 1
        fi
    else
        return 1
    fi
}

start_user_portrait() {
    read_ports
    log_info "启动用户画像服务 (端口 $USER_PORT)..."
    if check_port $USER_PORT "用户画像服务"; then
        cd remember
        if [ ! -f "./user_main" ]; then
            log_error "可执行文件 user_main 不存在，请先编译"
            cd ..
            return 1
        fi
        ./user_main > ../$LOG_DIR/user_portrait.log 2>&1 &
        local pid=$!
        echo $pid > /tmp/farshore_user.pid
        sleep 1
        if ps -p $pid > /dev/null 2>&1; then
            log_success "用户画像服务已启动，日志: $LOG_DIR/user_portrait.log"
            cd ..
            return 0
        else
            log_error "用户画像服务启动失败，请检查日志: $LOG_DIR/user_portrait.log"
            rm -f /tmp/farshore_user.pid
            cd ..
            return 1
        fi
    else
        return 1
    fi
}

start_topic_summary() {
    read_ports
    log_info "启动话题摘要服务 (端口 $TOPIC_PORT)..."
    if check_port $TOPIC_PORT "话题摘要服务"; then
        cd remember
        if [ ! -f "./topic_main" ]; then
            log_error "可执行文件 topic_main 不存在，请先编译"
            cd ..
            return 1
        fi
        ./topic_main > ../$LOG_DIR/topic_summary.log 2>&1 &
        local pid=$!
        echo $pid > /tmp/farshore_topic.pid
        sleep 1
        if ps -p $pid > /dev/null 2>&1; then
            log_success "话题摘要服务已启动，日志: $LOG_DIR/topic_summary.log"
            cd ..
            return 0
        else
            log_error "话题摘要服务启动失败，请检查日志: $LOG_DIR/topic_summary.log"
            rm -f /tmp/farshore_topic.pid
            cd ..
            return 1
        fi
    else
        return 1
    fi
}

start_chat_event() {
    read_ports
    log_info "启动聊天事件服务 (端口 $EVENT_PORT)..."
    if check_port $EVENT_PORT "聊天事件服务"; then
        cd remember
        if [ ! -f "./event_main" ]; then
            log_error "可执行文件 event_main 不存在，请先编译"
            cd ..
            return 1
        fi
        ./event_main > ../$LOG_DIR/chat_event.log 2>&1 &
        local pid=$!
        echo $pid > /tmp/farshore_chat.pid
        sleep 1
        if ps -p $pid > /dev/null 2>&1; then
            log_success "聊天事件服务已启动，日志: $LOG_DIR/chat_event.log"
            cd ..
            return 0
        else
            log_error "聊天事件服务启动失败，请检查日志: $LOG_DIR/chat_event.log"
            rm -f /tmp/farshore_chat.pid
            cd ..
            return 1
        fi
    else
        return 1
    fi
}

start_openai_service() {
    read_ports
    log_info "启动OpenAI服务 (端口 $OPENAI_PORT)..."
    if check_port $OPENAI_PORT "OpenAI服务"; then
        cd remember
        ./openai_main > ../$LOG_DIR/openai.log 2>&1 &
        echo $! > /tmp/farshore_openai.pid
        log_success "OpenAI服务已启动，日志: $LOG_DIR/openai.log"
        cd ..
    fi
}

start_web_service() {
    read_ports
    log_info "启动Web服务 (端口 $WEB_PORT)..."
    if check_port $WEB_PORT "Web服务"; then
        cd remember-web
        npm run dev > ../$LOG_DIR/web.log 2>&1 &
        echo $! > /tmp/farshore_web.pid
        log_success "Web服务已启动，日志: $LOG_DIR/web.log"
        cd ..
    fi
}

# 停止函数 - 修复版本
stop_service_by_pid() {
    local pid_file=$1 name=$2 port=$3 process_name=$4
    local killed=false
    
    # 方法1: 通过PID文件停止
    if [ -f "$pid_file" ]; then
        pid=$(cat $pid_file)
        if ps -p $pid > /dev/null 2>&1; then
            log_info "停止$name (PID: $pid)"
            kill $pid 2>/dev/null && killed=true
            sleep 1
            # 如果进程还在，强制杀死
            if ps -p $pid > /dev/null 2>&1; then
                log_warning "强制停止$name (PID: $pid)"
                kill -9 $pid 2>/dev/null && killed=true
            fi
        fi
        rm -f $pid_file
    fi
    
    # 方法2: 通过端口停止 (无论是否有PID文件都执行)
    local port_pids=$(lsof -ti :$port 2>/dev/null || true)
    if [ ! -z "$port_pids" ]; then
        log_info "停止占用端口 $port 的进程: $port_pids"
        kill $port_pids 2>/dev/null && killed=true
        sleep 1
        # 检查是否还有进程占用端口，有则强制杀死
        port_pids=$(lsof -ti :$port 2>/dev/null || true)
        if [ ! -z "$port_pids" ]; then
            log_warning "强制停止占用端口 $port 的进程: $port_pids"
            kill -9 $port_pids 2>/dev/null && killed=true
        fi
    fi
    
    # 方法3: 通过进程名停止
    local process_pids=$(pgrep -f "$process_name" 2>/dev/null || true)
    if [ ! -z "$process_pids" ]; then
        log_info "停止进程名包含 '$process_name' 的进程: $process_pids"
        kill $process_pids 2>/dev/null && killed=true
        sleep 1
        process_pids=$(pgrep -f "$process_name" 2>/dev/null || true)
        if [ ! -z "$process_pids" ]; then
            log_warning "强制停止进程名包含 '$process_name' 的进程: $process_pids"
            kill -9 $process_pids 2>/dev/null && killed=true
        fi
    fi
    
    if [ "$killed" = true ]; then
        log_success "$name 已停止"
    fi
}

stop_main_service() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_main.pid" "主服务" $MAIN_PORT "server_main"; 
}
stop_session_service() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_session.pid" "会话服务" $SESSION_PORT "messages_main"; 
}
stop_user_portrait() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_user.pid" "用户画像服务" $USER_PORT "user_main"; 
}
stop_topic_summary() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_topic.pid" "话题摘要服务" $TOPIC_PORT "topic_main"; 
}
stop_chat_event() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_chat.pid" "聊天事件服务" $EVENT_PORT "event_main"; 
}
stop_openai_service() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_openai.pid" "OpenAI服务" $OPENAI_PORT "openai_main"; 
}
stop_web_service() { 
    read_ports
    stop_service_by_pid "/tmp/farshore_web.pid" "Web服务" $WEB_PORT "vite"; 
}

# 启动全部
start_all() {
    log_info "开始启动所有微服务..."
    check_project_root
    
    local failed_services=()
    
    start_main_service || failed_services+=("主服务")
    sleep 1
    start_session_service || failed_services+=("会话服务")
    sleep 1
    start_user_portrait || failed_services+=("用户画像服务")
    sleep 1
    start_topic_summary || failed_services+=("话题摘要服务")
    sleep 1
    start_chat_event || failed_services+=("聊天事件服务")
    sleep 1
    start_openai_service || failed_services+=("OpenAI服务")
    sleep 1
    start_web_service || failed_services+=("Web服务")
    sleep 2
    
    show_status
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        log_success "所有服务启动完成！"
    else
        log_error "以下服务启动失败: ${failed_services[*]}"
        log_error "请检查日志文件获取详细信息:"
        for service in "${failed_services[@]}"; do
            case $service in
                "主服务") log_error "  - $service: $LOG_DIR/main.log" ;;
                "会话服务") log_error "  - $service: $LOG_DIR/session_messages.log" ;;
                "用户画像服务") log_error "  - $service: $LOG_DIR/user_portrait.log" ;;
                "话题摘要服务") log_error "  - $service: $LOG_DIR/topic_summary.log" ;;
                "聊天事件服务") log_error "  - $service: $LOG_DIR/chat_event.log" ;;
                "OpenAI服务") log_error "  - $service: $LOG_DIR/openai.log" ;;
                "Web服务") log_error "  - $service: $LOG_DIR/web.log" ;;
            esac
        done
        log_error "可能的原因:"
        log_error "  - 可执行文件不存在 (请先运行 build_all.sh 编译)"
        log_error "  - 端口被占用"
        log_error "  - 依赖服务未启动"
        exit 1
    fi
}

# 停止全部
stop_all() {
    log_info "停止所有服务..."
    stop_main_service
    stop_session_service
    stop_user_portrait
    stop_topic_summary
    stop_chat_event
    stop_openai_service
    stop_web_service
    cleanup
    log_success "所有服务已停止"
}

# 清理PID文件
cleanup() { 
    rm -f /tmp/farshore_*.pid
    log_info "已清理PID文件"
}

# 显示状态
show_status() {
    echo ""
    log_info "=== 服务状态 ==="
    read_ports
    
    services=(
        "$MAIN_PORT:主服务:/tmp/farshore_main.pid:$LOG_DIR/main.log"
        "$SESSION_PORT:会话服务:/tmp/farshore_session.pid:$LOG_DIR/session_messages.log"
        "$USER_PORT:用户画像服务:/tmp/farshore_user.pid:$LOG_DIR/user_portrait.log"
        "$TOPIC_PORT:话题摘要服务:/tmp/farshore_topic.pid:$LOG_DIR/topic_summary.log"
        "$EVENT_PORT:聊天事件服务:/tmp/farshore_chat.pid:$LOG_DIR/chat_event.log"
        "$OPENAI_PORT:OpenAI服务:/tmp/farshore_openai.pid:$LOG_DIR/openai.log"
        "$WEB_PORT:Web服务:/tmp/farshore_web.pid:$LOG_DIR/web.log"
    )
    
    for service in "${services[@]}"; do
        IFS=':' read -r port name pid_file log_file <<< "$service"
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
            log_success "$name (端口 $port) - 运行中 ✅ 日志: $log_file"
        else
            log_error "$name (端口 $port) - 未运行 ❌ 日志: $log_file"
        fi
    done
}

# 服务映射函数
get_start_service() {
    local service=$1
    case $service in
        main) echo "start_main_service" ;;
        session) echo "start_session_service" ;;
        user) echo "start_user_portrait" ;;
        topic) echo "start_topic_summary" ;;
        event) echo "start_chat_event" ;;
        openai) echo "start_openai_service" ;;
        web) echo "start_web_service" ;;
        *) echo "" ;;
    esac
}

get_stop_service() {
    local service=$1
    case $service in
        main) echo "stop_main_service" ;;
        session) echo "stop_session_service" ;;
        user) echo "stop_user_portrait" ;;
        topic) echo "stop_topic_summary" ;;
        event) echo "stop_chat_event" ;;
        openai) echo "stop_openai_service" ;;
        web) echo "stop_web_service" ;;
        *) echo "" ;;
    esac
}

# 重启单个服务
restart_service() {
    local service=$1
    local start_func=$(get_start_service $service)
    local stop_func=$(get_stop_service $service)
    
    if [[ -n "$start_func" && -n "$stop_func" ]]; then
        log_info "重启 $service 服务..."
        $stop_func
        sleep 2
        $start_func
        log_success "$service 服务重启完成"
    else
        log_error "未知服务: $service"
        exit 1
    fi
}

# 主入口
case "${1:-}" in
    "start")
        if [ -z "$2" ]; then
            start_all
        else
            local start_func=$(get_start_service $2)
            if [[ -n "$start_func" ]]; then
                check_project_root
                $start_func
            else
                log_error "未知服务: $2"
                exit 1
            fi
        fi
        ;;
    "stop")
        if [ -z "$2" ]; then
            stop_all
        else
            local stop_func=$(get_stop_service $2)
            if [[ -n "$stop_func" ]]; then
                $stop_func
            else
                log_error "未知服务: $2"
                exit 1
            fi
        fi
        ;;
    "restart")
        if [ -z "$2" ]; then
            log_info "重启所有服务..."
            stop_all
            sleep 2
            start_all
        else
            check_project_root
            restart_service "$2"
        fi
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        echo "用法: $0 [command] [service]"
        echo ""
        echo "命令:"
        echo "  start     - 启动所有服务 或指定服务"
        echo "  stop      - 停止所有服务 或指定服务"
        echo "  restart   - 重启所有服务 或指定服务"
        echo "  status    - 显示服务状态"
        echo "  cleanup   - 清理PID文件"
        echo ""
        echo "服务:"
        echo "  main      - 主服务 (端口 $MAIN_PORT)"
        echo "  session   - 会话服务 (端口 $SESSION_PORT)"
        echo "  user      - 用户画像服务 (端口 $USER_PORT)"
        echo "  topic     - 话题摘要服务 (端口 $TOPIC_PORT)"
        echo "  event     - 聊天事件服务 (端口 $EVENT_PORT)"
        echo "  openai    - OpenAI服务 (端口 $OPENAI_PORT)"
        echo "  web       - Web服务 (端口 $WEB_PORT)"
        exit 1
        ;;
esac
