#!/bin/bash

echo -e "\033[34m===================================================\033[0m"
echo -e "\033[34m       更新Nginx配置脚本\033[0m"
echo -e "\033[34m===================================================\033[0m"

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
  echo -e "\033[31m请使用sudo运行此脚本\033[0m"
  exit 1
fi

# 备份现有配置
echo -e "\033[33m备份现有Nginx配置...\033[0m"
TIMESTAMP=$(date +"%Y%m%d%H%M%S")
NGINX_CONF_DIR="/etc/nginx/conf.d"
mkdir -p /tmp/nginx_backup_$TIMESTAMP

if [ -f "$NGINX_CONF_DIR/frontend.conf" ]; then
  cp "$NGINX_CONF_DIR/frontend.conf" "/tmp/nginx_backup_$TIMESTAMP/"
fi

if [ -f "$NGINX_CONF_DIR/backend.conf" ]; then
  cp "$NGINX_CONF_DIR/backend.conf" "/tmp/nginx_backup_$TIMESTAMP/"
fi

echo -e "\033[32m备份完成: /tmp/nginx_backup_$TIMESTAMP/\033[0m"

# 复制新配置
echo -e "\033[33m更新Nginx配置...\033[0m"
cp frontend.conf "$NGINX_CONF_DIR/"
cp backend.conf "$NGINX_CONF_DIR/"

# 验证Nginx配置
echo -e "\033[33m验证Nginx配置...\033[0m"
nginx -t

if [ $? -ne 0 ]; then
  echo -e "\033[31mNginx配置验证失败，恢复备份...\033[0m"
  if [ -f "/tmp/nginx_backup_$TIMESTAMP/frontend.conf" ]; then
    cp "/tmp/nginx_backup_$TIMESTAMP/frontend.conf" "$NGINX_CONF_DIR/"
  fi
  if [ -f "/tmp/nginx_backup_$TIMESTAMP/backend.conf" ]; then
    cp "/tmp/nginx_backup_$TIMESTAMP/backend.conf" "$NGINX_CONF_DIR/"
  fi
  exit 1
fi

# 重新加载Nginx
echo -e "\033[33m重新加载Nginx...\033[0m"
systemctl reload nginx

if [ $? -ne 0 ]; then
  echo -e "\033[31mNginx重新加载失败，请检查错误信息\033[0m"
  exit 1
fi

echo -e "\033[32m===================================================\033[0m"
echo -e "\033[32mNginx配置已更新并重新加载!\033[0m"
echo -e "\033[32m===================================================\033[0m" 