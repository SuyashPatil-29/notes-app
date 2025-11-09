import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";

const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080",
    headers: {
        "Content-Type": "application/json",
    },
});

// This will be set by the App component when Clerk is initialized
let getAuthToken: (() => Promise<string | null>) | null = null;

export function setAuthTokenGetter(getter: () => Promise<string | null>) {
    getAuthToken = getter;
}

export function getStoredAuthToken(): Promise<string | null> {
    if (getAuthToken) {
        return getAuthToken();
    }
    return Promise.resolve(null);
}

// Track if we're currently refreshing to avoid multiple simultaneous refreshes
let isRefreshing = false;
let failedQueue: Array<{
    resolve: (value?: any) => void;
    reject: (reason?: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(token);
        }
    });

    failedQueue = [];
};

// Add request interceptor to include Clerk JWT token
api.interceptors.request.use(
    async (config) => {
        if (getAuthToken) {
            try {
                const token = await getAuthToken();
                if (token) {
                    config.headers.Authorization = `Bearer ${token}`;
                }
            } catch (error) {
                console.error('[API] Error getting auth token:', error);
            }
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Add response interceptor to handle 401 errors (token expiration)
api.interceptors.response.use(
    (response) => {
        return response;
    },
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

        // If we get a 401 and haven't already retried this request
        if (error.response?.status === 401 && !originalRequest._retry && getAuthToken) {
            if (isRefreshing) {
                // If already refreshing, queue this request
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                })
                    .then((token) => {
                        if (originalRequest.headers && token) {
                            originalRequest.headers.Authorization = `Bearer ${token}`;
                        }
                        return api(originalRequest);
                    })
                    .catch((err) => {
                        return Promise.reject(err);
                    });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                console.log('[API] Token expired, refreshing...');
                // Force token refresh by passing { skipCache: true }
                const newToken = await getAuthToken();

                if (newToken) {
                    console.log('[API] Token refreshed successfully');
                    if (originalRequest.headers) {
                        originalRequest.headers.Authorization = `Bearer ${newToken}`;
                    }
                    processQueue(null, newToken);
                    return api(originalRequest);
                } else {
                    console.error('[API] Failed to get new token');
                    processQueue(new Error('Failed to refresh token'), null);
                    return Promise.reject(error);
                }
            } catch (refreshError) {
                console.error('[API] Error refreshing token:', refreshError);
                processQueue(refreshError, null);
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
        }

        return Promise.reject(error);
    }
);

export default api;
