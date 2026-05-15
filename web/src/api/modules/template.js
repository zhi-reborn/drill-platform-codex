import { apiRequest } from '../request';
export const templateApi = {
    getList: (params) => {
        return apiRequest({
            url: '/v1/templates',
            method: 'GET',
            params,
        });
    },
    getById: (id) => {
        return apiRequest({
            url: `/v1/templates/${id}`,
            method: 'GET',
        });
    },
    create: (data) => {
        return apiRequest({
            url: '/v1/templates',
            method: 'POST',
            data,
        });
    },
    update: (id, data) => {
        return apiRequest({
            url: `/v1/templates/${id}`,
            method: 'PUT',
            data,
        });
    },
    delete: (id) => {
        return apiRequest({
            url: `/v1/templates/${id}`,
            method: 'DELETE',
        });
    },
    publish: (id) => {
        return apiRequest({
            url: `/v1/templates/${id}/publish`,
            method: 'POST',
        });
    },
};
