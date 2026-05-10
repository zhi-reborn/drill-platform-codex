import { apiRequest } from '../request';
export const displayApi = {
    getDashboard: () => {
        return apiRequest({
            url: '/api/v1/display/dashboard',
            method: 'GET',
        });
    },
    getActiveDrills: () => {
        return apiRequest({
            url: '/api/v1/display/active-drills',
            method: 'GET',
        });
    },
    getDrillDetail: (id) => {
        return apiRequest({
            url: `/api/v1/display/drills/${id}`,
            method: 'GET',
        });
    },
};
