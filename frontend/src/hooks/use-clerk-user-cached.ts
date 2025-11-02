import { useUser as useClerkUser } from "@clerk/clerk-react";
import { useMemo } from "react";

/**
 * Optimized Clerk user hook that memoizes user data to prevent unnecessary re-renders
 * Use this instead of useClerkUser directly when you only need basic user info
 */
export const useClerkUserCached = () => {
    const { user, isLoaded, isSignedIn } = useClerkUser();

    // Memoize user data to prevent unnecessary re-renders
    const cachedUser = useMemo(() => {
        if (!user) return null;
        
        return {
            id: user.id,
            fullName: user.fullName,
            firstName: user.firstName,
            lastName: user.lastName,
            username: user.username,
            imageUrl: user.imageUrl,
            primaryEmailAddress: user.primaryEmailAddress?.emailAddress,
            publicMetadata: user.publicMetadata,
        };
    }, [
        user?.id,
        user?.fullName,
        user?.firstName,
        user?.lastName,
        user?.username,
        user?.imageUrl,
        user?.primaryEmailAddress?.emailAddress,
        // Only re-memoize if publicMetadata changes (shallow comparison)
        JSON.stringify(user?.publicMetadata),
    ]);

    return {
        user: cachedUser,
        isLoaded,
        isSignedIn,
    };
};
