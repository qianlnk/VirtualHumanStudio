<template>
  <div class="messages-container">
    <div class="page-header">
      <h2>留言管理</h2>
    </div>

    <el-card class="messages-card">
      <div slot="header" class="card-header">
        <span>用户留言列表</span>
        <el-button 
          size="small" 
          type="primary" 
          icon="el-icon-refresh" 
          @click="loadMessages"
          :loading="loading"
        >刷新</el-button>
      </div>

      <!-- 留言列表 -->
      <el-table 
        v-loading="loading" 
        :data="messages" 
        stripe 
        border 
        style="width: 100%"
        @row-click="handleRowClick"
      >
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="name" label="姓名" width="120"></el-table-column>
        <el-table-column prop="subject" label="主题" width="180"></el-table-column>
        <el-table-column prop="email" label="邮箱" width="180"></el-table-column>
        <el-table-column prop="phone" label="电话" width="140"></el-table-column>
        <el-table-column label="状态" width="100">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status === 'unread'" type="danger">未读</el-tag>
            <el-tag v-else-if="scope.row.status === 'read'" type="warning">已读</el-tag>
            <el-tag v-else-if="scope.row.status === 'replied'" type="success">已回复</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="留言时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template slot-scope="scope">
            <el-button 
              v-if="scope.row.status === 'unread'" 
              type="text" 
              @click.stop="markAsRead(scope.row)"
            >标为已读</el-button>
            <el-button 
              v-if="scope.row.status !== 'replied'" 
              type="text" 
              @click.stop="showReplyDialog(scope.row)"
            >回复</el-button>
            <el-button 
              type="text" 
              class="danger-text" 
              @click.stop="confirmDelete(scope.row)"
            >删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :current-page="page"
          :page-sizes="[10, 20, 50, 100]"
          :page-size="limit"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
        ></el-pagination>
      </div>
    </el-card>

    <!-- 查看/回复留言对话框 -->
    <el-dialog 
      :title="currentMessage.status === 'replied' ? '查看留言详情' : '回复留言'" 
      :visible.sync="dialogVisible" 
      width="50%"
    >
      <div class="message-detail" v-if="currentMessage">
        <div class="message-info">
          <p><strong>留言人：</strong>{{ currentMessage.name }}</p>
          <p><strong>联系方式：</strong>{{ currentMessage.phone }}</p>
          <p><strong>邮箱：</strong>{{ currentMessage.email }}</p>
          <p><strong>主题：</strong>{{ currentMessage.subject }}</p>
          <p><strong>留言时间：</strong>{{ formatDate(currentMessage.created_at) }}</p>
          <div class="message-content">
            <strong>留言内容：</strong>
            <p>{{ currentMessage.content }}</p>
          </div>
          
          <div class="reply-section" v-if="currentMessage.status === 'replied'">
            <div class="reply-info">
              <strong>回复内容：</strong>
              <p>{{ currentMessage.reply_text }}</p>
              <p class="reply-time">回复时间：{{ formatDate(currentMessage.reply_time) }}</p>
            </div>
          </div>
        </div>
        
        <div class="reply-form" v-if="currentMessage.status !== 'replied'">
          <el-form :model="replyForm" ref="replyForm" :rules="replyRules">
            <el-form-item label="回复内容" prop="reply_text">
              <el-input 
                type="textarea" 
                v-model="replyForm.reply_text" 
                :rows="5" 
                placeholder="请输入回复内容"
              ></el-input>
            </el-form-item>
          </el-form>
        </div>
      </div>
      
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button 
          v-if="currentMessage.status !== 'replied'" 
          type="primary" 
          @click="submitReply" 
          :loading="submitting"
        >提交回复</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getMessages, deleteMessage, markMessageAsRead, replyMessage } from '@/api/message'

export default {
  name: 'MessagesManagement',
  data() {
    return {
      loading: false,
      submitting: false,
      page: 1,
      limit: 10,
      total: 0,
      messages: [],
      dialogVisible: false,
      currentMessage: {},
      replyForm: {
        reply_text: ''
      },
      replyRules: {
        reply_text: [
          { required: true, message: '请输入回复内容', trigger: 'blur' },
          { min: 5, max: 500, message: '长度在 5 到 500 个字符', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.loadMessages()
  },
  methods: {
    async loadMessages() {
      try {
        this.loading = true
        const result = await getMessages(this.page, this.limit)
        
        if (result.success) {
          this.messages = result.data.list
          this.total = result.data.total
        } else {
          this.$message.error(result.error || '获取留言列表失败')
        }
      } catch (error) {
        console.error('加载留言失败:', error)
        this.$message.error('系统错误，请稍后再试')
      } finally {
        this.loading = false
      }
    },
    
    handleSizeChange(val) {
      this.limit = val
      this.loadMessages()
    },
    
    handleCurrentChange(val) {
      this.page = val
      this.loadMessages()
    },
    
    handleRowClick(row) {
      this.currentMessage = row
      this.dialogVisible = true
      this.replyForm.reply_text = ''
    },
    
    showReplyDialog(row) {
      this.currentMessage = row
      this.dialogVisible = true
      this.replyForm.reply_text = ''
    },
    
    async markAsRead(row) {
      try {
        const result = await markMessageAsRead(row.id)
        
        if (result.success) {
          this.$message.success('标记为已读成功')
          // 更新本地数据状态
          const index = this.messages.findIndex(item => item.id === row.id)
          if (index !== -1) {
            this.messages[index].status = 'read'
          }
        } else {
          this.$message.error(result.error || '标记失败')
        }
      } catch (error) {
        console.error('标记留言为已读失败:', error)
        this.$message.error('系统错误，请稍后再试')
      }
    },
    
    confirmDelete(row) {
      this.$confirm('确定要删除这条留言吗？此操作不可逆', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteMessage(row.id)
      }).catch(() => {
        // 取消删除
      })
    },
    
    async deleteMessage(id) {
      try {
        const result = await deleteMessage(id)
        
        if (result.success) {
          this.$message.success('删除留言成功')
          // 更新本地数据，移除已删除的留言
          this.messages = this.messages.filter(item => item.id !== id)
          // 如果当前显示的是这条留言，关闭对话框
          if (this.currentMessage.id === id) {
            this.dialogVisible = false
          }
        } else {
          this.$message.error(result.error || '删除失败')
        }
      } catch (error) {
        console.error('删除留言失败:', error)
        this.$message.error('系统错误，请稍后再试')
      }
    },
    
    submitReply() {
      this.$refs.replyForm.validate(async (valid) => {
        if (valid) {
          try {
            this.submitting = true
            const result = await replyMessage(this.currentMessage.id, this.replyForm.reply_text)
            
            if (result.success) {
              this.$message.success('回复留言成功')
              this.dialogVisible = false
              // 更新本地数据状态
              const index = this.messages.findIndex(item => item.id === this.currentMessage.id)
              if (index !== -1) {
                this.messages[index].status = 'replied'
                this.messages[index].reply_text = this.replyForm.reply_text
                this.messages[index].reply_time = new Date()
              }
            } else {
              this.$message.error(result.error || '回复失败')
            }
          } catch (error) {
            console.error('回复留言失败:', error)
            this.$message.error('系统错误，请稍后再试')
          } finally {
            this.submitting = false
          }
        }
      })
    },
    
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    }
  }
}
</script>

<style scoped>
.messages-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.messages-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.message-detail {
  padding: 10px;
}

.message-info p {
  margin: 8px 0;
}

.message-content {
  margin-top: 15px;
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
}

.message-content p {
  margin-top: 10px;
  white-space: pre-wrap;
}

.reply-section {
  margin-top: 20px;
  border-top: 1px solid #dcdfe6;
  padding-top: 15px;
}

.reply-info p {
  margin-top: 10px;
  white-space: pre-wrap;
}

.reply-time {
  color: #909399;
  font-size: 12px;
  margin-top: 5px;
}

.reply-form {
  margin-top: 20px;
}

.danger-text {
  color: #f56c6c;
}

.danger-text:hover {
  color: #f78989;
}
</style> 