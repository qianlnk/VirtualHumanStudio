#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       Element UI依赖修复脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

echo -e "\033[33m安装Element UI及相关依赖...\033[0m"
npm install element-ui@2.15.13 --save --legacy-peer-deps

echo -e "\033[33m安装可能缺少的其他依赖...\033[0m"
npm install axios@0.27.2 --save --legacy-peer-deps
npm install vue-router@3.6.5 --save --legacy-peer-deps
npm install vuex@3.6.2 --save --legacy-peer-deps
npm install js-cookie@3.0.1 --save --legacy-peer-deps
npm install core-js@3.30.2 --save --legacy-peer-deps

echo -e "\033[32m依赖安装完成!\033[0m"
echo -e "\033[33m现在可以尝试运行 HOST=0.0.0.0 npm run serve 启动项目\033[0m" 