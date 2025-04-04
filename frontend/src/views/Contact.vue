<template>
  <div class="contact-container">
    <div class="page-header">
      <h2>联系我们</h2>
    </div>
    
    <div class="contact-content">
      <el-row :gutter="30">
        <!-- 左侧合作说明 -->
        <el-col :span="14">
          <el-card class="glass-card">
            <div slot="header" class="card-header">
              <span>大量合成业务合作</span>
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
          </el-card>
          
          <!-- 留言表单 -->
          <el-card class="glass-card message-form-card">
            <div slot="header" class="card-header">
              <span>留言咨询</span>
            </div>
            <el-form :model="messageForm" :rules="rules" ref="messageForm" label-width="80px">
              <el-form-item label="姓名" prop="name">
                <el-input v-model="messageForm.name" placeholder="请输入您的姓名"></el-input>
              </el-form-item>
              <el-form-item label="电话" prop="phone">
                <el-input v-model="messageForm.phone" placeholder="请输入您的联系电话"></el-input>
              </el-form-item>
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="messageForm.email" placeholder="请输入您的邮箱"></el-input>
              </el-form-item>
              <el-form-item label="留言内容" prop="message">
                <el-input type="textarea" v-model="messageForm.message" :rows="4" placeholder="请输入您的留言内容"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="submitMessage">提交留言</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>
        
        <!-- 右侧微信二维码 -->
        <el-col :span="10">
          <el-card class="glass-card qrcode-card">
            <div slot="header" class="card-header">
              <span>扫码联系我们</span>
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
          </el-card>
          
          <el-card class="glass-card service-card">
            <div slot="header" class="card-header">
              <span>我们的服务</span>
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
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Contact',
  data() {
    return {
      messageForm: {
        name: '',
        phone: '',
        email: '',
        message: ''
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
        message: [
          { required: true, message: '请输入留言内容', trigger: 'blur' },
          { min: 5, max: 500, message: '长度在 5 到 500 个字符', trigger: 'blur' }
        ]
      }
    }
  },
  methods: {
    submitMessage() {
      this.$refs.messageForm.validate((valid) => {
        if (valid) {
          // 这里可以添加提交留言的API调用
          // 由于目前没有实际的后端API，我们只显示一个成功消息
          this.$message.success('留言提交成功，我们将尽快与您联系！')
          // 重置表单
          this.$refs.messageForm.resetFields()
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
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.contact-content {
  margin-top: 20px;
}

.glass-card {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.25);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  border-radius: 10px;
  overflow: hidden;
  transition: all 0.3s ease;
  margin-bottom: 20px;
}

.glass-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 15px 30px rgba(0, 0, 0, 0.2);
}

.card-header {
  font-size: 18px;
  font-weight: 700;
  color: #e3f2fd;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  letter-spacing: 0.5px;
}

.cooperation-info {
  padding: 10px 0;
}

.cooperation-info p {
  line-height: 1.6;
  margin-bottom: 15px;
  color: #ffffff;
  font-weight: 500;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

.cooperation-info ul {
  padding-left: 20px;
  margin-bottom: 20px;
}

.cooperation-info li {
  margin-bottom: 10px;
  line-height: 1.6;
  color: #ffffff;
  font-weight: 500;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

.contact-methods {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.contact-methods h3 {
  margin-bottom: 15px;
  color: #e3f2fd;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  letter-spacing: 0.5px;
}

.contact-methods p {
  margin-bottom: 10px;
  color: #ffffff;
  font-weight: 500;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

.contact-methods i {
  margin-right: 10px;
  color: #90caf9;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.message-form-card {
  margin-top: 20px;
}

.qrcode-card {
  height: 100%;
}

.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
}

.qrcode-image {
  margin-bottom: 20px;
  background-color: #fff;
  padding: 10px;
  border-radius: 5px;
}

.qrcode-tip {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 15px;
  color: #e3f2fd;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  letter-spacing: 0.5px;
}

.qrcode-info {
  text-align: center;
  margin-top: 15px;
}

.qrcode-info p {
  margin-bottom: 5px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  font-weight: 500;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

.service-card {
  margin-top: 20px;
}

.service-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  padding: 10px 0;
}

.service-item {
  display: flex;
  align-items: center;
  padding: 10px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 5px;
  transition: all 0.3s ease;
}

.service-item:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
}

.service-item i {
  font-size: 24px;
  margin-right: 10px;
  color: #90caf9;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.service-item span {
  color: #ffffff;
  font-weight: 500;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.3);
}

/* 响应式调整 */
@media (max-width: 1200px) {
  .el-col-14 {
    width: 100%;
  }
  
  .el-col-10 {
    width: 100%;
  }
}
</style>