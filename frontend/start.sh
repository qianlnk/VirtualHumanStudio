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
print_blue "       前端环境初始化和启动脚本"
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
if ! command -v npm &> /dev/null; then
  print_red "未检测到npm，请检查Node.js安装"
  exit 1
fi

NPM_VERSION=$(npm -v)
print_green "npm版本: $NPM_VERSION"

# 清理旧的依赖（可选，如果遇到依赖问题则使用）
print_yellow "是否需要清理旧的依赖? (y/n)"
read -r clean_deps

if [ "$clean_deps" = "y" ] || [ "$clean_deps" = "Y" ]; then
  print_yellow "正在清理node_modules和package-lock.json..."
  rm -rf node_modules package-lock.json
  print_green "清理完成!"
fi

# 安装依赖
print_yellow "正在安装项目依赖..."
npm install --legacy-peer-deps

# 检查安装结果
if [ $? -ne 0 ]; then
  print_red "依赖安装失败，请检查错误信息"
  exit 1
fi

print_green "依赖安装成功!"

# 启动开发服务器
print_yellow "正在启动开发服务器..."
print_blue "==================================================="
print_blue "  开发服务器启动中，请稍候..."
print_blue "  按Ctrl+C可以停止服务器"
print_blue "==================================================="

# 启动开发服务器
npm run serve 