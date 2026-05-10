import { apiRequest } from '../request';
export const userApi = {
    getList: (params) => {
        return apiRequest({
            url: '/api/v1/users',
            method: 'GET',
            params,
        });
    },
    getById: (id) => {
        return apiRequest({
            url: `/api/v1/users/${id}`,
            method: 'GET',
        });
    },
    create: (data) => {
        return apiRequest({
            url: '/api/v1/users',
            method: 'POST',
            data,
        });
    },
    update: (id, data) => {
        return apiRequest({
            url: `/api/v1/users/${id}`,
            method: 'PUT',
            data,
        });
    },
    delete: (id) => {
        return apiRequest({
            url: `/api/v1/users/${id}`,
            method: 'DELETE',
        });
    },
    resetPassword: (id, newPassword) => {
        return apiRequest({
            url: `/api/v1/users/${id}/reset-password`,
            method: 'POST',
            data: { password: newPassword },
        });
    },
};
