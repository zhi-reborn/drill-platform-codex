import { apiRequest } from '../request';
export const configApi = {
    getSystemConfig: () => {
        return apiRequest({
            url: '/v1/config/system',
            method: 'GET',
        });
    },
    updateSystemConfig: (data) => {
        return apiRequest({
            url: '/v1/config/system',
            method: 'PUT',
            data,
        });
    },
};
