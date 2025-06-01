#!/bin/bash

# MCP Agent Web UI 构建脚本
# 用于构建前端和后端的完整Web应用

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
BUILD_DIR="$PROJECT_ROOT/build"

print_info "MCP Agent Web UI 构建脚本"
print_info "项目根目录: $PROJECT_ROOT"

# 检查必要的命令
print_info "检查构建环境..."
check_command "node"
check_command "npm"
check_command "go"

# 检查Node.js版本
NODE_VERSION=$(node --version | cut -d'v' -f2)
NODE_MAJOR=$(echo $NODE_VERSION | cut -d'.' -f1)
if [ "$NODE_MAJOR" -lt 18 ]; then
    print_error "需要Node.js 18或更高版本，当前版本: $NODE_VERSION"
    exit 1
fi

print_success "环境检查通过"

# 创建构建目录
print_info "创建构建目录..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 构建前端
print_info "构建前端应用..."
cd "$WEB_DIR"

if [ ! -d "node_modules" ]; then
    print_info "安装前端依赖..."
    npm install
fi

print_info "执行前端构建..."
npm run build

if [ ! -d "dist" ]; then
    print_error "前端构建失败，dist目录不存在"
    exit 1
fi

print_success "前端构建完成"

# 复制前端构建产物
print_info "复制前端文件..."
cp -r "$WEB_DIR/dist" "$BUILD_DIR/web"

# 构建后端
print_info "构建后端应用..."
cd "$PROJECT_ROOT"

# 下载Go依赖
print_info "下载Go依赖..."
go mod download

# 构建后端二进制文件
print_info "编译后端二进制文件..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$BUILD_DIR/mcpagent-web" ./cmd/mcpagent-web
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o "$BUILD_DIR/mcpagent-web-darwin" ./cmd/mcpagent-web
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o "$BUILD_DIR/mcpagent-web.exe" ./cmd/mcpagent-web

print_success "后端构建完成"

# 复制配置文件
print_info "复制配置文件..."
cp "$PROJECT_ROOT/config_public.yaml" "$BUILD_DIR/"
cp "$PROJECT_ROOT/mcp_servers.json" "$BUILD_DIR/"

# 创建启动脚本
print_info "创建启动脚本..."

# Linux/macOS启动脚本
cat > "$BUILD_DIR/start.sh" << 'EOF'
#!/bin/bash

# MCP Agent Web UI 启动脚本

# 检测操作系统
OS="$(uname -s)"
case "${OS}" in
    Linux*)     BINARY="./mcpagent-web";;
    Darwin*)    BINARY="./mcpagent-web-darwin";;
    *)          echo "不支持的操作系统: ${OS}"; exit 1;;
esac

# 检查二进制文件是否存在
if [ ! -f "$BINARY" ]; then
    echo "错误: 找不到可执行文件 $BINARY"
    exit 1
fi

# 设置执行权限
chmod +x "$BINARY"

# 启动服务器
echo "启动 MCP Agent Web UI..."
echo "访问地址: http://localhost:8080"
echo "按 Ctrl+C 停止服务器"

exec "$BINARY" -config config_public.yaml "$@"
EOF

# Windows启动脚本
cat > "$BUILD_DIR/start.bat" << 'EOF'
@echo off
echo 启动 MCP Agent Web UI...
echo 访问地址: http://localhost:8080
echo 按 Ctrl+C 停止服务器

mcpagent-web.exe -config config_public.yaml %*
EOF

# 设置执行权限
chmod +x "$BUILD_DIR/start.sh"

# 创建README
print_info "创建部署说明..."
cat > "$BUILD_DIR/README.md" << 'EOF'
# MCP Agent Web UI - 部署包

这是MCP Agent Web UI的完整部署包，包含前端和后端的所有文件。

## 文件说明

- `mcpagent-web` - Linux版本的后端可执行文件
- `mcpagent-web-darwin` - macOS版本的后端可执行文件  
- `mcpagent-web.exe` - Windows版本的后端可执行文件
- `web/` - 前端静态文件目录
- `config_public.yaml` - 配置文件示例
- `mcp_servers.json` - MCP服务器配置文件
- `start.sh` - Linux/macOS启动脚本
- `start.bat` - Windows启动脚本

## 快速启动

### Linux/macOS
```bash
./start.sh
```

### Windows
```cmd
start.bat
```

### 手动启动
```bash
# Linux
./mcpagent-web -config config_public.yaml

# macOS  
./mcpagent-web-darwin -config config_public.yaml

# Windows
mcpagent-web.exe -config config_public.yaml
```

## 访问地址

启动后访问: http://localhost:8080

## 配置

编辑 `config_public.yaml` 文件来修改配置：
- LLM设置（API密钥、模型等）
- MCP服务器配置
- 系统提示词
- 其他参数

## 端口配置

默认端口为8080，可以通过参数修改：
```bash
./mcpagent-web -config config_public.yaml -port 9000
```

## 故障排除

1. 确保配置文件中的API密钥正确
2. 检查MCP服务器依赖是否安装（uvx, npx等）
3. 确认防火墙允许相应端口访问
4. 查看控制台输出的错误信息
EOF

# 计算文件大小
print_info "计算构建产物大小..."
BUILD_SIZE=$(du -sh "$BUILD_DIR" | cut -f1)

print_success "构建完成！"
print_info "构建目录: $BUILD_DIR"
print_info "构建大小: $BUILD_SIZE"
print_info ""
print_info "启动方法:"
print_info "  cd $BUILD_DIR"
print_info "  ./start.sh"
print_info ""
print_info "访问地址: http://localhost:8080"
