import { useUser as useClerkUser } from "@clerk/clerk-react";
import type { AuthenticatedUser } from "@/types/backend";
import { useEffect, useState } from "react";
import api from "@/utils/api";

export const useUser = () => {
    const { user: clerkUser, isLoaded } = useClerkUser();
    const [user, setUser] = useState<AuthenticatedUser | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchBackendUser = async () => {
            if (isLoaded && clerkUser) {
                setLoading(true);
                try {
                    // Fetch user from our backend to get additional data
                    const response = await api.get('/auth/user');
                    setUser(response.data);
                } catch (error) {
                    console.error('Error fetching user from backend:', error);
                    // If backend user fetch fails, create a minimal user from Clerk data
                    setUser({
                        id: 0, // Backend will assign proper ID
                        clerkUserId: clerkUser.id,
                        name: clerkUser.fullName || clerkUser.firstName || 'User',
                        email: clerkUser.primaryEmailAddress?.emailAddress || '',
                        imageUrl: clerkUser.imageUrl || null,
                        onboardingCompleted: false,
                        hasApiKey: false,
                    });
                }
                setLoading(false);
            } else if (isLoaded && !clerkUser) {
                setUser(null);
                setLoading(false);
            }
        };

        fetchBackendUser();
    }, [clerkUser, isLoaded]);

    return { 
        user, 
        loading: !isLoaded || loading, 
        refetch: async () => {
            if (clerkUser) {
                try {
                    const response = await api.get('/auth/user');
                    setUser(response.data);
                } catch (error) {
                    console.error('Error refetching user:', error);
                }
            }
        }
    };
};
