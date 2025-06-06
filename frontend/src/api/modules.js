import axios from 'axios'

// 获取图像处理模块列表
export const getImageProcessingModules = () => {
    return axios.get('/api/image-processing/modules')
        .then(response => {
            if (response.data && response.data.modules) {
                // 处理模块数据，添加完整的路由路径
                const modules = response.data.modules.map(module => ({
                    ...module,
                    route: `/image-processing${module.route}`
                }))
                return {
                    success: true,
                    modules
                }
            }
            return {
                success: false,
                error: '获取模块列表失败'
            }
        })
        .catch(error => ({
            success: false,
            error: error.message || '获取模块列表失败'
        }))
}

// 获取模块详细信息
export const getModuleDetail = (moduleId) => {
    return axios.get(`/api/image-processing/modules/${moduleId}`)
        .then(response => ({
            success: true,
            module: response.data
        }))
        .catch(error => ({
            success: false,
            error: error.message || '获取模块详情失败'
        }))
}