import { apiRequest } from '../request';
export const notificationApi = {
    getList: (params) => {
        return apiRequest({
            url: '/api/v1/notifications',
            method: 'GET',
            params,
        });
    },
    markAsRead: (id) => {
        return apiRequest({
            url: `/api/v1/notifications/${id}/read`,
            method: 'POST',
        });
    },
    markAllAsRead: () => {
        return apiRequest({
            url: '/api/v1/notifications/read-all',
            method: 'POST',
        });
    },
    delete: (id) => {
        return apiRequest({
            url: `/api/v1/notifications/${id}`,
            method: 'DELETE',
        });
    },
};
