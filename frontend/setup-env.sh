#!/bin/bash

# 输出彩色文本的函数
print_blue() {
  echo -e "\033[34m$1\033[0m"
}

print_green() {
  echo -e "\033[32m$1\033[0m"
}

print_yellow() {
  echo -e "\033[33m$1\033[0m"
}

print_red() {
  echo -e "\033[31m$1\033[0m"
}

# 显示脚本开始标志
print_blue "==================================================="
print_blue "       前端环境完整初始化脚本"
print_blue "==================================================="

# 检查是否为root用户
if [ "$(id -u)" != "0" ]; then
   print_red "此脚本需要root权限，请使用sudo运行"
   exit 1
fi

print_yellow "开始环境初始化..."

# 更新系统包
print_yellow "更新系统包..."
apt update -y
apt upgrade -y

# 安装基础工具
print_yellow "安装基础工具..."
apt install -y curl wget git build-essential

# 安装NVM (Node Version Manager)
print_yellow "安装NVM..."
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# 加载NVM
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

# 如果NVM没有正确加载，提示用户
if ! command -v nvm &> /dev/null; then
  print_red "NVM安装后未能正确加载，请尝试关闭终端并重新打开，然后重新运行此脚本"
  print_yellow "或者手动运行以下命令："
  print_yellow "export NVM_DIR=\"\$HOME/.nvm\""
  print_yellow "[ -s \"\$NVM_DIR/nvm.sh\" ] && \. \"\$NVM_DIR/nvm.sh\""
  exit 1
fi

# 安装Node.js 16（稳定版，兼容性好）
print_yellow "安装Node.js 16..."
nvm install 16
nvm use 16
nvm alias default 16

# 验证Node.js和npm安装
NODE_VERSION=$(node -v)
NPM_VERSION=$(npm -v)
print_green "Node.js版本: $NODE_VERSION"
print_green "npm版本: $NPM_VERSION"

# 更新npm到最新版本
print_yellow "更新npm到最新版本..."
npm install -g npm@latest

# 安装常用的全局npm包
print_yellow "安装常用全局npm包..."
npm install -g @vue/cli
npm install -g serve

# 切换到前端目录
cd "$(dirname "$0")"
print_yellow "当前工作目录: $(pwd)"

# 清理旧的依赖
print_yellow "清理旧的依赖..."
rm -rf node_modules package-lock.json
print_green "清理完成!"

# 修复package.json中的依赖问题
print_yellow "检查package.json中的Vue版本..."

# 备份原始package.json
cp package.json package.json.bak
print_yellow "已备份原始package.json到package.json.bak"

# 安装Vue 2.x和相关依赖
print_yellow "安装Vue 2.x和相关依赖..."
npm install vue@2.6.14 --save --legacy-peer-deps
npm install @vue/compiler-sfc@2.6.14 --save-dev --legacy-peer-deps

# 移除可能冲突的依赖
print_yellow "移除可能冲突的依赖..."
npm uninstall @vitejs/plugin-vue vite --legacy-peer-deps

# 安装其他项目依赖
print_yellow "安装项目依赖..."
npm install --legacy-peer-deps --force

# 检查安装结果
if [ $? -ne 0 ]; then
  print_red "依赖安装失败，请检查错误信息"
  exit 1
fi

print_green "依赖安装成功!"

# 创建启动脚本
print_yellow "创建启动脚本..."
cat > start.sh << 'EOF'
#!/bin/bash
echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       启动前端开发服务器\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 启动开发服务器，使用HOST=0.0.0.0让服务器可以从外部访问
HOST=0.0.0.0 npm run serve
EOF

chmod +x start.sh
print_green "启动脚本创建完成!"

print_blue "==================================================="
print_blue "       环境初始化完成!"
print_blue "==================================================="
print_yellow "您现在可以运行 ./start.sh 来启动前端开发服务器"
print_yellow "或者直接运行 HOST=0.0.0.0 npm run serve"