import { apiRequest } from '../request';
export const authApi = {
    login: (credentials) => {
        return apiRequest({
            url: '/v1/auth/login',
            method: 'POST',
            data: credentials,
        });
    },
    logout: () => {
        return apiRequest({
            url: '/v1/auth/logout',
            method: 'POST',
        });
    },
    getCurrentUser: () => {
        return apiRequest({
            url: '/v1/auth/me',
            method: 'GET',
        });
    },
    refreshToken: () => {
        return apiRequest({
            url: '/v1/auth/refresh',
            method: 'POST',
        });
    },
    devUsers: () => {
        return apiRequest({
            url: '/v1/auth/dev-users',
            method: 'GET',
        });
    },
};
