import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

export const axioHandler: AxiosInstance = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json'
    },
});

axioHandler.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('accessToken');

        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

axioHandler.interceptors.response.use(
    (response: AxiosResponse) => {
        return response;
    },

    async (error) => {
        const originalRequest = error.config;

        // Don't intercept auth endpoints to prevent loops
        const isAuthEndpoint = originalRequest?.url?.startsWith('/auth/');

        if (
            error.response?.status === 401 &&
            !isAuthEndpoint &&
            !originalRequest._retry
        ) {
            originalRequest._retry = true;

            const refreshToken = localStorage.getItem('refreshToken');
            if (refreshToken) {
                try {
                    const { data } = await axios.post(
                        `${API_BASE_URL}/auth/refresh`,
                        { refresh_token: refreshToken }
                    );
                    const newAccessToken = data.data.access_token;
                    localStorage.setItem('accessToken', newAccessToken);
                    originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
                    return axioHandler(originalRequest);
                } catch {
                    // Refresh also failed — clear storage and redirect
                    localStorage.removeItem('accessToken');
                    localStorage.removeItem('refreshToken');
                    localStorage.removeItem('user');
                }
            }

            // Only redirect if not already on login page
            if (window.location.pathname !== '/login') {
                window.location.href = '/login';
            }
        }

        return Promise.reject(error);
    }
);

export default axioHandler;