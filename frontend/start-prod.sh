#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       生产环境前端启动脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 创建临时配置文件
echo -e "\033[33m创建vue.config.js配置文件...\033[0m"
cat > vue.config.js << 'EOF'
module.exports = {
  devServer: {
    disableHostCheck: true,
    host: '0.0.0.0',
    port: 8081,
    public: 'www.aivhs.cn',
    proxy: {
      '/api': {
        target: 'http://backend.aivhs.cn',
        changeOrigin: true,
        pathRewrite: {
          '^/api': ''
        }
      }
    }
  }
}
EOF

echo -e "\033[32m配置文件创建完成!\033[0m"

# 启动开发服务器
echo -e "\033[33m启动开发服务器...\033[0m"
echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m  开发服务器启动中，请稍候...\033[0m"
echo -e "\033[34m  按Ctrl+C可以停止服务器\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 使用NODE_ENV=production参数启动
NODE_ENV=production HOST=0.0.0.0 npm run serve 