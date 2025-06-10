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
print_blue "       前端环境自动初始化和启动脚本"
print_blue "==================================================="

# 检查Node.js版本
print_yellow "检查Node.js环境..."
if ! command -v node &> /dev/null; then
  print_red "未检测到Node.js，请先安装Node.js"
  exit 1
fi

NODE_VERSION=$(node -v)
print_green "Node.js版本: $NODE_VERSION"

# 检查Node.js版本是否满足要求（需要Node.js 14+）
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

# 检查npm版本
NPM_VERSION=$(npm -v)
print_green "npm版本: $NPM_VERSION"

# 强制清理依赖以解决版本冲突问题
print_yellow "是否需要强制清理依赖以解决版本冲突? (y/n，推荐选择y)"
read -r clean_deps

if [ "$clean_deps" = "y" ] || [ "$clean_deps" = "Y" ]; then
  print_yellow "正在清理node_modules和package-lock.json..."
  rm -rf node_modules package-lock.json
  print_green "清理完成!"
  
  # 检查package.json并修复Vue依赖
  print_yellow "正在检查Vue依赖..."
  
  # 确保安装Vue 2.x版本和相关依赖
  print_yellow "安装Vue 2.x和相关依赖..."
  npm install vue@2.6.14 --save --legacy-peer-deps
  npm install @vue/compiler-sfc@2.6.14 --save-dev --legacy-peer-deps
  
  # 移除可能冲突的依赖
  print_yellow "移除可能冲突的依赖..."
  npm uninstall @vitejs/plugin-vue vite --legacy-peer-deps
  
  print_yellow "正在安装项目依赖..."
  # 使用--force参数强制解决依赖冲突
  npm install --legacy-peer-deps
  
  if [ $? -ne 0 ]; then
    print_red "依赖安装失败，尝试使用其他参数..."
    npm install --legacy-peer-deps --no-optional
    
    if [ $? -ne 0 ]; then
      print_yellow "再次尝试使用--force参数..."
      npm install --legacy-peer-deps --force
      
      if [ $? -ne 0 ]; then
        print_red "依赖安装再次失败，请手动解决问题"
        exit 1
      fi
    fi
  fi
  print_green "依赖安装成功!"
else
  # 检查node_modules是否存在
  if [ ! -d "node_modules" ]; then
    print_yellow "未检测到node_modules，将安装依赖..."
    
    # 确保安装Vue 2.x版本和相关依赖
    print_yellow "安装Vue 2.x和相关依赖..."
    npm install vue@2.6.14 --save --legacy-peer-deps
    npm install @vue/compiler-sfc@2.6.14 --save-dev --legacy-peer-deps
    
    # 移除可能冲突的依赖
    print_yellow "移除可能冲突的依赖..."
    npm uninstall @vitejs/plugin-vue vite --legacy-peer-deps
    
    npm install --legacy-peer-deps
    
    if [ $? -ne 0 ]; then
      print_red "依赖安装失败，尝试清理并重新安装..."
      rm -rf node_modules package-lock.json
      npm install --legacy-peer-deps --force
      
      if [ $? -ne 0 ]; then
        print_red "依赖安装再次失败，请手动解决问题"
        exit 1
      fi
    fi
    
    print_green "依赖安装成功!"
  else
    print_green "检测到node_modules已存在，跳过安装步骤"
  fi
fi

# 启动开发服务器
print_yellow "正在启动开发服务器..."
print_blue "==================================================="
print_blue "  开发服务器启动中，请稍候..."
print_blue "  按Ctrl+C可以停止服务器"
print_blue "==================================================="

# 启动开发服务器，使用HOST=0.0.0.0让服务器可以从外部访问
HOST=0.0.0.0 npm run serve 