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

/**
 * 点赞灵感内容
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 点赞结果
 */
export const likeShareTask = (shareTaskId) => {
  return axios.post(`/api/share/like`, { share_task_id: shareTaskId })
    .then(response => ({
      success: true,
      message: response.data.message || '点赞成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '点赞失败'
    }))
}

/**
 * 取消点赞灵感内容
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 取消点赞结果
 */
export const unlikeShareTask = (shareTaskId) => {
  return axios.delete(`/api/share/like/${shareTaskId}`)
    .then(response => ({
      success: true,
      message: response.data.message || '取消点赞成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '取消点赞失败'
    }))
}

/**
 * 收藏灵感内容
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 收藏结果
 */
export const favoriteShareTask = (shareTaskId) => {
  return axios.post(`/api/share/favorite`, { share_task_id: shareTaskId })
    .then(response => ({
      success: true,
      message: response.data.message || '收藏成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '收藏失败'
    }))
}

/**
 * 取消收藏灵感内容
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 取消收藏结果
 */
export const unfavoriteShareTask = (shareTaskId) => {
  return axios.delete(`/api/share/favorite/${shareTaskId}`)
    .then(response => ({
      success: true,
      message: response.data.message || '取消收藏成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '取消收藏失败'
    }))
}

/**
 * 添加评论
 * @param {number} shareTaskId - 分享任务ID
 * @param {string} content - 评论内容
 * @returns {Promise<Object>} 评论结果
 */
export const addComment = (shareTaskId, content) => {
  return axios.post(`/api/share/comment`, {
    share_task_id: shareTaskId,
    content: content
  })
    .then(response => ({
      success: true,
      message: response.data.message || '评论成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '评论失败'
    }))
}

/**
 * 删除评论
 * @param {number} commentId - 评论ID
 * @returns {Promise<Object>} 删除结果
 */
export const deleteComment = (commentId) => {
  return axios.delete(`/api/share/comment/${commentId}`)
    .then(response => ({
      success: true,
      message: response.data.message || '删除评论成功'
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '删除评论失败'
    }))
}

/**
 * 获取点赞列表
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 点赞列表
 */
export const getLikes = (shareTaskId) => {
  return axios.get(`/api/share/likes/${shareTaskId}`)
    .then(response => ({
      success: true,
      userInfos: response.data.user_infos || []
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '获取点赞列表失败'
    }))
}

/**
 * 获取收藏列表
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 收藏列表
 */
export const getFavorites = (shareTaskId) => {
  return axios.get(`/api/share/favorites/${shareTaskId}`)
    .then(response => ({
      success: true,
      userInfos: response.data.user_infos || []
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '获取收藏列表失败'
    }))
}

/**
 * 获取评论列表
 * @param {number} shareTaskId - 分享任务ID
 * @returns {Promise<Object>} 评论列表
 */
export const getComments = (shareTaskId) => {
  return axios.get(`/api/share/comments/${shareTaskId}`)
    .then(response => ({
      success: true,
      comments: response.data.comments || [],
      userInfos: response.data.user_infos || []
    }))
    .catch(error => ({
      success: false,
      message: (error.response && error.response.data && error.response.data.error) || '获取评论列表失败'
    }))
}