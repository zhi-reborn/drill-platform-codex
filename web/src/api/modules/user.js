import { apiRequest } from '../request';
export const userApi = {
    getList: (params) => {
        return apiRequest({
            url: '/v1/users',
            method: 'GET',
            params,
        });
    },
    getById: (id) => {
        return apiRequest({
            url: `/v1/users/${id}`,
            method: 'GET',
        });
    },
    create: (data) => {
        return apiRequest({
            url: '/v1/users',
            method: 'POST',
            data: {
                ...data,
                real_name: data.name,
            },
        });
    },
    update: (id, data) => {
        return apiRequest({
            url: `/v1/users/${id}`,
            method: 'PUT',
            data,
        });
    },
    delete: (id) => {
        return apiRequest({
            url: `/v1/users/${id}`,
            method: 'DELETE',
        });
    },
    resetPassword: (id, newPassword) => {
        return apiRequest({
            url: `/v1/users/${id}/reset-password`,
            method: 'POST',
            data: { password: newPassword },
        });
    },
};
