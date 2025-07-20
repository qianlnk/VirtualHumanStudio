#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       SSL证书配置脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
    echo -e "\033[31m请使用root权限运行此脚本\033[0m"
    exit 1
fi

# 安装certbot
echo -e "\033[33m正在安装certbot...\033[0m"
if command -v apt-get &> /dev/null; then
    # Ubuntu/Debian
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
elif command -v yum &> /dev/null; then
    # CentOS/RHEL
    yum install -y epel-release
    yum install -y certbot python3-certbot-nginx
else
    echo -e "\033[31m不支持的包管理器\033[0m"
    exit 1
fi

# 获取SSL证书
echo -e "\033[33m正在获取SSL证书...\033[0m"

# 检查域名
DOMAIN="www.aivhs.cn"
echo -e "\033[33m配置域名: $DOMAIN\033[0m"

certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@aivhs.cn

# 检查证书是否获取成功
if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
    echo -e "\033[32mSSL证书获取成功!\033[0m"
 
 # 检查域名
DOMAIN="backend.aivhs.cn"
echo -e "\033[33m配置域名: $DOMAIN\033[0m"

certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@aivhs.cn

# 检查证书是否获取成功
if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
    echo -e "\033[32mSSL证书获取成功!\033[0m"