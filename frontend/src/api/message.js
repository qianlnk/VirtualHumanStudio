import axios from 'axios'

// 创建留言
export const createMessage = (messageData) => {
    return axios.post('/api/message', messageData)
        .then(response => ({
            success: true,
            data: response.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '提交留言失败'
        }))
}

// 获取留言列表
export const getMessages = (page = 1, limit = 10) => {
    return axios.get(`/api/messages?page=${page}&limit=${limit}`)
        .then(response => ({
            success: true,
            data: response.data.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '获取留言列表失败'
        }))
}

// 获取单个留言
export const getMessage = (id) => {
    return axios.get(`/api/message/${id}`)
        .then(response => ({
            success: true,
            data: response.data.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '获取留言详情失败'
        }))
}

// 删除留言
export const deleteMessage = (id) => {
    return axios.delete(`/api/message/${id}`)
        .then(response => ({
            success: true,
            data: response.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '删除留言失败'
        }))
}

// 管理员专用：回复留言
export const replyMessage = (id, replyText) => {
    return axios.put(`/api/admin/message/${id}/reply`, { reply_text: replyText })
        .then(response => ({
            success: true,
            data: response.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '回复留言失败'
        }))
}

// 管理员专用：标记留言为已读
export const markMessageAsRead = (id) => {
    return axios.put(`/api/admin/message/${id}/read`)
        .then(response => ({
            success: true,
            data: response.data
        }))
        .catch(error => ({
            success: false,
            error: (error.response && error.response.data && error.response.data.message) || '标记留言为已读失败'
        }))
} 