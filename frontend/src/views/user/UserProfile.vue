<template>
  <div class="user-profile-container">
    <div class="page-header">
      <h2>个人中心</h2>
    </div>
    
    <div class="profile-content">
      <el-row :gutter="20">
        <!-- 个人信息卡片 -->
        <el-col :span="8">
          <el-card class="profile-card" shadow="hover">
            <div class="user-avatar">
              <el-avatar :size="100" icon="el-icon-user-solid"></el-avatar>
            </div>
            <div class="user-info">
              <h3>{{ user.username }}</h3>
              <p class="user-role">{{ userRole }}</p>
              <p class="user-since">注册时间: {{ formatDate(user.created_at) }}</p>
              
              <!-- 添加会员中心入口 -->
              <div class="membership-entry">
                <router-link to="/membership" class="membership-link">
                  <el-button type="primary" size="small" icon="el-icon-medal">
                    会员中心
                  </el-button>
                </router-link>
              </div>
            </div>
            <div class="user-stats">
              <div class="stat-item">
                <span class="stat-value">{{ stats.voice_clones || 0 }}</span>
                <span class="stat-label">音色克隆</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ stats.tts_tasks || 0 }}</span>
                <span class="stat-label">TTS任务</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ stats.digital_humans || 0 }}</span>
                <span class="stat-label">数字人</span>
              </div>
            </div>
          </el-card>
        </el-col>
        
        <!-- 个人信息编辑 -->
        <el-col :span="16">
          <el-card class="info-edit-card">
            <div slot="header" class="clearfix">
              <span>个人信息</span>
            </div>
            
            <el-form :model="form" :rules="rules" ref="form" label-width="100px">
              <el-form-item label="用户名" prop="username">
                <el-input v-model="form.username" disabled></el-input>
              </el-form-item>
              
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="form.email" placeholder="请输入邮箱"></el-input>
              </el-form-item>
              
              <el-form-item label="手机号" prop="phone">
                <el-input v-model="form.phone" placeholder="请输入手机号"></el-input>
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" :loading="updating" @click="updateProfile">保存修改</el-button>
              </el-form-item>
            </el-form>
          </el-card>
          
          <!-- 修改密码 -->
          <el-card class="password-card">
            <div slot="header" class="clearfix">
              <span>修改密码</span>
            </div>
            
            <el-form :model="passwordForm" :rules="passwordRules" ref="passwordForm" label-width="100px">
              <el-form-item label="当前密码" prop="current_password">
                <el-input v-model="passwordForm.current_password" type="password" placeholder="请输入当前密码"></el-input>
              </el-form-item>
              
              <el-form-item label="新密码" prop="new_password">
                <el-input v-model="passwordForm.new_password" type="password" placeholder="请输入新密码"></el-input>
              </el-form-item>
              
              <el-form-item label="确认新密码" prop="confirm_password">
                <el-input v-model="passwordForm.confirm_password" type="password" placeholder="请再次输入新密码"></el-input>
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" :loading="changingPassword" @click="changePassword">修改密码</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'

export default {
  name: 'UserProfile',
  computed: {
    ...mapGetters(['user', 'userRole'])
  },
  created() {
    // 初始化表单数据
    this.form.username = this.user.username
    this.form.email = this.user.email
    this.form.phone = this.user.phone
  },
  data() {
    // 确认密码验证
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.passwordForm.new_password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }
    
    // 邮箱验证
    const validateEmail = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请输入邮箱地址'));
      } else {
        const emailRegex = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/;
        if (!emailRegex.test(value)) {
          callback(new Error('请输入有效的邮箱地址'));
        } else {
          callback();
        }
      }
    };

    return {
      form: {
        username: '',
        email: '',
        phone: ''
      },
      rules: {
        email: [{ validator: validateEmail, trigger: 'blur' }]
      },
      passwordForm: {
        current_password: '',
        new_password: '',
        confirm_password: ''
      },
      passwordRules: {
        confirm_password: [{ validator: validateConfirmPassword, trigger: 'blur' }]
      },
      updating: false,
      changingPassword: false,
      stats: {
        voice_clones: 0,
        tts_tasks: 0,
        digital_humans: 0
      }
    }
  },
  methods: {
    formatDate(date) {
      if (!date) return ''
      return new Date(date).toLocaleDateString()
    },
    updateProfile() {
      // 实现更新个人信息的逻辑
    },
    changePassword() {
      // 实现修改密码的逻辑
    }
  }
}
</script>

<style scoped>
/* 添加会员中心入口样式 */
.membership-entry {
  margin-top: 15px;
}

.membership-link {
  text-decoration: none;
}
</style>