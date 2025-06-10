#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       Vue依赖修复脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

echo -e "\033[33m正在清理node_modules和package-lock.json...\033[0m"
rm -rf node_modules package-lock.json
echo -e "\033[32m清理完成!\033[0m"

echo -e "\033[33m安装Vue 2.x和相关依赖...\033[0m"
npm install vue@2.6.14 --save --legacy-peer-deps
npm install @vue/compiler-sfc@2.6.14 --save-dev --legacy-peer-deps

echo -e "\033[33m移除可能冲突的依赖...\033[0m"
npm uninstall @vitejs/plugin-vue vite --legacy-peer-deps

echo -e "\033[33m正在安装其他项目依赖...\033[0m"
npm install --legacy-peer-deps --force

echo -e "\033[32m依赖修复完成!\033[0m"
echo -e "\033[33m现在可以尝试运行 ./start-auto.sh 或 ./start-ci.sh 启动项目\033[0m"