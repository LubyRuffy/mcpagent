#!/bin/bash

# MCP Agent Web UI 开发启动脚本
# 同时启动前端开发服务器和后端API服务器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未安装或不在PATH中"
        exit 1
    fi
}

# 获取项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$PROJECT_ROOT/web"

print_info "MCP Agent Web UI 开发环境启动"
print_info "项目根目录: $PROJECT_ROOT"

# 检查必要的命令
print_info "检查开发环境..."
check_command "node"
check_command "npm"
check_command "go"

print_success "环境检查通过"

# 检查前端依赖
if [ ! -d "$WEB_DIR/node_modules" ]; then
    print_info "安装前端依赖..."
    cd "$WEB_DIR"
    npm install
    cd "$PROJECT_ROOT"
fi

# 创建日志目录
LOG_DIR="$PROJECT_ROOT/logs"
mkdir -p "$LOG_DIR"

# 清理函数
cleanup() {
    print_info "正在停止开发服务器..."
    
    # 杀死后台进程
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        print_info "后端服务器已停止"
    fi
    
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
        print_info "前端服务器已停止"
    fi
    
    print_success "开发环境已关闭"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

print_info "启动后端开发服务器..."
cd "$PROJECT_ROOT"
go build -o mcpagent-web cmd/mcpagent-web/main.go
./mcpagent-web -port 8081 > "$LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!

# 等待后端启动
sleep 3

# 检查后端是否启动成功
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    print_error "后端服务器启动失败"
    print_info "查看日志: cat $LOG_DIR/backend.log"
    exit 1
fi

print_success "后端服务器已启动 (PID: $BACKEND_PID)"
print_info "后端地址: http://localhost:8080"

print_info "启动前端开发服务器..."
cd "$WEB_DIR"
npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!

# 等待前端启动
sleep 5

# 检查前端是否启动成功
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    print_error "前端服务器启动失败"
    print_info "查看日志: cat $LOG_DIR/frontend.log"
    cleanup
    exit 1
fi

print_success "前端服务器已启动 (PID: $FRONTEND_PID)"
print_info "前端地址: http://localhost:3000"

print_success "开发环境启动完成！"
print_info ""
print_info "访问地址:"
print_info "  前端开发服务器: http://localhost:3000"
print_info "  后端API服务器:  http://localhost:8081"
print_info ""
print_info "日志文件:"
print_info "  后端日志: $LOG_DIR/backend.log"
print_info "  前端日志: $LOG_DIR/frontend.log"
print_info ""
print_info "按 Ctrl+C 停止所有服务器"

# 实时显示日志
print_info "实时日志输出 (后端):"
tail -f "$LOG_DIR/backend.log" &
TAIL_PID=$!

# 等待用户中断
wait $BACKEND_PID $FRONTEND_PID

# 清理
kill $TAIL_PID 2>/dev/null || true
cleanup
