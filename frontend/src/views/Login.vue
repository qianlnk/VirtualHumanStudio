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
  name: 'Login',
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
          
          // 构建请求数据
          const data = {
            username: this.loginForm.username,
            password: this.loginForm.password
          }
          
          // 发送登录请求
          this.$http.post('/api/login', data)
            .then(response => {
              this.loading = false
              
              // 保存令牌和用户信息
              this.$store.commit('setToken', response.data.token)
              this.$store.commit('setUser', response.data.user)
              
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
  background-color: #f5f7fa;
}

.login-box {
  width: 400px;
  padding: 30px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.login-title {
  text-align: center;
  margin-bottom: 30px;
}

.login-title h2 {
  font-size: 24px;
  color: #303133;
  margin-bottom: 10px;
}

.login-title p {
  font-size: 14px;
  color: #909399;
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
  color: #606266;
}

.login-options a {
  color: #409EFF;
  margin-left: 5px;
}
</style>