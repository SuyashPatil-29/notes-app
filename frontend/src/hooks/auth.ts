import { useCallback, useEffect, useState } from "react";
import type { AuthenticatedUser } from "@/types/backend";
import { getCurrentUser } from "@/utils/auth";

export const useUser = () => {
    const [user, setUser] = useState<AuthenticatedUser | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchUser = useCallback(async () => {
        setLoading(true);
            const currentUser = await getCurrentUser();
            setUser(currentUser);
            setLoading(false);
    }, []);

    useEffect(() => {
        fetchUser();
    }, [fetchUser]);

    return { user, loading, refetch: fetchUser };
};