// 文件访问工具

/**
 * 获取文件访问URL
 * @param {string} path - 文件路径
 * @returns {string} 文件访问URL
 */
export function getFileAccessUrl(path) {
    if (!path) return ''
    return path
}

/**
 * 创建带认证信息的URL
 * @param {string} path - 文件路径
 * @returns {string} 带认证信息的URL
 */
export function createAuthUrl(path) {
    if (!path) return ''
    
    // 如果已经是完整URL或者是blob URL，直接返回
    if (path.startsWith('blob:') || path.startsWith('data:')) {
        return path
    }
    
    const token = localStorage.getItem('token')
    // 添加随机参数，避免缓存问题
    const urlWithCache = path.includes('?') 
        ? `${path}&_t=${Date.now()}&token=${token}` 
        : `${path}?_t=${Date.now()}&token=${token}`
    
    return urlWithCache
}

/**
 * 下载文件
 * @param {string} path - 文件路径
 * @param {string} filename - 下载时的文件名
 */
export async function downloadFile(path, filename) {
    if (!path) return
    try {
        // 清理URL，处理可能的URL拼接错误
        // 检查是否包含blob:，如果包含但不是以blob:开头，则提取blob:部分
        const blobIndex = path.indexOf('blob:')
        if (blobIndex > 0) {
            path = path.substring(blobIndex)
        }

        // 检查是否为blob URL，如果是则直接使用，不需要再次fetch
        if (path.startsWith('blob:')) {
            const link = document.createElement('a')
            link.href = path
            link.download = filename || ''
            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
            return
        }

        // 确保URL格式正确
        if (!path.startsWith('http://') && !path.startsWith('https://') && !path.startsWith('/')) {
            throw new Error(`无效的URL格式: ${path}`)
        }

        const token = localStorage.getItem('token')
        const response = await fetch(path, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })

        if (!response.ok) {
            throw new Error(`下载失败: ${response.statusText}`)
        }

        const blob = await response.blob()
        const downloadUrl = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = filename || ''
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(downloadUrl)
    } catch (error) {
        console.error('文件下载失败:', error)
        throw error
    }
}

/**
 * 获取音频URL
 * @param {string} path - 音频文件路径
 * @param {Object} options - 可选配置
 * @param {number} options.timeout - 超时时间(毫秒)
 * @param {number} options.retries - 重试次数
 * @param {boolean} options.useBlob - 是否使用blob URL (默认为false)
 * @returns {string} 音频访问URL
 */
export async function getAudioUrl(path, options = {}) {
    if (!path) return ''
    
    // 如果不需要使用blob (默认)，直接返回带认证的URL
    if (options.useBlob !== true) {
        return createAuthUrl(path)
    }
    
    // 以下是原来的blob URL生成逻辑
    const timeout = options.timeout || 15000 // 默认15秒超时
    const maxRetries = options.retries || 2 // 默认重试2次
    let retries = 0
    
    const fetchWithTimeout = (url, options, timeout) => {
        return Promise.race([
            fetch(url, options),
            new Promise((_, reject) => 
                setTimeout(() => reject(new Error('请求超时')), timeout)
            )
        ])
    }
    
    async function attemptFetch() {
        try {
            const token = localStorage.getItem('token')
            
            // 添加随机参数，避免缓存问题
            const urlWithCache = path.includes('?') 
                ? `${path}&_t=${Date.now()}` 
                : `${path}?_t=${Date.now()}`
                
            const response = await fetchWithTimeout(urlWithCache, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Cache-Control': 'no-cache'
                }
            }, timeout)
            
            if (!response.ok) {
                throw new Error(`音频加载失败: ${response.statusText}`)
            }
            
            const blob = await response.blob()
            return window.URL.createObjectURL(blob)
        } catch (error) {
            // 如果还有重试次数，则重试
            if (retries < maxRetries) {
                retries++
                console.log(`音频加载失败，正在进行第 ${retries} 次重试...`)
                // 指数退避策略，每次重试等待时间增加
                await new Promise(resolve => setTimeout(resolve, 1000 * retries))
                return attemptFetch()
            }
            console.error('音频加载失败:', error)
            throw error
        }
    }
    
    return attemptFetch()
}

/**
 * 获取视频URL
 * @param {string} path - 视频文件路径
 * @param {Object} options - 可选配置
 * @param {boolean} options.useBlob - 是否使用blob URL (默认为false)
 * @returns {string} 视频访问URL
 */
export async function getVideoUrl(path, options = {}) {
    if (!path) return ''
    
    // 如果不需要使用blob (默认)，直接返回带认证的URL
    if (options.useBlob !== true) {
        return createAuthUrl(path)
    }
    
    // 以下是原来的blob URL生成逻辑
    try {
        const token = localStorage.getItem('token')
        const response = await fetch(path, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        if (!response.ok) {
            throw new Error(`视频加载失败: ${response.statusText}`)
        }
        const blob = await response.blob()
        return window.URL.createObjectURL(blob)
    } catch (error) {
        console.error('视频加载失败:', error)
        throw error
    }
}

/**
 * 获取图片URL
 * @param {string} path - 图片文件路径
 * @param {Object} options - 可选配置
 * @param {boolean} options.useBlob - 是否使用blob URL (默认为false)
 * @returns {string} 图片访问URL
 */
export async function getImageUrl(path, options = {}) {
    if (!path) return ''
    
    // 如果不需要使用blob (默认)，直接返回带认证的URL
    if (options.useBlob !== true) {
        return createAuthUrl(path)
    }
    
    // 以下是原来的blob URL生成逻辑
    try {
        const token = localStorage.getItem('token')
        const response = await fetch(path, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })

        if (!response.ok) {
            throw new Error(`图片加载失败: ${response.statusText}`)
        }

        const blob = await response.blob()
        return window.URL.createObjectURL(blob)
    } catch (error) {
        console.error('图片URL处理失败:', error)
        throw error
    }
}

/**
 * 获取直接访问文件的URL（无Blob转换，仅添加token）
 * @param {string} path - 文件路径
 * @returns {string} 附带token的URL
 */
export function getDirectFileUrl(path) {
    if (!path) return ''
    
    // 如果是HTTP/HTTPS完整URL，不需要特殊处理，直接添加token
    if (path.startsWith('http://') || path.startsWith('https://')) {
        const token = localStorage.getItem('token')
        const separator = path.includes('?') ? '&' : '?'
        return `${path}${separator}token=${token}`
    }
    
    // 如果是相对路径，转为绝对路径
    const baseURL = process.env.VUE_APP_API_URL || (window.location.origin + '/api')
    if (path.startsWith('/')) {
        path = path.substring(1) // 移除前导斜杠
    }
    
    const fullUrl = `${baseURL}/${path}`
    const token = localStorage.getItem('token')
    return `${fullUrl}?token=${token}`
}