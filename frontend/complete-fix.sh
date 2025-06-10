#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       完整前端环境修复脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

echo -e "\033[33m正在清理node_modules和package-lock.json...\033[0m"
rm -rf node_modules package-lock.json
echo -e "\033[32m清理完成!\033[0m"

echo -e "\033[33m备份原始package.json...\033[0m"
cp package.json package.json.bak
echo -e "\033[32m备份完成!\033[0m"

# 创建vue.config.js配置代理
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

# 修改main.js中的axios baseURL
echo -e "\033[33m修改main.js中的axios配置...\033[0m"
# 检查main.js是否存在
if [ -f src/main.js ]; then
  # 使用sed替换axios.defaults.baseURL
  sed -i 's|axios.defaults.baseURL = process.env.VUE_APP_API_URL || '\''http://localhost:8080'\''|axios.defaults.baseURL = '\''/api'\''|g' src/main.js
  echo -e "\033[32mmain.js修改完成!\033[0m"
else
  echo -e "\033[31msrc/main.js文件不存在，无法修改axios配置\033[0m"
fi

echo -e "\033[33m安装Vue 2.x和相关依赖...\033[0m"
npm install vue@2.6.14 --save --legacy-peer-deps
npm install vue-template-compiler@2.6.14 --save-dev --legacy-peer-deps
npm install @vue/compiler-sfc@2.6.14 --save-dev --legacy-peer-deps

echo -e "\033[33m安装Element UI及相关依赖...\033[0m"
npm install element-ui@2.15.13 --save --legacy-peer-deps

echo -e "\033[33m安装Vue路由和状态管理...\033[0m"
npm install vue-router@3.6.5 --save --legacy-peer-deps
npm install vuex@3.6.2 --save --legacy-peer-deps

echo -e "\033[33m安装其他常用依赖...\033[0m"
npm install axios@0.27.2 --save --legacy-peer-deps
npm install js-cookie@3.0.1 --save --legacy-peer-deps
npm install core-js@3.30.2 --save --legacy-peer-deps

echo -e "\033[33m移除可能冲突的依赖...\033[0m"
npm uninstall @vitejs/plugin-vue vite --legacy-peer-deps

echo -e "\033[33m安装Vue CLI相关依赖...\033[0m"
npm install @vue/cli-service@4.5.19 --save-dev --legacy-peer-deps
npm install @vue/cli-plugin-babel@4.5.19 --save-dev --legacy-peer-deps
npm install @vue/cli-plugin-eslint@4.5.19 --save-dev --legacy-peer-deps
npm install vue-loader@15.9.8 --save-dev --legacy-peer-deps

echo -e "\033[33m安装其他项目依赖...\033[0m"
npm install --legacy-peer-deps --force

echo -e "\033[32m=================================================\033[0m"
echo -e "\033[32m依赖修复完成!\033[0m"
echo -e "\033[32m=================================================\033[0m"

# 创建启动脚本
echo -e "\033[33m创建启动脚本...\033[0m"
cat > start.sh << 'EOF'
#!/bin/bash
echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       启动前端开发服务器\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 启动开发服务器，使用HOST=0.0.0.0让服务器可以从外部访问
HOST=0.0.0.0 npm run serve
EOF

chmod +x start.sh
echo -e "\033[32m启动脚本创建完成!\033[0m"

echo -e "\033[33m现在可以运行 ./start.sh 启动项目\033[0m"
echo -e "\033[33m前端将通过代理访问后端API: http://backend.aivhs.cn\033[0m"