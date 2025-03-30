<template>
  <div class="admin-users-container">
    <div class="page-header">
      <h2>用户管理</h2>
    </div>
    
    <div v-loading="loading" class="users-list">
      <el-table :data="users" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="username" label="用户名" width="180"></el-table-column>
        <el-table-column prop="email" label="邮箱" width="220"></el-table-column>
        <el-table-column prop="phone" label="手机号" width="150"></el-table-column>
        <el-table-column prop="role" label="角色" width="100">
          <template slot-scope="scope">
            <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'success'">
              {{ scope.row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
              {{ scope.row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="toggleUserStatus(scope.row)"
              :disabled="scope.row.role === 'admin'"
            >
              {{ scope.row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-container" v-if="total > pageSize">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="currentPage"
          :page-sizes="[10, 20, 30, 50]"
          :page-size="pageSize"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total">
        </el-pagination>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AdminUsers',
  data() {
    return {
      loading: false,
      users: [],
      currentPage: 1,
      pageSize: 10,
      total: 0
    }
  },
  created() {
    this.fetchUsers()
  },
  methods: {
    // 获取用户列表
    fetchUsers() {
      this.loading = true
      this.$http.get(`/api/admin/users?page=${this.currentPage}&size=${this.pageSize}`)
        .then(response => {
          this.users = response.data.users || []
          this.total = response.data.total || 0
        })
        .catch(error => {
          console.error('获取用户列表失败', error)
          this.$message.error('获取用户列表失败')
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 切换用户状态
    toggleUserStatus(user) {
      const newStatus = user.status === 'active' ? 'inactive' : 'active'
      const statusText = newStatus === 'active' ? '启用' : '禁用'
      
      this.$confirm(`确定要${statusText}用户 ${user.username} 吗?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => {
          this.$http.put(`/api/admin/user/${user.id}/status`, {
            status: newStatus
          })
            .then(() => {
              this.$message.success(`${statusText}用户成功`)
              // 更新本地数据
              user.status = newStatus
            })
            .catch(error => {
              console.error(`${statusText}用户失败`, error)
              this.$message.error(`${statusText}用户失败`)
            })
        })
        .catch(() => {
          this.$message.info('已取消操作')
        })
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchUsers()
    },
    
    // 处理每页显示数量变化
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchUsers()
    }
  }
}
</script>

<style scoped>
.admin-users-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.users-list {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: center;
}
</style>