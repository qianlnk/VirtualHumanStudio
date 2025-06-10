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
print_blue "       前端环境快速初始化脚本"
print_blue "==================================================="

# 验证Node.js和npm安装
NODE_VERSION=$(node -v)
NPM_VERSION=$(npm -v)
print_green "Node.js版本: $NODE_VERSION"
print_green "npm版本: $NPM_VERSION"

# 检查Node.js版本
NODE_VERSION_NUM=$(echo $NODE_VERSION | cut -c 2- | awk -F. '{print $1}')
if [ "$NODE_VERSION_NUM" -lt 14 ]; then
  print_red "警告: 当前Node.js版本($NODE_VERSION)过低，推荐使用Node.js 14或更高版本"
  print_yellow "是否继续? (y/n)"
  read -r continue_with_old_node
  if [ "$continue_with_old_node" != "y" ] && [ "$continue_with_old_node" != "Y" ]; then
    print_yellow "已取消操作，请升级Node.js后再试"
    exit 0
  fi
  print_yellow "继续使用当前Node.js版本..."
fi

# 切换到前端目录
cd "$(dirname "$0")"
print_yellow "当前工作目录: $(pwd)"

# 清理旧的依赖
print_yellow "清理旧的依赖..."
rm -rf node_modules package-lock.json
print_green "清理完成!"

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

# 安装项目依赖
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