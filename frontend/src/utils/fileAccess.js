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
 * @returns {string} 音频访问URL
 */
export async function getAudioUrl(path) {
    if (!path) return ''
    try {
        const token = localStorage.getItem('token')
        const response = await fetch(path, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        if (!response.ok) {
            throw new Error(`音频加载失败: ${response.statusText}`)
        }
        const blob = await response.blob()
        return window.URL.createObjectURL(blob)
    } catch (error) {
        console.error('音频加载失败:', error)
        throw error
    }
}

/**
 * 获取视频URL
 * @param {string} path - 视频文件路径
 * @returns {string} 视频访问URL
 */
export async function getVideoUrl(path) {
    if (!path) return ''
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
 * @returns {string} 图片访问URL（通过请求头认证获取）
 */
export async function getImageUrl(path) {
    if (!path) return ''
    try {
        const token = localStorage.getItem('token')
        // 使用axios发送带有Authorization头的请求获取图片
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