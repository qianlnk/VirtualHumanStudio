<template>
  <div class="membership-orders-container">
    <div class="page-header">
      <h2>会员订单管理</h2>
      <el-button 
        type="primary" 
        icon="el-icon-refresh" 
        class="refresh-btn" 
        size="small"
        @click="fetchOrders" 
        :loading="loading">
        刷新数据
      </el-button>
    </div>
    
    <!-- 查询条件 -->
    <el-card shadow="hover" class="filter-card">
      <el-form :model="filterForm" :inline="true" class="filter-form">
        <el-form-item label="订单状态">
          <el-select v-model="filterForm.status" placeholder="选择状态" clearable>
            <el-option label="待审核" value="pending"></el-option>
            <el-option label="已通过" value="approved"></el-option>
            <el-option label="已拒绝" value="rejected"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="filterForm.username" placeholder="输入用户名" clearable></el-input>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="yyyy-MM-dd"
            :picker-options="pickerOptions">
          </el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <!-- 订单表格 -->
    <el-card shadow="hover" class="orders-card">
      <div slot="header" class="card-header">
        <span>订单列表</span>
        <span class="pending-count" v-if="pendingOrdersCount > 0">
          {{ pendingOrdersCount }} 个订单待处理
        </span>
      </div>

      <el-table 
        :data="orders" 
        stripe 
        style="width: 100%"
        v-loading="loading"
        empty-text="暂无订单数据">
        <el-table-column prop="id" label="订单ID" width="80"></el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180">
          <template slot-scope="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="user.username" label="用户名" width="120"></el-table-column>
        <el-table-column prop="plan_name" label="会员方案" width="120"></el-table-column>
        <el-table-column prop="payment_method" label="支付方式" width="100">
          <template slot-scope="scope">
            {{ scope.row.payment_method === 'wechat' ? '微信支付' : '支付宝' }}
          </template>
        </el-table-column>
        <el-table-column prop="price" label="金额" width="100">
          <template slot-scope="scope">
            ¥{{ scope.row.price }}
          </template>
        </el-table-column>
        <el-table-column prop="payment_remark" label="支付备注"></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="scope">
            <el-tag type="warning" v-if="scope.row.status === 'pending'">待审核</el-tag>
            <el-tag type="success" v-else-if="scope.row.status === 'approved'">已通过</el-tag>
            <el-tag type="danger" v-else-if="scope.row.status === 'rejected'">已拒绝</el-tag>
            <el-tag v-else>{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template slot-scope="scope">
            <el-button
              size="mini"
              type="success"
              @click="handleApprove(scope.row)"
              :loading="scope.row.approving"
              v-if="scope.row.status === 'pending'">
              通过
            </el-button>
            <el-button
              size="mini"
              type="danger"
              @click="handleReject(scope.row)"
              :loading="scope.row.rejecting"
              v-if="scope.row.status === 'pending'">
              拒绝
            </el-button>
            <span v-else>
              {{ scope.row.status === 'approved' ? '已通过' : '已拒绝' }}
              {{ scope.row.status === 'rejected' && scope.row.reject_reason ? ' - ' + scope.row.reject_reason : '' }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 拒绝订单对话框 -->
    <el-dialog
      title="拒绝订单"
      :visible.sync="rejectDialogVisible"
      width="500px">
      <div v-if="currentOrder">
        <p class="reject-info">
          您即将拒绝订单 <strong>#{{ currentOrder.id }}</strong>，
          用户 <strong>{{ currentOrder.user && currentOrder.user.username }}</strong> 的
          <strong>{{ currentOrder.plan_name }}</strong> 方案购买请求。
        </p>
        
        <el-form :model="rejectForm" label-width="80px">
          <el-form-item label="拒绝原因" required>
            <el-input
              type="textarea"
              v-model="rejectForm.reason"
              :rows="4"
              placeholder="请输入拒绝原因，该信息将可能显示给用户">
            </el-input>
          </el-form-item>
        </el-form>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmReject" :loading="processing">确认拒绝</el-button>
      </span>
    </el-dialog>
    
    <!-- 批准订单确认对话框 -->
    <el-dialog
      title="批准订单"
      :visible.sync="approveDialogVisible"
      width="500px">
      <div v-if="currentOrder">
        <p class="approve-info">
          您即将批准订单 <strong>#{{ currentOrder.id }}</strong>，
          用户 <strong>{{ currentOrder.user && currentOrder.user.username }}</strong> 的
          <strong>{{ currentOrder.plan_name }}</strong> 方案购买请求。
        </p>
        <div class="order-details">
          <p><strong>订单金额：</strong> ¥{{ currentOrder.price }}</p>
          <p><strong>支付方式：</strong> {{ currentOrder.payment_method === 'wechat' ? '微信支付' : '支付宝' }}</p>
          <p><strong>支付备注：</strong> {{ currentOrder.payment_remark }}</p>
          <p><strong>提交时间：</strong> {{ formatDateTime(currentOrder.created_at) }}</p>
        </div>
        <div class="warning-box">
          <i class="el-icon-warning"></i>
          <span>确认批准后，系统将自动为用户开通相应会员权限。请确认已收到该笔付款。</span>
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button type="success" @click="confirmApprove" :loading="processing">确认批准</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getAllOrders, approveOrder, rejectOrder } from '@/api/membership'

export default {
  name: 'MembershipOrders',
  data() {
    return {
      orders: [],
      pendingOrdersCount: 0,
      loading: false,
      rejectDialogVisible: false,
      approveDialogVisible: false,
      currentOrder: null,
      rejectForm: {
        reason: ''
      },
      processing: false,
      filterForm: {
        status: 'pending', // 默认显示待审核订单
        username: '',
      },
      dateRange: [],
      pickerOptions: {
        shortcuts: [{
          text: '最近一周',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
            picker.$emit('pick', [start, end]);
          }
        }, {
          text: '最近一个月',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
            picker.$emit('pick', [start, end]);
          }
        }, {
          text: '最近三个月',
          onClick(picker) {
            const end = new Date();
            const start = new Date();
            start.setTime(start.getTime() - 3600 * 1000 * 24 * 90);
            picker.$emit('pick', [start, end]);
          }
        }]
      }
    }
  },
  created() {
    // 检查用户是否为管理员
    if (!this.$store.getters.isAdmin) {
      this.$message.error('权限不足，只有管理员可以访问此页面')
      this.$router.push('/')
      return
    }
    
    this.fetchOrders()
  },
  methods: {
    // 获取订单数据
    async fetchOrders() {
      this.loading = true
      try {
        // 构建查询参数
        const params = {}
        if (this.filterForm.status) params.status = this.filterForm.status
        if (this.filterForm.username) params.username = this.filterForm.username
        if (this.dateRange && this.dateRange.length === 2) {
          params.start_date = this.dateRange[0]
          params.end_date = this.dateRange[1]
        }
        
        const response = await getAllOrders(params)
        if (response.success) {
          // 添加操作状态标记
          this.orders = response.orders.map(order => ({
            ...order,
            approving: false,
            rejecting: false
          }))
          // 计算待处理订单数量
          this.pendingOrdersCount = this.orders.filter(order => order.status === 'pending').length
        } else {
          this.$message.error(response.message)
        }
      } catch (error) {
        console.error('获取订单失败:', error)
        this.$message.error('获取订单失败')
      } finally {
        this.loading = false
      }
    },
    
    // 查询按钮点击
    handleSearch() {
      this.fetchOrders()
    },
    
    // 重置查询条件
    handleReset() {
      this.filterForm = {
        status: 'pending',
        username: ''
      }
      this.dateRange = []
      this.fetchOrders()
    },
    
    // 处理批准订单
    handleApprove(order) {
      this.currentOrder = order
      this.approveDialogVisible = true
    },
    
    // 处理拒绝订单
    handleReject(order) {
      this.currentOrder = order
      this.rejectForm.reason = ''
      this.rejectDialogVisible = true
    },
    
    // 确认批准订单
    async confirmApprove() {
      if (!this.currentOrder) return
      
      this.processing = true
      // 更新表格中对应订单的loading状态
      this.updateOrderLoadingState(this.currentOrder.id, 'approving', true)
      
      try {
        const response = await approveOrder(this.currentOrder.id)
        if (response.success) {
          this.$message.success('订单已批准，会员已激活')
          this.approveDialogVisible = false
          // 刷新订单列表
          this.fetchOrders()
        } else {
          this.$message.error(response.message)
          // 恢复loading状态
          this.updateOrderLoadingState(this.currentOrder.id, 'approving', false)
        }
      } catch (error) {
        console.error('批准订单失败:', error)
        this.$message.error('批准订单失败')
        // 恢复loading状态
        this.updateOrderLoadingState(this.currentOrder.id, 'approving', false)
      } finally {
        this.processing = false
      }
    },
    
    // 确认拒绝订单
    async confirmReject() {
      if (!this.currentOrder) return
      
      if (!this.rejectForm.reason.trim()) {
        this.$message.warning('请输入拒绝原因')
        return
      }
      
      this.processing = true
      // 更新表格中对应订单的loading状态
      this.updateOrderLoadingState(this.currentOrder.id, 'rejecting', true)
      
      try {
        const response = await rejectOrder(this.currentOrder.id, this.rejectForm.reason)
        if (response.success) {
          this.$message.success('订单已拒绝')
          this.rejectDialogVisible = false
          // 刷新订单列表
          this.fetchOrders()
        } else {
          this.$message.error(response.message)
          // 恢复loading状态
          this.updateOrderLoadingState(this.currentOrder.id, 'rejecting', false)
        }
      } catch (error) {
        console.error('拒绝订单失败:', error)
        this.$message.error('拒绝订单失败')
        // 恢复loading状态
        this.updateOrderLoadingState(this.currentOrder.id, 'rejecting', false)
      } finally {
        this.processing = false
      }
    },
    
    // 更新订单的loading状态
    updateOrderLoadingState(orderId, stateKey, value) {
      const orderIndex = this.orders.findIndex(order => order.id === orderId)
      if (orderIndex !== -1) {
        this.orders[orderIndex][stateKey] = value
      }
    },
    
    // 格式化日期时间
    formatDateTime(dateTime) {
      if (!dateTime) return '';
      const date = new Date(dateTime);
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
    }
  }
}
</script>

<style scoped>
.membership-orders-container {
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
  color: #409EFF;
  font-weight: 400;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pending-count {
  font-size: 14px;
  color: #E6A23C;
  font-weight: bold;
}

.orders-card {
  margin-bottom: 20px;
}

.reject-info, .approve-info {
  margin-bottom: 20px;
  line-height: 1.5;
}

.order-details {
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.order-details p {
  margin: 8px 0;
}

.warning-box {
  background-color: #fff6f7;
  border-left: 3px solid #F56C6C;
  padding: 10px 15px;
  margin: 15px 0;
  display: flex;
  align-items: center;
}

.warning-box i {
  color: #F56C6C;
  font-size: 18px;
  margin-right: 10px;
}
</style> 