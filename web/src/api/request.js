import axios from 'axios';
import { ElMessage } from 'element-plus';
const request = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL,
    timeout: 15000,
});
request.interceptors.request.use((config) => {
    const auth = localStorage.getItem('drill_auth');
    if (auth) {
        const { access_token } = JSON.parse(auth);
        config.headers.Authorization = `Bearer ${access_token}`;
    }
    return config;
}, (error) => Promise.reject(error));
request.interceptors.response.use((response) => response, (error) => {
    const status = error.response?.status;
    switch (status) {
        case 401:
            localStorage.removeItem('drill_auth');
            window.location.href = '/login';
            break;
        case 403:
            ElMessage.error('没有权限执行此操作');
            break;
        case 500:
            ElMessage.error('服务器错误');
            break;
        default:
            ElMessage.error(error.response?.data?.message || '请求失败');
    }
    return Promise.reject(error);
});
export function apiRequest(config) {
    return request(config);
}
