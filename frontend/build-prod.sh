#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       生产环境构建脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 创建.env.production文件
echo -e "\033[33m创建.env.production配置文件...\033[0m"
cat > .env.production << 'EOF'
VUE_APP_API_URL=/api
NODE_ENV=production
EOF
echo -e "\033[32m配置文件创建完成!\033[0m"

# 创建vue.config.js配置文件
echo -e "\033[33m创建vue.config.js配置文件...\033[0m"
cat > vue.config.js << 'EOF'
module.exports = {
  publicPath: '/',
  productionSourceMap: false
}
EOF
echo -e "\033[32m配置文件创建完成!\033[0m"

# 执行构建
echo -e "\033[33m开始构建生产环境版本...\033[0m"
npm run build

if [ $? -ne 0 ]; then
  echo -e "\033[31m构建失败，请检查错误信息\033[0m"
  exit 1
fi

echo -e "\033[32m构建完成!\033[0m"

# 部署说明
echo -e "\033[33m部署说明:\033[0m"
echo -e "\033[33m1. 将dist目录复制到服务器\033[0m"
echo -e "\033[33m2. 配置Nginx提供静态文件\033[0m" 