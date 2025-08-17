<template>
    <div class="register-container">
        <div class="register-box">
            <div class="register-title">
                <h2>注册账号</h2>
                <p>Virtual Human Studio</p>
            </div>

            <el-form ref="registerForm" :model="registerForm" :rules="registerRules" class="register-form">
                <el-form-item prop="username">
                    <el-input v-model="registerForm.username" prefix-icon="el-icon-user" placeholder="用户名"></el-input>
                </el-form-item>

                <el-form-item prop="password">
                    <el-input v-model="registerForm.password" prefix-icon="el-icon-lock" type="password"
                        placeholder="密码"></el-input>
                </el-form-item>

                <el-form-item prop="confirmPassword">
                    <el-input v-model="registerForm.confirmPassword" prefix-icon="el-icon-lock" type="password"
                        placeholder="确认密码"></el-input>
                </el-form-item>

                <el-form-item prop="email">
                    <el-input v-model="registerForm.email" prefix-icon="el-icon-message"
                        placeholder="邮箱（选填）"></el-input>
                </el-form-item>

                <el-form-item prop="phone">
                    <el-input v-model="registerForm.phone" prefix-icon="el-icon-mobile-phone"
                        placeholder="手机号（选填）"></el-input>
                </el-form-item>

                <el-form-item>
                    <el-button type="primary" class="register-button" :loading="loading"
                        @click="handleRegister">注册</el-button>
                </el-form-item>

                <div class="register-options">
                    <span>已有账号？</span>
                    <router-link to="/login">立即登录</router-link>
                </div>
            </el-form>
        </div>
    </div>
</template>

<script>
export default {
    name: 'RegisterView',
    data() {
        // 确认密码验证
        const validateConfirmPassword = (rule, value, callback) => {
            if (value !== this.registerForm.password) {
                callback(new Error('两次输入的密码不一致'))
            } else {
                callback()
            }
        }

        // 邮箱验证
        const validateEmail = (rule, value, callback) => {
            if (value === '') {
                callback()
            } else {
                const emailRegex = /^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/
                if (emailRegex.test(value)) {
                    callback()
                } else {
                    callback(new Error('请输入有效的邮箱地址'))
                }
            }
        }

        // 手机号验证
        const validatePhone = (rule, value, callback) => {
            if (value === '') {
                callback()
            } else {
                const phoneRegex = /^1[3-9]\d{9}$/
                if (phoneRegex.test(value)) {
                    callback()
                } else {
                    callback(new Error('请输入有效的手机号'))
                }
            }
        }

        return {
            registerForm: {
                username: '',
                password: '',
                confirmPassword: '',
                email: '',
                phone: ''
            },
            registerRules: {
                username: [
                    { required: true, message: '请输入用户名', trigger: 'blur' },
                    { min: 3, max: 20, message: '用户名长度在3到20个字符之间', trigger: 'blur' }
                ],
                password: [
                    { required: true, message: '请输入密码', trigger: 'blur' },
                    { min: 6, message: '密码长度至少为6个字符', trigger: 'blur' }
                ],
                confirmPassword: [
                    { required: true, message: '请确认密码', trigger: 'blur' },
                    { validator: validateConfirmPassword, trigger: 'blur' }
                ],
                email: [
                    { validator: validateEmail, trigger: 'blur' }
                ],
                phone: [
                    { validator: validatePhone, trigger: 'blur' }
                ]
            },
            loading: false
        }
    },
    methods: {
        handleRegister() {
            this.$refs.registerForm.validate(valid => {
                if (valid) {
                    this.loading = true
                    
                    // 构建请求数据
                    const data = {
                        username: this.registerForm.username,
                        password: this.registerForm.password,
                        email: this.registerForm.email,
                        phone: this.registerForm.phone
                    }
                    
                    // 发送注册请求
                    this.$http.post('/api/register', data)
                        .then(() => {
                            this.loading = false
                            
                            // 显示成功消息
                            this.$message.success('注册成功')
                            
                            // 跳转到登录页
                            this.$router.push('/login')
                        })
                        .catch(error => {
                            this.loading = false
                            
                            // 显示错误消息
                            let errorMessage = '注册失败，请稍后重试'
                            if (error.response && error.response.data) {
                                if (typeof error.response.data === 'string') {
                                    errorMessage = error.response.data
                                } else if (error.response.data.error) {
                                    errorMessage = error.response.data.error
                                }
                            }
                            this.$message.error(errorMessage)
                        })
                }
            })
        }
    }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  position: relative;
  overflow: hidden;
}

.register-box {
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

.register-box:hover {
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  transform: translateY(-5px);
}

.register-title {
  text-align: center;
  margin-bottom: 30px;
}

.register-title h2 {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(120deg, #2c3e50, #4a6572);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 12px;
}

.register-title p {
  font-size: 16px;
  color: rgba(44, 62, 80, 0.7);
  letter-spacing: 0.5px;
}

.register-form {
  margin-bottom: 20px;
}

.register-button {
  width: 100%;
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border: none !important;
  border-radius: 8px;
  padding: 12px 0;
  font-weight: 600;
  letter-spacing: 0.5px;
  transition: all 0.3s ease;
}

.register-button:hover,
.register-button:focus,
.register-button:active {
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border-color: transparent !important;
  color: #ffffff !important;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(44, 62, 80, 0.2);
}

/* 覆盖所有可能的Element UI样式 */
.el-button.el-button--primary.register-button:hover,
.el-button.el-button--primary.register-button:focus,
.el-button.el-button--primary.register-button:active {
  background: linear-gradient(90deg, #2c3e50, #4a6572) !important;
  border-color: transparent !important;
}

.register-options {
  text-align: center;
  font-size: 14px;
  color: rgba(44, 62, 80, 0.7);
  margin-top: 15px;
}

.register-options a {
  color: #2c3e50;
  margin-left: 5px;
  font-weight: 600;
  text-decoration: none;
  position: relative;
  padding-bottom: 2px;
}

.register-options a::after {
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

.register-options a:hover::after {
  transform: scaleX(1);
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

.register-box {
  animation: fadeIn 0.6s ease forwards;
}

@media (max-width: 767px) {
  .register-box {
    width: 85%;
    padding: 30px 20px;
  }
}
</style>