import axios from 'axios'
import store from '@/store'

// 创建axios实例
const api = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    // 从store获取认证token
    const token = store.state.token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    console.error('API请求错误:', error)
    
    // 处理认证错误
    if (error.response && error.response.status === 401) {
      // 清除store中的认证信息
      store.commit('clearAuth')
      // 跳转到登录页
      window.location.href = '/login'
    }
    
    return Promise.reject(error)
  }
)

// 聊天相关API
export const chatAPI = {
  // 创建新会话
  createSession(title) {
    return api.post('/chat/sessions', { title })
  },

  // 获取会话列表
  getSessions() {
    return api.get('/chat/sessions')
  },

  // 获取会话消息历史
  getSessionMessages(sessionId) {
    return api.get(`/chat/sessions/${sessionId}/messages`)
  },

  // 发送消息
  sendMessage(sessionId, message, models, imageUrl = '') {
    // 创建一个可取消的请求
    const controller = {
      source: axios.CancelToken.source(),
      aborted: false,
      
      // 中断请求的方法
      abort() {
        console.log('正在中断API请求')
        this.aborted = true
        this.source.cancel('用户取消了请求')
      },
      
      // 请求Promise
      requestPromise: null
    }
    
    // 创建请求
    controller.requestPromise = api.post('/chat/messages', {
      session_id: sessionId,
      message,
      models,
      image_url: imageUrl,
      stream: false // 明确指定非流式请求
    }, {
      cancelToken: controller.source.token
    }).then(response => {
      // 确保数据结构一致
      console.log('非流式消息响应:', response)
      return response
    }).catch(error => {
      if (axios.isCancel(error)) {
        console.log('请求被用户取消')
        return { aborted: true }
      }
      throw error
    })
    
    // 返回控制器
    return controller
  },

  // 流式发送消息
  sendMessageStream(sessionId, message, models, imageUrl = '', onChunk, onStart, onComplete, onError, onEnd) {
    // 创建一个可以中断的请求控制器
    const controller = {
      xhr: null,
      aborted: false,
      
      // 中断请求的方法
      abort() {
        if (this.xhr && this.xhr.readyState !== 4) {
          console.log('正在中断XHR请求')
          this.aborted = true
          this.xhr.abort()
          onError && onError({ error: '用户取消了请求' })
        }
      },
      
      // 请求Promise
      requestPromise: null
    }
    
    // 创建请求Promise
    controller.requestPromise = new Promise((resolve, reject) => {
      // 创建FormData
      const formData = new FormData()
      formData.append('session_id', sessionId)
      formData.append('message', message)
      formData.append('models', JSON.stringify(models))
      formData.append('image_url', imageUrl || '')
      formData.append('stream', 'true')
      
      // 创建XMLHttpRequest来处理流式响应
      const xhr = new XMLHttpRequest()
      
      // 保存xhr到控制器
      controller.xhr = xhr
      
      // 先打开连接
      xhr.open('POST', `${process.env.VUE_APP_API_BASE_URL || 'http://localhost:8080/api'}/chat/messages`)
      
      // 设置请求头
      xhr.setRequestHeader('Content-Type', 'application/json')
      
      // 获取token并设置认证头
      const token = store.state.token
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`)
      }
      
      let buffer = ''
      let currentEvent = ''  // 当前事件类型
      
      xhr.onprogress = function() {
        const newData = xhr.responseText.substring(buffer.length)
        buffer = xhr.responseText
        
        // 按行分割数据
        const lines = newData.split('\n')
        for (const line of lines) {
          // 忽略空行
          if (!line || line.trim() === '') continue
          
          // 检测事件行
          if (line.startsWith('event:')) {
            currentEvent = line.slice(6).trim() // 获取事件类型
            console.log('收到SSE事件:', currentEvent)
            continue
          }
          
          // 检测数据行
          if (line.startsWith('data:')) {
            try {
              const dataStr = line.slice(5).trim() // 去掉 'data:' 前缀
              console.log(`收到SSE数据(${currentEvent}):`, dataStr)
              
              // 解析JSON数据
              const data = JSON.parse(dataStr)
              
              // 根据事件类型处理
              switch (currentEvent) {
                case 'start':
                  console.log('流式响应开始:', data)
                  onStart && onStart(data)
                  break
                case 'chunk':
                  console.log('收到数据块:', data)
                  onChunk && onChunk(data)
                  break
                case 'complete':
                  console.log('响应完成:', data)
                  onComplete && onComplete(data)
                  break
                case 'error':
                  console.error('响应错误:', data)
                  onError && onError(data)
                  break
                case 'end':
                  console.log('流式响应结束:', data)
                  onEnd && onEnd(data)
                  resolve(data)
                  break
                default:
                  console.warn('未知事件类型:', currentEvent, data)
              }
            } catch (e) {
              console.error('解析SSE数据失败:', e, '原始数据:', line)
            }
          }
        }
      }
      
      xhr.onerror = function() {
        if (!controller.aborted) {
          reject(new Error('网络错误'))
        }
      }
      
      xhr.onload = function() {
        if (controller.aborted) {
          return
        }
        
        if (xhr.status !== 200) {
          reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`))
          return
        }
        
        // 如果响应结束但没有收到end事件，手动触发结束
        if (!buffer.includes('"event":"end"')) {
          console.log('响应结束但没有收到end事件，手动触发')
          onEnd && onEnd({ success: true })
          resolve({ success: true })
        }
      }
      
      // 请求中断处理
      xhr.onabort = function() {
        console.log('XHR请求已被中断')
        resolve({ aborted: true })
      }
      
      // 发送请求
      xhr.send(JSON.stringify({
        session_id: sessionId,
        message: message,
        models: models,
        image_url: imageUrl || '',
        stream: true
      }))
    })
    
    // 返回控制器对象，包含请求Promise和中断方法
    return controller
  },

  // 获取可用模型列表
  getAvailableModels() {
    return api.get('/chat/models')
  },

  // 删除会话
  deleteSession(sessionId) {
    return api.delete(`/chat/sessions/${sessionId}`)
  }
}

export default api 