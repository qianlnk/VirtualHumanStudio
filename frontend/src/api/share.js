import axios from 'axios'

/**
 * 分享任务到灵感页面
 * @param {Object} data - 分享数据
 * @param {number} data.taskId - 任务ID
 * @param {string} data.mode - 模式 (comfyui或digital_human)
 * @param {string} data.taskType - 任务类型 (accessory、workflow等)
 * @returns {Promise<Object>} 分享结果
 */
export const shareTask = (data) => {
  return axios.post('/api/share', {
    task_id: data.taskId,
    mode: data.mode,
    task_type: data.taskType
  })
    .then(response => ({
      success: true,
      message: response.data.message || '分享成功，等待审核'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '分享失败'
    }))
}

/**
 * 获取灵感页面内容
 * @param {number} page - 页码
 * @param {number} size - 每页数量
 * @returns {Promise<Object>} 灵感内容
 */
export const getInspirationContent = (page = 1, size = 12) => {
  return axios.get(`/api/inspiration?page=${page}&size=${size}`)
    .then(response => ({
      success: true,
      tasks: response.data.tasks || [],
      total: response.data.total || 0
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '获取灵感内容失败'
    }))
}

/**
 * 获取待审核任务
 * @param {number} page - 页码
 * @param {number} size - 每页数量
 * @returns {Promise<Object>} 待审核任务
 */
export const getPendingReviewTasks = (page = 1, size = 12) => {
  return axios.get(`/api/admin/review/pending?page=${page}&size=${size}`)
    .then(response => ({
      success: true,
      tasks: response.data.tasks || [],
      total: response.data.total || 0
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '获取待审核任务失败'
    }))
}

/**
 * 审核任务
 * @param {Object} data - 审核数据
 * @param {number} data.shareId - 分享ID
 * @param {number} data.taskId - 任务ID
 * @param {string} data.mode - 模式 (comfyui或digital_human)
 * @param {string} data.taskType - 任务类型 (accessory、workflow等)
 * @param {string} data.status - 审核状态 (approved或rejected)
 * @param {string} data.rejectReason - 拒绝原因 (仅在status为rejected时需要)
 * @returns {Promise<Object>} 审核结果
 */
export const reviewTask = (data) => {
  return axios.post('/api/admin/review', {
    share_id: data.shareId,
    task_id: data.taskId,
    mode: data.mode,
    task_type: data.taskType,
    status: data.status,
    reject_reason: data.rejectReason || ''
  })
    .then(response => ({
      success: true,
      message: response.data.message || '审核成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '审核失败'
    }))
}