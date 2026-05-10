import { apiRequest } from '../request';
export const configApi = {
    getSystemConfig: () => {
        return apiRequest({
            url: '/api/v1/config/system',
            method: 'GET',
        });
    },
    updateSystemConfig: (data) => {
        return apiRequest({
            url: '/api/v1/config/system',
            method: 'PUT',
            data,
        });
    },
};
