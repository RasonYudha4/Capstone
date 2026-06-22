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

let refreshPromise: Promise<string> | null = null;

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
            const accessToken = localStorage.getItem('accessToken');
            
            if (refreshToken) {
                if (!refreshPromise) {
                    refreshPromise = axios.post(
                        `${API_BASE_URL}/auth/refresh`,
                        { refresh_token: refreshToken }
                    ).then((res) => {
                        const newAccessToken = res.data.data.access_token;
                        const newRefreshToken = res.data.data.refresh_token;
                        localStorage.setItem('accessToken', newAccessToken);
                        if (newRefreshToken) {
                            localStorage.setItem('refreshToken', newRefreshToken);
                        }
                        refreshPromise = null;
                        return newAccessToken;
                    }).catch((err) => {
                        refreshPromise = null;
                        localStorage.removeItem('accessToken');
                        localStorage.removeItem('refreshToken');
                        localStorage.removeItem('user');
                        throw err;
                    });
                }

                try {
                    const newAccessToken = await refreshPromise;
                    originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
                    return axioHandler(originalRequest);
                } catch (refreshError) {
                    // Only redirect if not already on login page
                    if (window.location.pathname !== '/login') {
                        window.location.href = '/login';
                    }
                    return Promise.reject(refreshError);
                }
            }

            if (!accessToken && !refreshToken) {
                return Promise.reject(error);
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