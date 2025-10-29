import { useEffect, useState } from "react";
import type { AuthenticatedUser } from "@/types/backend";
import { getCurrentUser } from "@/utils/auth";

export const useUser = () => {
    const [user, setUser] = useState<AuthenticatedUser | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchUser = async () => {
            const currentUser = await getCurrentUser();
            setUser(currentUser);
            setLoading(false);
        };

        fetchUser();
    }, []);

    return { user, loading };
};