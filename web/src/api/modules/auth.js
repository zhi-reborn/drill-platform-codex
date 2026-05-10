import { apiRequest } from '../request';
export const authApi = {
    login: (credentials) => {
        return apiRequest({
            url: '/api/v1/auth/login',
            method: 'POST',
            data: credentials,
        });
    },
    logout: () => {
        return apiRequest({
            url: '/api/v1/auth/logout',
            method: 'POST',
        });
    },
    getCurrentUser: () => {
        return apiRequest({
            url: '/api/v1/auth/current',
            method: 'GET',
        });
    },
    refreshToken: () => {
        return apiRequest({
            url: '/api/v1/auth/refresh',
            method: 'POST',
        });
    },
};
