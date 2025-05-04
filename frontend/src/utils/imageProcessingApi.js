// 图像处理API服务
import axios from 'axios'

/**
 * 获取图像处理模块列表
 * @returns {Promise} 图像处理模块列表
 */
export async function getImageProcessingModules() {
    try {
        const response = await axios.get('/api/image-processing/modules')
        return response.data
    } catch (error) {
        console.error('获取图像处理模块列表失败:', error)
        throw error
    }
}

/**
 * 获取可用的工作流类型
 * @returns {Promise} 工作流类型列表
 */
export async function getWorkflowTypes() {
    try {
        const response = await axios.get('/api/image-processing/workflow-types')
        return response.data
    } catch (error) {
        console.error('获取工作流类型失败:', error)
        throw error
    }
}

/**
 * 创建图像处理任务
 * @param {string} moduleId - 模块ID
 * @param {Object} params - 任务参数
 * @returns {Promise} 创建的任务信息
 */
export async function createImageProcessingTask(moduleId, formData) {
    try {
        const response = await axios.post(`/api/image-processing/tasks/${moduleId}`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })
        return response.data
    } catch (error) {
        console.error('创建图像处理任务失败:', error)
        throw error
    }
}

/**
 * 获取图像处理任务列表
 * @param {string} moduleId - 模块ID
 * @returns {Promise} 任务列表
 */
export async function getImageProcessingTasks(moduleId) {
    try {
        const response = await axios.get(`/api/image-processing/tasks/${moduleId}`)
        return response.data
    } catch (error) {
        console.error('获取图像处理任务列表失败:', error)
        throw error
    }
}

/**
 * 获取图像处理任务详情
 * @param {string} moduleId - 模块ID
 * @param {number|string} taskId - 任务ID
 * @returns {Promise} 任务详情
 */
export async function getImageProcessingTaskDetail(moduleId, taskId) {
    try {
        const response = await axios.get(`/api/image-processing/tasks/${moduleId}/${taskId}`)
        return response.data
    } catch (error) {
        console.error('获取图像处理任务详情失败:', error)
        throw error
    }
}

/**
 * 删除图像处理任务
 * @param {string} moduleId - 模块ID
 * @param {number|string} taskId - 任务ID
 * @returns {Promise} 删除结果
 */
export async function deleteImageProcessingTask(moduleId, taskId) {
    try {
        const response = await axios.delete(`/api/image-processing/tasks/${moduleId}/${taskId}`)
        return response.data
    } catch (error) {
        console.error('删除图像处理任务失败:', error)
        throw error
    }
}

/**
 * 重试图像处理任务
 * @param {string} moduleId - 模块ID
 * @param {number|string} taskId - 任务ID
 * @returns {Promise} 重试结果
 */
export async function retryImageProcessingTask(moduleId, taskId) {
    try {
        const response = await axios.post(`/api/image-processing/tasks/${moduleId}/${taskId}/retry`)
        return response.data
    } catch (error) {
        console.error('重试图像处理任务失败:', error)
        throw error
    }
}