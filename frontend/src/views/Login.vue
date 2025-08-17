<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-title">
        <h2>登录</h2>
        <p>Virtual Human Studio</p>
      </div>
      
      <el-form ref="loginForm" :model="loginForm" :rules="loginRules" class="login-form">
        <el-form-item prop="username">
          <el-input 
            v-model="loginForm.username" 
            prefix-icon="el-icon-user" 
            placeholder="用户名"
          ></el-input>
        </el-form-item>
        
        <el-form-item prop="password">
          <el-input 
            v-model="loginForm.password" 
            prefix-icon="el-icon-lock" 
            type="password" 
            placeholder="密码"
            @keyup.enter.native="handleLogin"
          ></el-input>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            class="login-button" 
            :loading="loading" 
            @click="handleLogin"
          >登录</el-button>
        </el-form-item>
        
        <div class="login-options">
          <span>没有账号？</span>
          <router-link to="/register">立即注册</router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LoginView',
  data() {
    return {
      loginForm: {
        username: '',
        password: ''
      },
      loginRules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' }
        ]
      },
      loading: false
    }
  },
  methods: {
    handleLogin() {
      this.$refs.loginForm.validate(valid => {
        if (valid) {
          this.loading = true
          
          // 使用store的login action
          this.$store.dispatch('login', {
            username: this.loginForm.username,
            password: this.loginForm.password
          })
            .then(() => {
              this.loading = false
              
              // 显示成功消息
              this.$message.success('登录成功')
              
              // 跳转到首页
              this.$router.push('/')
            })
            .catch(error => {
              this.loading = false
              
              // 显示错误消息
              if (error.response && error.response.data && error.response.data.error) {
                this.$message.error(error.response.data.error)
              } else {
                this.$message.error('登录失败，请稍后重试')
              }
            })
        }
      })
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  position: relative;
  overflow: hidden;
}

.login-box {
  width: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-radius: 16px;
  border: 1px solid rgba(200, 200, 200, 0.4);
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
  transition: all 0.4s cubic-bezier(0.165, 0.84, 0.44, 1);
  position: relative;
  z-index: 1;
}

.login-box:hover {
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  transform: translateY(-5px);
}

.login-title {
  text-align: center;
  margin-bottom: 30px;
}

.login-title h2 {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(120deg, #2c3e50, #4a6572);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 12px;
}

.login-title p {
  font-size: 16px;
  color: rgba(44, 62, 80, 0.7);
  letter-spacing: 0.5px;
}

.login-form {
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
}

.login-options {
  text-align: center;
  font-size: 14px;
  color: rgba(44, 62, 80, 0.7);
  margin-top: 15px;
}

.login-options a {
  color: #2c3e50;
  margin-left: 5px;
  font-weight: 600;
  text-decoration: none;
  position: relative;
  padding-bottom: 2px;
}

.login-options a::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, #2c3e50, transparent);
  transform: scaleX(0.5);
  transform-origin: left;
  transition: transform 0.3s ease;
}

.login-options a:hover::after {
  transform: scaleX(1);
}

.login-button {
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border: none !important;
  border-radius: 8px;
  padding: 12px 0;
  font-weight: 600;
  letter-spacing: 0.5px;
  transition: all 0.3s ease;
}

.login-button:hover,
.login-button:focus,
.login-button:active {
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border-color: transparent !important;
  color: #ffffff !important;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}

/* 覆盖所有可能的Element UI样式 */
.el-button.el-button--primary.login-button:hover,
.el-button.el-button--primary.login-button:focus,
.el-button.el-button--primary.login-button:active {
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border-color: transparent !important;
}

.el-input__inner {
  border-radius: 8px;
  border: 1px solid rgba(44, 62, 80, 0.15);
  padding: 12px;
  height: 45px;
}

.el-input__inner:focus {
  border-color: rgba(44, 62, 80, 0.5);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-box {
  animation: fadeIn 0.6s ease forwards;
}

@media (max-width: 767px) {
  .login-box {
    width: 85%;
    padding: 30px 20px;
  }
}
</style>