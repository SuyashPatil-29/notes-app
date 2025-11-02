import { createContext, useContext, useState, useEffect, type ReactNode } from "react";
import { useOrganization, useOrganizationList } from "@clerk/clerk-react";
import type { Organization } from "../types/backend";
import { getUserOrganizations } from "../utils/organization";

interface OrganizationContextType {
  // Active organization state
  activeOrg: Organization | null;
  setActiveOrg: (org: Organization | null) => void;
  isPersonalContext: boolean;
  
  // Organization list
  organizations: Organization[];
  isLoadingOrgs: boolean;
  refreshOrganizations: () => Promise<void>;
  
  // Clerk organization state
  clerkOrganization: ReturnType<typeof useOrganization>["organization"];
  clerkIsLoaded: boolean;
}

const OrganizationContext = createContext<OrganizationContextType | undefined>(undefined);

const STORAGE_KEY = "active_organization";

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const [activeOrg, setActiveOrgState] = useState<Organization | null>(null);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoadingOrgs, setIsLoadingOrgs] = useState(true);

  // Clerk hooks
  const { organization: clerkOrganization, isLoaded: clerkIsLoaded, membership } = useOrganization();
  const { setActive } = useOrganizationList();

  const isPersonalContext = activeOrg === null;

  // Load active org from localStorage on mount
  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && stored !== "null") {
      try {
        const parsed = JSON.parse(stored);
        setActiveOrgState(parsed);
      } catch (error) {
        console.error("Failed to parse stored organization:", error);
        localStorage.removeItem(STORAGE_KEY);
      }
    }
  }, []);

  // Fetch user's organizations
  const refreshOrganizations = async () => {
    try {
      setIsLoadingOrgs(true);
      const orgs = await getUserOrganizations();
      setOrganizations(orgs);
    } catch (error) {
      console.error("Failed to fetch organizations:", error);
    } finally {
      setIsLoadingOrgs(false);
    }
  };

  // Load organizations on mount
  useEffect(() => {
    refreshOrganizations();
  }, []);

  // Set active organization and sync with Clerk
  const setActiveOrg = async (org: Organization | null) => {
    setActiveOrgState(org);

    // Persist to localStorage
    if (org) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(org));
      // Sync with Clerk
      if (setActive) {
        try {
          await setActive({ organization: org.id });
        } catch (error) {
          console.error("Failed to set active organization in Clerk:", error);
        }
      }
    } else {
      localStorage.setItem(STORAGE_KEY, "null");
      // Clear Clerk organization
      if (setActive) {
        try {
          await setActive({ organization: null });
        } catch (error) {
          console.error("Failed to clear active organization in Clerk:", error);
        }
      }
    }
  };

  // Sync Clerk organization changes with our state and refresh role from backend
  useEffect(() => {
    if (!clerkIsLoaded) return;

    // If Clerk has an active organization but we don't, update our state
    if (clerkOrganization && !activeOrg) {
      // Determine user's role from membership
      const userRole = membership?.role === 'org:admin' ? 'admin' : 'member';
      
      const clerkOrgData: Organization = {
        id: clerkOrganization.id,
        name: clerkOrganization.name,
        slug: clerkOrganization.slug || "",
        imageUrl: clerkOrganization.imageUrl,
        createdAt: clerkOrganization.createdAt.toISOString(),
        membersCount: clerkOrganization.membersCount,
        role: userRole,
      };
      setActiveOrgState(clerkOrgData);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(clerkOrgData));
    }

    // If Clerk cleared organization but we still have one, clear ours
    if (!clerkOrganization && activeOrg) {
      setActiveOrgState(null);
      localStorage.setItem(STORAGE_KEY, "null");
    }

    // If activeOrg exists but role is missing, refresh from backend
    if (activeOrg && !activeOrg.role && organizations.length > 0) {
      const orgWithRole = organizations.find(o => o.id === activeOrg.id);
      if (orgWithRole && orgWithRole.role) {
        const updatedOrg = { ...activeOrg, role: orgWithRole.role };
        setActiveOrgState(updatedOrg);
        localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedOrg));
      }
    }
  }, [clerkOrganization, clerkIsLoaded, activeOrg, membership, organizations]);

  const value: OrganizationContextType = {
    activeOrg,
    setActiveOrg,
    isPersonalContext,
    organizations,
    isLoadingOrgs,
    refreshOrganizations,
    clerkOrganization,
    clerkIsLoaded,
  };

  return (
    <OrganizationContext.Provider value={value}>
      {children}
    </OrganizationContext.Provider>
  );
}

// Hook to use the organization context
export function useOrganizationContext() {
  const context = useContext(OrganizationContext);
  if (context === undefined) {
    throw new Error("useOrganizationContext must be used within OrganizationProvider");
  }
  return context;
}

