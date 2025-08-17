<template>
  <div class="contact-container">
    <div class="section-title">
      <div class="section-accent"></div>
      <h2>联系我们</h2>
    </div>

    <el-row :gutter="30" class="contact-content">
      <el-col :xs="24" :md="14">
        <div class="glass-card">
          <div class="section-header">
            <div class="section-accent"></div>
            <h3>大量合成业务合作</h3>
          </div>
          <div class="cooperation-info">
            <p>感谢您对Virtual Human Studio的关注！如果您有以下需求，欢迎与我们联系：</p>
            <ul>
              <li>大批量数字人合成需求</li>
              <li>定制化数字人解决方案</li>
              <li>企业级语音克隆服务</li>
              <li>商业合作与授权</li>
              <li>技术支持与咨询</li>
            </ul>

            <div class="contact-methods">
              <h3>联系方式</h3>
              <p><i class="el-icon-message"></i> qianlnk@163.com</p>
              <p><i class="el-icon-phone"></i> 电话：18926447244</p>
              <p><i class="el-icon-location"></i> 地址：广州市天河区</p>
            </div>
          </div>
        </div>

        <div class="glass-card message-form-card">
          <div class="section-header">
            <div class="section-accent"></div>
            <h3>留言咨询</h3>
          </div>
          <el-form :model="messageForm" :rules="rules" ref="messageForm" label-width="88px" size="small" class="message-form">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="messageForm.name" placeholder="请输入您的姓名"></el-input>
            </el-form-item>
            <el-form-item label="电话" prop="phone">
              <el-input v-model="messageForm.phone" placeholder="请输入您的联系电话"></el-input>
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="messageForm.email" placeholder="请输入您的邮箱"></el-input>
            </el-form-item>
            <el-form-item label="主题" prop="subject">
              <el-input v-model="messageForm.subject" placeholder="请输入留言主题"></el-input>
            </el-form-item>
            <el-form-item label="留言内容" prop="content">
              <el-input type="textarea" v-model="messageForm.content" :rows="4" placeholder="请输入您的留言内容"></el-input>
            </el-form-item>
            <div class="form-actions">
              <el-button type="primary" class="action-button" @click="submitMessage" :loading="submitting">提交留言</el-button>
            </div>
          </el-form>
        </div>
      </el-col>

      <el-col :xs="24" :md="10">
        <div class="glass-card qrcode-card">
          <div class="section-header">
            <div class="section-accent"></div>
            <h3>扫码联系我们</h3>
          </div>
          <div class="qrcode-container">
            <div class="qrcode-image">
              <!-- 这里使用SVG作为示例二维码，实际使用时可替换为真实的二维码图片 -->
              <svg width="200" height="200" viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
                <rect x="0" y="0" width="200" height="200" fill="#ffffff"/>
                <g fill="#000000">
                  <!-- 外框 -->
                  <rect x="20" y="20" width="160" height="160" fill="none" stroke="#000000" stroke-width="2"/>
                  <!-- 左上角定位图案 -->
                  <rect x="40" y="40" width="30" height="30"/>
                  <rect x="45" y="45" width="20" height="20" fill="#ffffff"/>
                  <rect x="50" y="50" width="10" height="10"/>
                  <!-- 右上角定位图案 -->
                  <rect x="130" y="40" width="30" height="30"/>
                  <rect x="135" y="45" width="20" height="20" fill="#ffffff"/>
                  <rect x="140" y="50" width="10" height="10"/>
                  <!-- 左下角定位图案 -->
                  <rect x="40" y="130" width="30" height="30"/>
                  <rect x="45" y="135" width="20" height="20" fill="#ffffff"/>
                  <rect x="50" y="140" width="10" height="10"/>
                  <!-- 二维码数据点 (简化示例) -->
                  <rect x="85" y="40" width="5" height="5"/>
                  <rect x="95" y="40" width="5" height="5"/>
                  <rect x="105" y="40" width="5" height="5"/>
                  <rect x="85" y="50" width="5" height="5"/>
                  <rect x="95" y="50" width="5" height="5"/>
                  <rect x="105" y="50" width="5" height="5"/>
                  <rect x="40" y="85" width="5" height="5"/>
                  <rect x="50" y="85" width="5" height="5"/>
                  <rect x="60" y="85" width="5" height="5"/>
                  <rect x="70" y="85" width="5" height="5"/>
                  <rect x="80" y="85" width="5" height="5"/>
                  <rect x="90" y="85" width="5" height="5"/>
                  <rect x="100" y="85" width="5" height="5"/>
                  <rect x="110" y="85" width="5" height="5"/>
                  <rect x="120" y="85" width="5" height="5"/>
                  <rect x="130" y="85" width="5" height="5"/>
                  <rect x="140" y="85" width="5" height="5"/>
                  <rect x="150" y="85" width="5" height="5"/>
                  <rect x="85" y="95" width="5" height="5"/>
                  <rect x="95" y="95" width="5" height="5"/>
                  <rect x="105" y="95" width="5" height="5"/>
                  <rect x="85" y="105" width="5" height="5"/>
                  <rect x="95" y="105" width="5" height="5"/>
                  <rect x="105" y="105" width="5" height="5"/>
                  <rect x="85" y="115" width="5" height="5"/>
                  <rect x="95" y="115" width="5" height="5"/>
                  <rect x="105" y="115" width="5" height="5"/>
                  <rect x="85" y="125" width="5" height="5"/>
                  <rect x="95" y="125" width="5" height="5"/>
                  <rect x="105" y="125" width="5" height="5"/>
                  <rect x="85" y="135" width="5" height="5"/>
                  <rect x="95" y="135" width="5" height="5"/>
                  <rect x="105" y="135" width="5" height="5"/>
                  <rect x="85" y="145" width="5" height="5"/>
                  <rect x="95" y="145" width="5" height="5"/>
                  <rect x="105" y="145" width="5" height="5"/>
                  <rect x="85" y="155" width="5" height="5"/>
                  <rect x="95" y="155" width="5" height="5"/>
                  <rect x="105" y="155" width="5" height="5"/>
                </g>
              </svg>
            </div>
            <p class="qrcode-tip">扫描上方二维码添加客服微信</p>
            <div class="qrcode-info">
              <p>工作时间：周一至周五 9:00-18:00</p>
              <p>我们将尽快回复您的咨询</p>
            </div>
          </div>
        </div>

        <div class="glass-card service-card">
          <div class="section-header">
            <div class="section-accent"></div>
            <h3>我们的服务</h3>
          </div>
          <div class="service-list">
            <div class="service-item">
              <i class="el-icon-video-camera"></i>
              <span>数字人定制</span>
            </div>
            <div class="service-item">
              <i class="el-icon-microphone"></i>
              <span>语音克隆</span>
            </div>
            <div class="service-item">
              <i class="el-icon-headset"></i>
              <span>TTS服务</span>
            </div>
            <div class="service-item">
              <i class="el-icon-data-analysis"></i>
              <span>技术咨询</span>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { createMessage } from '@/api/message'

export default {
  name: 'ContactView',
  data() {
    return {
      submitting: false,
      messageForm: {
        name: '',
        phone: '',
        email: '',
        subject: '',
        content: ''
      },
      rules: {
        name: [
          { required: true, message: '请输入您的姓名', trigger: 'blur' },
          { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
        ],
        phone: [
          { required: true, message: '请输入您的联系电话', trigger: 'blur' },
          { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码', trigger: 'blur' }
        ],
        email: [
          { required: true, message: '请输入您的邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
        ],
        subject: [
          { required: true, message: '请输入留言主题', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        content: [
          { required: true, message: '请输入留言内容', trigger: 'blur' },
          { min: 5, max: 500, message: '长度在 5 到 500 个字符', trigger: 'blur' }
        ]
      }
    }
  },
  methods: {
    submitMessage() {
      this.$refs.messageForm.validate(async (valid) => {
        if (valid) {
          try {
            this.submitting = true
            const result = await createMessage(this.messageForm)

            if (result.success) {
              this.$message.success('留言提交成功，我们将尽快与您联系！')
              // 重置表单
              this.$refs.messageForm.resetFields()
            } else {
              this.$message.error(result.error || '留言提交失败，请稍后再试')
            }
          } catch (error) {
            console.error('提交留言出错:', error)
            this.$message.error('系统错误，请稍后再试')
          } finally {
            this.submitting = false
          }
        } else {
          this.$message.error('请正确填写表单信息')
          return false
        }
      })
    }
  }
}
</script>

<style scoped>
.contact-container {
  min-height: 100vh;
  padding: 20px;
  color: rgba(44, 62, 80, 0.9);
  position: relative;
  overflow-x: hidden;
  transition: all 0.3s ease;
}

@media (min-width: 768px) {
  .contact-container {
    padding: 40px;
  }
}

@media (max-width: 767px) {
  .contact-container {
    padding: 15px;
  }
}

.section-title {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding-left: 10px;
}

.section-title h2 {
  font-size: 1.7em;
  font-weight: 700;
  margin: 0;
  color: #2c3e50;
  position: relative;
}

@media (min-width: 768px) {
  .section-title {
    margin-bottom: 30px;
  }
  .section-title h2 {
    font-size: 2em;
  }
}

.section-accent {
  width: 36px;
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  box-shadow: 0 0 10px rgba(0,0,0,0.08);
  margin-right: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.section-header h3 {
  font-size: 1.2em;
  margin: 0;
  color: #222222;
}

.glass-card {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-radius: 16px;
  border: 1px solid rgba(200, 200, 200, 0.4);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.06);
  transition: all 0.4s cubic-bezier(0.165, 0.84, 0.44, 1);
  padding: 20px;
  margin-bottom: 20px;
}

@media (min-width: 768px) {
  .glass-card {
    padding: 30px;
    margin-bottom: 30px;
  }
}

.glass-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12);
}

.cooperation-info p {
  line-height: 1.7;
  margin-bottom: 12px;
  color: #555555;
}

.cooperation-info ul {
  padding-left: 18px;
  margin: 10px 0 18px;
  color: #555555;
}

.cooperation-info li {
  margin-bottom: 8px;
  line-height: 1.6;
}

.contact-methods {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}

.contact-methods h3 {
  margin: 0 0 10px 0;
  color: #2c3e50;
  font-weight: 600;
}

.contact-methods p {
  margin-bottom: 8px;
  color: #333333;
}

.contact-methods i {
  margin-right: 8px;
  color: #2c3e50;
}

.message-form-card .el-form-item {
  margin-bottom: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-start;
}

.action-button {
  padding: 10px 20px;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.3s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.action-button.el-button--primary {
  background: linear-gradient(90deg, #2c3e50, #4a6572);
  border: none;
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}

.action-button.el-button--primary:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(44, 62, 80, 0.3);
}

.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10px 0;
}

.qrcode-image {
  margin-bottom: 16px;
  background-color: #fff;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid rgba(0,0,0,0.08);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
}

.qrcode-tip {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 10px;
  color: #2c3e50;
}

.qrcode-info {
  text-align: center;
  margin-top: 8px;
}

.qrcode-info p {
  margin-bottom: 4px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.6);
}

.service-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 6px 0;
}

.service-item {
  display: flex;
  align-items: center;
  padding: 10px;
  background: rgba(0, 0, 0, 0.03);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.service-item:hover {
  background: rgba(0, 0, 0, 0.06);
  transform: translateY(-2px);
}

.service-item i {
  font-size: 20px;
  margin-right: 10px;
  color: #2c3e50;
}

.service-item span {
  color: #333333;
  font-weight: 500;
}

@media (max-width: 1200px) {
  .el-col-14,
  .el-col-md-14 {
    width: 100% !important;
  }
  .el-col-10,
  .el-col-md-10 {
    width: 100% !important;
  }
}
</style>