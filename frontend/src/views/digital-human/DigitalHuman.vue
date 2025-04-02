<template>
  <div class="digital-human-container">
    <div class="page-header">
      <h2>数字人合成</h2>
      <el-button type="primary" @click="showCreateDialog">创建数字人合成任务</el-button>
    </div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="digitalHumans.length === 0" description="暂无数字人合成任务"></el-empty>
      
      <el-table v-else :data="digitalHumans" style="width: 100%">
        <el-table-column prop="name" label="任务名称" width="800"></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="200">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250">
          <template slot-scope="scope">
            <el-button 
              type="text" 
              size="small" 
              @click="viewDetail(scope.row.id)"
            >查看</el-button>
            <el-button 
              type="text" 
              size="small" 
              @click="downloadResult(scope.row.id)" 
              :disabled="scope.row.status !== 'completed'"
            >下载</el-button>
            <el-button type="text" size="small" @click="confirmDelete(scope.row.id)">删除</el-button>
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
    
    <!-- 创建数字人任务对话框 -->
    <el-dialog title="创建数字人合成任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        
        <el-form-item label="音频文件" prop="audio_file">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleAudioChange"
            :limit="1"
            :file-list="audioFileList">
            <el-button size="small" type="primary">选择音频文件</el-button>
            <div slot="tip" class="el-upload__tip">只能上传WAV/MP3文件</div>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="视频文件" prop="video_file">
          <el-upload
            class="upload-demo"
            action="#"
            :auto-upload="false"
            :on-change="handleVideoChange"
            :limit="1"
            :file-list="videoFileList">
            <el-button size="small" type="primary">选择视频文件</el-button>
            <div slot="tip" class="el-upload__tip">只能上传MP4文件</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">创建</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import axios from 'axios'
import { v4 as uuidv4 } from 'uuid'

export default {
  name: 'DigitalHuman',
  data() {
    return {
      loading: false,
      submitting: false,
      dialogVisible: false,
      digitalHumans: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      form: {
        name: '',
        description: '',
        task_code: '',
        chaofen: 0,
        watermark_switch: 0,
        pn: 1,
        audio_file: null,
        video_file: null
      },
      rules: {
        audio_file: [
          { required: true, message: '请上传音频文件', trigger: 'change' }
        ],
        video_file: [
          { required: true, message: '请上传视频文件', trigger: 'change' }
        ]
      },
      audioFileList: [],
      videoFileList: []
    }
  },
  created() {
    this.fetchDigitalHumans()
  },
  methods: {
    // 获取数字人列表
    fetchDigitalHumans() {
      this.loading = true
      axios.get('/api/digital-human', {
        params: {
          page: this.currentPage,
          page_size: this.pageSize
        }
      })
        .then(response => {
          this.digitalHumans = response.data.digital_humans || []
          this.total = response.data.total || 0
        })
        .catch(error => {
          console.error('获取数字人列表失败:', error)
          this.$message.error('获取数字人列表失败')
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 显示创建对话框
    showCreateDialog() {
      this.dialogVisible = true
      this.resetForm()
    },
    
    // 重置表单
    resetForm() {
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
      const uuid = uuidv4()
      this.form = {
        name: uuid,
        description: '',
        task_code: uuid,
        chaofen: 0,
        watermark_switch: 0,
        pn: 1,
        audio_file: null,
        video_file: null
      }
      this.audioFileList = []
      this.videoFileList = []
    },
    
    // 处理音频文件变化
    handleAudioChange(file) {
      this.form.audio_file = file.raw
    },
    
    // 处理视频文件变化
    handleVideoChange(file) {
      this.form.video_file = file.raw
    },
    
    // 提交表单
    submitForm() {
      this.$refs.form.validate(valid => {
        if (valid) {
          this.createDigitalHuman()
        } else {
          return false
        }
      })
    },
    
    // 创建数字人任务
    createDigitalHuman() {
      this.submitting = true
      
      // 创建FormData对象
      const formData = new FormData()
      formData.append('name', this.form.name)
      formData.append('description', this.form.description)
      formData.append('task_code', this.form.task_code)
      formData.append('chaofen', this.form.chaofen)
      formData.append('watermark_switch', this.form.watermark_switch)
      formData.append('pn', this.form.pn)
      formData.append('audio_file', this.form.audio_file)
      formData.append('video_file', this.form.video_file)
      
      // 发送请求
      axios.post('/api/digital-human', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
        .then(() => {
          this.$message.success('创建数字人任务成功')
          this.dialogVisible = false
          this.fetchDigitalHumans()
        })
        .catch(error => {
          console.error('创建数字人任务失败:', error)
          this.$message.error('创建数字人任务失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
        .finally(() => {
          this.submitting = false
        })
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/digital-human/${id}`)
    },
    
    // 下载结果
    downloadResult(id) {
      window.open(`/api/digital-human/${id}/download`, '_blank')
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('确定要删除这个数字人任务吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.deleteDigitalHuman(id)
      }).catch(() => {
        // 取消删除
      })
    },
    
    // 删除数字人任务
    deleteDigitalHuman(id) {
      axios.delete(`/api/digital-human/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          this.fetchDigitalHumans()
        })
        .catch(error => {
          console.error('删除失败:', error)
          this.$message.error('删除失败: ' + ((error.response && error.response.data && error.response.data.error) || error.message))
        })
    },
    
    // 处理分页大小变化
    handleSizeChange(size) {
      this.pageSize = size
      this.fetchDigitalHumans()
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchDigitalHumans()
    },
    
    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    
    // 获取状态类型
    getStatusType(status) {
      switch (status) {
        case 'completed': return 'success'
        case 'processing': return 'warning'
        case 'pending': return 'info'
        case 'failed': return 'danger'
        default: return 'info'
      }
    },
    
    // 获取状态文本
    getStatusText(status) {
      switch (status) {
        case 'completed': return '已完成'
        case 'processing': return '处理中'
        case 'pending': return '等待中'
        case 'failed': return '失败'
        default: return '未知'
      }
    }
  }
}
</script>

<style scoped>
.digital-human-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.task-list {
  background-color: #fff;
  padding: 20px;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.status-tag {
  margin-left: 10px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.dialog-footer {
  text-align: right;
}

.upload-demo {
  width: 100%;
}
</style>