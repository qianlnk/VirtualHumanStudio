import axios from 'axios'

// 获取会员计划列表
export const getMembershipPlans = async () => {
  try {
    const response = await axios.get('/api/membership/plans')
    return { success: true, plans: response.data }
  } catch (error) {
    console.error('获取会员计划失败:', error)
    var errorMsg = '获取会员计划失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 获取用户会员信息
export const getUserMembership = async () => {
  try {
    const response = await axios.get('/api/membership/user')
    return { success: true, membership: response.data }
  } catch (error) {
    console.error('获取会员信息失败:', error)
    var errorMsg = '获取会员信息失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 购买会员
export const purchaseMembership = async (data) => {
  try {
    const response = await axios.post('/api/membership/purchase', data)
    return { success: true, membership: response.data }
  } catch (error) {
    console.error('购买会员失败:', error)
    var errorMsg = '购买会员失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 取消自动续费
export const cancelAutoRenew = async () => {
  try {
    const response = await axios.post('/api/membership/cancel')
    return { success: true, message: response.data.message }
  } catch (error) {
    console.error('取消自动续费失败:', error)
    var errorMsg = '取消自动续费失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 获取每日使用量
export const getDailyUsage = async () => {
  try {
    const response = await axios.get('/api/membership/daily-usage')
    return { success: true, usage: response.data }
  } catch (error) {
    console.error('获取使用量失败:', error)
    var errorMsg = '获取使用量失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 获取用户待审核订单
export const getUserPendingOrders = async () => {
  try {
    const response = await axios.get('/api/membership/pending-orders')
    return { success: true, orders: response.data }
  } catch (error) {
    console.error('获取待审核订单失败:', error)
    var errorMsg = '获取待审核订单失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 管理员获取所有待审核订单
export const getAllPendingOrders = async () => {
  try {
    const response = await axios.get('/api/admin/membership/orders/pending')
    return { success: true, orders: response.data }
  } catch (error) {
    console.error('获取所有待审核订单失败:', error)
    var errorMsg = '获取所有待审核订单失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 管理员审核通过订单
export const approveOrder = async (orderId) => {
  try {
    const response = await axios.post(`/api/admin/membership/orders/${orderId}/approve`)
    return { success: true, data: response.data }
  } catch (error) {
    console.error('审核通过订单失败:', error)
    var errorMsg = '审核通过订单失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 管理员拒绝订单
export const rejectOrder = async (orderId, rejectReason) => {
  try {
    const response = await axios.post(`/api/admin/membership/orders/${orderId}/reject`, {
      reject_reason: rejectReason
    })
    return { success: true, data: response.data }
  } catch (error) {
    console.error('拒绝订单失败:', error)
    var errorMsg = '拒绝订单失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 管理员获取所有订单（支持筛选）
export const getAllOrders = async (params) => {
  try {
    const response = await axios.get('/api/admin/membership/orders', { params })
    return { success: true, orders: response.data }
  } catch (error) {
    console.error('获取所有订单失败:', error)
    var errorMsg = '获取所有订单失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
}

// 获取用户订单历史记录
export const getUserOrderHistory = async () => {
  try {
    // 使用正确的API路径获取完整订单历史
    const response = await axios.get('/api/membership/orders/history')
    return { success: true, orders: response.data }
  } catch (error) {
    console.error('获取订单历史记录失败:', error)
    var errorMsg = '获取订单历史记录失败'
    if (error.response && error.response.data && error.response.data.error) {
      errorMsg = error.response.data.error
    }
    return { success: false, message: errorMsg }
  }
} 