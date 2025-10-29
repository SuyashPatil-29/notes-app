import type { AuthenticatedUser } from '@/types/backend';
import api from '@/utils/api';

const API_BASE_URL = 'http://localhost:8080';

export const handleGoogleLogin = () => {
    window.location.href = `${API_BASE_URL}/auth/google`;
};

export const handleGoogleLogout = () => {
    window.location.href = `${API_BASE_URL}/logout/google`;
};

export const getCurrentUser = async (): Promise<AuthenticatedUser | null> => {
    try {
        const response = await api.get('auth/user');

        console.log(response);

        if (response.status === 401) {
            // User is not logged in
            console.log('User is not logged in');
            throw new Error('Failed to fetch user');
        }

        const user: AuthenticatedUser = response.data;
        return user;

    } catch (error) {
        console.error('Error fetching current user:', error);
        return null;
    }
};