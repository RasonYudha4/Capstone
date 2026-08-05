import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

/** Auth routes that must not trigger silent token refresh (login flow / refresh itself). */
const NO_REFRESH_PATHS = [
    '/auth/login',
    '/auth/verify-otp',
    '/auth/resend-otp',
    '/auth/refresh',
    '/auth/logout',
    '/auth/complete-invitation',
    '/auth/forgot-password',
    '/auth/reset-password',
];

function shouldSkipRefresh(url?: string): boolean {
    if (!url) return true;
    return NO_REFRESH_PATHS.some((path) => url === path || url.startsWith(`${path}?`));
}

export const axioHandler: AxiosInstance = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json'
    },
    // Send HttpOnly auth cookies on cross-origin API calls.
    withCredentials: true,
});

let refreshPromise: Promise<void> | null = null;

function clearLegacyTokenStorage() {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
}

axioHandler.interceptors.response.use(
    (response: AxiosResponse) => {
        return response;
    },

    async (error) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        if (
            error.response?.status === 401 &&
            originalRequest &&
            !originalRequest._retry &&
            !shouldSkipRefresh(originalRequest.url)
        ) {
            originalRequest._retry = true;

            if (!refreshPromise) {
                refreshPromise = axios
                    .post(
                        `${API_BASE_URL}/auth/refresh`,
                        {},
                        { withCredentials: true },
                    )
                    .then(() => {
                        // New tokens arrive only as Set-Cookie (HttpOnly).
                        refreshPromise = null;
                    })
                    .catch((err) => {
                        refreshPromise = null;
                        clearLegacyTokenStorage();
                        throw err;
                    });
            }

            try {
                await refreshPromise;
                return axioHandler(originalRequest);
            } catch (refreshError) {
                // /auth/me is the session probe — let AuthContext handle guest state without hard redirect.
                const isSessionProbe =
                    originalRequest.url === '/auth/me' ||
                    originalRequest.url?.startsWith('/auth/me?');
                if (!isSessionProbe && window.location.pathname !== '/login') {
                    window.location.href = '/login';
                }
                return Promise.reject(refreshError);
            }
        }

        return Promise.reject(error);
    }
);

export default axioHandler;
