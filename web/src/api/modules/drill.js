import { apiRequest } from '../request';
export const drillApi = {
    getList: (params) => {
        return apiRequest({
            url: '/api/v1/drills',
            method: 'GET',
            params,
        });
    },
    getById: (id) => {
        return apiRequest({
            url: `/api/v1/drills/${id}`,
            method: 'GET',
        });
    },
    create: (data) => {
        return apiRequest({
            url: '/api/v1/drills',
            method: 'POST',
            data,
        });
    },
    start: (id) => {
        return apiRequest({
            url: `/api/v1/drills/${id}/start`,
            method: 'POST',
        });
    },
    pause: (id) => {
        return apiRequest({
            url: `/api/v1/drills/${id}/pause`,
            method: 'POST',
        });
    },
    resume: (id) => {
        return apiRequest({
            url: `/api/v1/drills/${id}/resume`,
            method: 'POST',
        });
    },
    terminate: (id) => {
        return apiRequest({
            url: `/api/v1/drills/${id}/terminate`,
            method: 'POST',
        });
    },
    getSteps: (drillId) => {
        return apiRequest({
            url: `/api/v1/drills/${drillId}/steps`,
            method: 'GET',
        });
    },
    getStepLogs: (stepId) => {
        return apiRequest({
            url: `/api/v1/steps/${stepId}/logs`,
            method: 'GET',
        });
    },
};
