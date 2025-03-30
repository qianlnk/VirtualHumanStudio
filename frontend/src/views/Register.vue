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
    name: 'Register',
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