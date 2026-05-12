import { apiRequest } from '../request';
import axios from 'axios';

const getBaseUrl = () => {
  const urls = ['http://localhost:8080', 'http://host.docker.internal:8080'];
  return urls[0];
};

export const authApi = {
    login: (credentials) => {
        const base = getBaseUrl();
        return axios.post(`${base}/api/v1/auth/login`, credentials).then(r => r.data.data);
    },
    logout: () => {
        return apiRequest({
            url: '/api/v1/auth/logout',
            method: 'POST',
        });
    },
    getCurrentUser: () => {
        return apiRequest({
            url: '/api/v1/auth/me',
            method: 'GET',
        });
    },
    refreshToken: () => {
        return apiRequest({
            url: '/api/v1/auth/refresh',
            method: 'POST',
        });
    },

    devUsers: () => {
        const base = getBaseUrl();
        return axios.get(`${base}/api/v1/auth/dev-users`).then(r => r.data.data);
    },
};
