<template>
  <div class="voice-clone-container">
    <div class="page-header">
      <h2>音色克隆</h2>
      <el-button type="primary" @click="createVoiceClone">创建音色克隆任务</el-button>
    </div>
    
    <!-- 任务列表 -->
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0" description="暂无音色克隆任务"></el-empty>
      
      <el-table v-else :data="tasks" style="width: 100%">
        <el-table-column prop="name" label="任务名称" width="180"></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template slot-scope="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template slot-scope="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <el-button type="text" size="small" @click="viewDetail(scope.row.id)">查看</el-button>
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
    
    <!-- 音色克隆创建表单 -->
    <el-dialog title="创建音色克隆任务" :visible.sync="dialogVisible" width="600px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称"></el-input>
        </el-form-item>
        
        <el-form-item label="音频文件" prop="audio_file">
          <el-upload
            class="upload-demo"
            action="#"
            :http-request="uploadAudio"
            :limit="1"
            :file-list="fileList"
            :before-upload="beforeUpload">
            <el-button size="small" type="primary">点击上传</el-button>
            <div slot="tip" class="el-upload__tip">只能上传mp3/wav文件，且不超过50MB</div>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            type="textarea" 
            v-model="form.description" 
            placeholder="请输入任务描述"
            :rows="3">
          </el-input>
        </el-form-item>

        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="form.model_name" placeholder="请选择模型名称"></el-input>
        </el-form-item>

        <el-form-item label="提示文本" prop="prompt_text">
          <el-input 
            type="textarea" 
            v-model="form.prompt_text" 
            placeholder="请输入提示文本"
            :rows="3">
          </el-input>
        </el-form-item>

        <el-form-item label="说话人名称" prop="speaker_name">
          <el-input v-model="form.speaker_name" placeholder="请输入说话人名称"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogVisible = false">取 消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'VoiceClone',
  data() {
    return {
      loading: false,
      submitting: false,
      tasks: [],
      currentPage: 1,
      pageSize: 10,
      total: 0,
      dialogVisible: false,
      fileList: [],
      form: {
        name: '',
        description: '',
        audio_file: null,
        model_name: '',
        prompt_text: '',
        speaker_name: ''
      },
      rules: {
        name: [
          { required: true, message: '请输入任务名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        audio_file: [
          { required: true, message: '请上传音频文件', trigger: 'change' }
        ],
        model_name: [
          { required: true, message: '请选择模型名称', trigger: 'change' }
        ],
        prompt_text: [
          { required: true, message: '请输入提示文本', trigger: 'blur' }
        ],
        speaker_name: [
          { required: true, message: '请输入说话人名称', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.fetchTasks()
  },
  methods: {
    // 获取音色克隆任务列表
    fetchTasks() {
      this.loading = true
      this.$http.get(`/api/voice/clones?page=${this.currentPage}&size=${this.pageSize}`)
        .then(response => {
          this.tasks = response.data.voice_clones || []
          this.total = response.data.total || 0
        })
        .catch(error => {
          console.error('获取音色克隆任务列表失败', error)
          this.$message.error('获取音色克隆任务列表失败')
        })
        .finally(() => {
          this.loading = false
        })
    },
    
    // 格式化日期
    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },
    
    // 获取状态类型
    getStatusType(status) {
      const statusMap = {
        'pending': 'info',
        'processing': 'warning',
        'completed': 'success',
        'failed': 'danger'
      }
      return statusMap[status] || 'info'
    },
    
    // 创建音色克隆任务
    createVoiceClone() {
      this.dialogVisible = true
      this.form = {
        name: '',
        description: '',
        audio_file: null,
        model_name: '',
        prompt_text: '',
        speaker_name: ''
      }
      this.fileList = []
    },
    
    // 上传前检查音频文件
    beforeUpload(file) {
      const isAudio = file.type.startsWith('audio/') || file.name.endsWith('.mp3') || file.name.endsWith('.wav')
      const isLt50M = file.size / 1024 / 1024 < 50
      
      if (!isAudio) {
        this.$message.error('只能上传音频文件!')
        return false
      }
      if (!isLt50M) {
        this.$message.error('音频文件大小不能超过 50MB!')
        return false
      }
      
      this.form.audio_file = file
      return false
    },
    
    // 上传音频文件
    uploadAudio(options) {
      this.form.audio_file = options.file
      this.fileList = [{ name: options.file.name, url: '' }]
    },
    
    // 提交表单
    submitForm() {
      this.$refs.form.validate(valid => {
        if (valid) {
          if (!this.form.audio_file) {
            return this.$message.error('请上传音频文件')
          }
          
          this.submitting = true
          
          const formData = new FormData()
          formData.append('name', this.form.name)
          formData.append('description', this.form.description)
          formData.append('prompt_file', this.form.audio_file)
          formData.append('model_name', this.form.model_name)
          formData.append('prompt_text', this.form.prompt_text)
          formData.append('speaker_name', this.form.speaker_name)
          
          this.$http.post('/api/voice/clone', formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            }
          })
            .then(response => {
              this.$message.success('创建音色克隆任务成功')
              this.dialogVisible = false
              this.fetchTasks()
              
              // 跳转到详情页
              this.$router.push(`/voice-clone/${response.data.voice_clone.id}`)
            })
            .catch(error => {
              console.error('创建音色克隆任务失败', error)
              this.$message.error('创建音色克隆任务失败: ' + ((error.response && error.response.data && error.response.data.message) || '未知错误'))
            })
            .finally(() => {
              this.submitting = false
            })
        }
      })
    },
    
    // 查看详情
    viewDetail(id) {
      this.$router.push(`/voice-clone/${id}`)
    },
    
    // 确认删除
    confirmDelete(id) {
      this.$confirm('此操作将永久删除该音色克隆任务, 是否继续?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() => {
          this.deleteTask(id)
        })
        .catch(() => {
          this.$message.info('已取消删除')
        })
    },
    
    // 删除任务
    deleteTask(id) {
      this.$http.delete(`/api/voice/clone/${id}`)
        .then(() => {
          this.$message.success('删除成功')
          this.fetchTasks()
        })
        .catch(error => {
          console.error('删除失败', error)
          this.$message.error('删除失败')
        })
    },
    
    // 处理页码变化
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchTasks()
    },
    
    // 处理每页显示数量变化
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchTasks()
    }
  }
}
</script>

<style scoped>
.voice-clone-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 22px;
  color: #303133;
}

.task-list {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  text-align: center;
}
</style>