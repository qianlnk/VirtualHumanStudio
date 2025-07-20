# SSL证书配置说明

## 问题描述

在云主机上部署前端应用后，录音功能无法使用，提示"您的浏览器不支持录音功能"。

## 问题原因

现代浏览器（特别是Chrome）出于安全考虑，要求在使用敏感API（如麦克风、摄像头、地理位置等）时，必须通过HTTPS协议访问网站。

- **本地开发环境**：`http://localhost:8081` - Chrome允许HTTP访问
- **生产环境**：`http://www.aivhs.cn` - Chrome要求HTTPS访问

## 解决方案

### 方案1：使用Let's Encrypt免费证书（推荐）

1. **在云主机上运行SSL配置脚本**：
   ```bash
   cd /root/frontend
   chmod +x ssl-setup.sh
   sudo ./ssl-setup.sh
   ```

2. **脚本会自动完成以下操作**：
   - 安装certbot工具
   - 获取Let's Encrypt免费SSL证书
   - 配置nginx支持HTTPS
   - 设置HTTP到HTTPS的自动跳转
   - 配置证书自动续期

3. **配置完成后**：
   - 网站将通过 `https://www.aivhs.cn` 访问
   - 录音功能将正常工作
   - 证书每90天自动续期

### 方案2：使用已有SSL证书

如果你已经有SSL证书：

1. **修改手动配置脚本**：
   ```bash
   cd /root/frontend
   chmod +x manual-ssl-setup.sh
   ```

2. **编辑脚本中的证书路径**：
   ```bash
   SSL_CERT_PATH="/path/to/your/certificate.crt"
   SSL_KEY_PATH="/path/to/your/private.key"
   ```

3. **运行配置脚本**：
   ```bash
   sudo ./manual-ssl-setup.sh
   ```

### 方案3：临时解决方案（仅用于测试）

如果暂时不想配置HTTPS，可以：

1. **使用Firefox浏览器**：Firefox在某些情况下对HTTPS要求较宽松
2. **在Chrome中启用不安全内容**：不推荐，仅用于测试
3. **使用本地开发环境测试录音功能**

## 验证配置

配置完成后，可以通过以下方式验证：

1. **访问HTTPS地址**：`https://www.aivhs.cn`
2. **检查浏览器地址栏**：应该显示锁图标
3. **测试录音功能**：应该可以正常使用麦克风

## 注意事项

1. **域名解析**：确保域名正确解析到云主机IP
2. **防火墙设置**：确保443端口开放
3. **证书续期**：Let's Encrypt证书有效期为90天，脚本会自动设置续期
4. **备份配置**：脚本会自动备份原始nginx配置

## 故障排除

### 证书获取失败
- 检查域名是否正确解析
- 确保80端口开放（Let's Encrypt验证需要）
- 检查防火墙设置

### nginx配置错误
- 检查nginx配置文件语法：`nginx -t`
- 查看nginx错误日志：`tail -f /var/log/nginx/error.log`

### 录音功能仍然不工作
- 确认使用HTTPS访问
- 检查浏览器控制台是否有错误信息
- 确认已授予麦克风权限

## 相关文件

- `ssl-setup.sh` - Let's Encrypt自动配置脚本
- `manual-ssl-setup.sh` - 手动SSL配置脚本
- `frontend-ssl` - HTTPS nginx配置文件
- `src/utils/audioUtils.js` - 音频工具函数

## 技术细节

### 浏览器安全策略
- Chrome 47+ 要求HTTPS访问敏感API
- Firefox 44+ 要求HTTPS访问敏感API
- Safari 11+ 要求HTTPS访问敏感API

### 支持的音频格式
- WAV (推荐，兼容性最好)
- MP3 (广泛支持)
- WebM (Chrome原生支持)
- OGG (Firefox原生支持)
- MP4 (iOS Safari支持) 