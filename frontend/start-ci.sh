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
print_blue "       前端环境CI/CD自动初始化和启动脚本"
print_blue "==================================================="

# 检查Node.js版本
print_yellow "检查Node.js环境..."
if ! command -v node &> /dev/null; then
  print_red "未检测到Node.js，请先安装Node.js"
  exit 1
fi

NODE_VERSION=$(node -v)
print_green "Node.js版本: $NODE_VERSION"

# 检查npm版本
NPM_VERSION=$(npm -v)
print_green "npm版本: $NPM_VERSION"

# 强制清理依赖
print_yellow "正在清理node_modules和package-lock.json..."
rm -rf node_modules package-lock.json
print_green "清理完成!"

# 安装依赖
print_yellow "正在安装项目依赖..."
npm install --legacy-peer-deps --force

if [ $? -ne 0 ]; then
  print_yellow "尝试使用其他参数安装..."
  npm install --legacy-peer-deps --no-optional
  
  if [ $? -ne 0 ]; then
    print_red "依赖安装失败，退出"
    exit 1
  fi
fi

print_green "依赖安装成功!"

# 启动开发服务器
print_yellow "正在启动开发服务器..."
print_blue "==================================================="
print_blue "  开发服务器启动中，请稍候..."
print_blue "  按Ctrl+C可以停止服务器"
print_blue "==================================================="

# 启动开发服务器，使用HOST=0.0.0.0让服务器可以从外部访问
HOST=0.0.0.0 npm run serve