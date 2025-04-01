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