import { useUser as useClerkUser } from "@clerk/clerk-react";
import type { AuthenticatedUser } from "@/types/backend";
import { useEffect, useState, useRef } from "react";
import api from "@/utils/api";

export const useUser = () => {
    const { user: clerkUser, isLoaded } = useClerkUser();
    const [user, setUser] = useState<AuthenticatedUser | null>(null);
    const [loading, setLoading] = useState(true);
    const fetchedRef = useRef(false);
    const clerkUserIdRef = useRef<string | null>(null);

    useEffect(() => {
        const fetchBackendUser = async () => {
            // Skip if already fetched for this user
            if (clerkUser && clerkUserIdRef.current === clerkUser.id && fetchedRef.current) {
                return;
            }

            if (isLoaded && clerkUser) {
                setLoading(true);
                try {
                    // Fetch user from our backend to get additional data
                    const response = await api.get('/auth/user');
                    setUser(response.data);
                    fetchedRef.current = true;
                    clerkUserIdRef.current = clerkUser.id;
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
                    fetchedRef.current = true;
                    clerkUserIdRef.current = clerkUser.id;
                }
                setLoading(false);
            } else if (isLoaded && !clerkUser) {
                setUser(null);
                setLoading(false);
                fetchedRef.current = false;
                clerkUserIdRef.current = null;
            }
        };

        fetchBackendUser();
    }, [clerkUser?.id, isLoaded]);

    return { 
        user, 
        loading: !isLoaded || loading, 
        refetch: async () => {
            if (clerkUser) {
                try {
                    const response = await api.get('/auth/user');
                    setUser(response.data);
                    fetchedRef.current = true;
                } catch (error) {
                    console.error('Error refetching user:', error);
                }
            }
        }
    };
};
