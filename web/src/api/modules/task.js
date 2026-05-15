import { apiRequest } from '../request';
export const taskApi = {
    getMyTasks: (params) => {
        return apiRequest({
            url: '/v1/tasks/my',
            method: 'GET',
            params,
        });
    },
    getById: (id) => {
        return apiRequest({
            url: `/v1/tasks/${id}`,
            method: 'GET',
        });
    },
    executeAction: (id, action) => {
        return apiRequest({
            url: `/v1/tasks/${id}/action`,
            method: 'POST',
            data: action,
        });
    },
    assign: (taskId, userId) => {
        return apiRequest({
            url: `/v1/tasks/${taskId}/assign`,
            method: 'POST',
            data: { user_id: userId },
        });
    },
};
