import { useState } from "react";
import { useOrganizationList } from "@clerk/clerk-react";
import { useNavigate } from "react-router-dom";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Building2, ChevronDown, Plus, Users, Check } from "lucide-react";
import { useOrganizationContext } from "@/contexts/OrganizationContext";
import { CreateOrganizationDialog } from "./CreateOrganizationDialog";

export function OrganizationSwitcher() {
  const { activeOrg, setActiveOrg } = useOrganizationContext();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const navigate = useNavigate();

  const { isLoaded, userMemberships, setActive } = useOrganizationList({
    userMemberships: {
      infinite: true,
    },
  });

  const handleSelectOrganization = async (orgId: string) => {
    if (!setActive) return;
    try {
      const org = userMemberships.data?.find(
        (m) => m.organization.id === orgId
      )?.organization;
      if (org) {
        const orgData = {
          id: org.id,
          name: org.name,
          slug: org.slug || "",
          imageUrl: org.imageUrl,
          createdAt: org.createdAt.toISOString(),
          membersCount: org.membersCount,
        };
        await setActive({ organization: orgId });
        setActiveOrg(orgData);
        // Navigate to Dashboard after switching organizations
        navigate("/");
      }
    } catch (error) {
      console.error("Failed to switch organization:", error);
    }
  };

  const handleSelectPersonal = async () => {
    if (!setActive) return;
    try {
      await setActive({ organization: null });
      setActiveOrg(null);
      // Navigate to Dashboard after switching to personal
      navigate("/");
    } catch (error) {
      console.error("Failed to switch to personal:", error);
    }
  };

  if (!isLoaded) {
    return <Skeleton className="h-9 w-[200px]" />;
  }

  const orgName = activeOrg?.name || "Personal";

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            className="w-[200px] justify-between"
            aria-label="Switch organization"
          >
            <div className="flex items-center gap-2">
              <Building2 className="h-4 w-4" />
              <span className="truncate">{orgName}</span>
            </div>
            <ChevronDown className="h-4 w-4 opacity-50" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-[250px]">
          <DropdownMenuLabel>Organizations</DropdownMenuLabel>
          <DropdownMenuSeparator />

          {/* Personal Account */}
          <DropdownMenuItem
            onClick={handleSelectPersonal}
            className="cursor-pointer"
          >
            <div className="flex items-center justify-between w-full">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                <span>Personal</span>
              </div>
              {!activeOrg && <Check className="h-4 w-4" />}
            </div>
          </DropdownMenuItem>

          <DropdownMenuSeparator />

          {/* Organization List */}
          {userMemberships.data?.map((membership) => (
            <DropdownMenuItem
              key={membership.id}
              onClick={() => handleSelectOrganization(membership.organization.id)}
              className="cursor-pointer"
            >
              <div className="flex items-center justify-between w-full">
                <div className="flex items-center gap-2">
                  <Building2 className="h-4 w-4" />
                  <span className="truncate">
                    {membership.organization.name}
                  </span>
                </div>
                {activeOrg?.id === membership.organization.id && (
                  <Check className="h-4 w-4" />
                )}
              </div>
            </DropdownMenuItem>
          ))}

          {userMemberships.hasNextPage && (
            <DropdownMenuItem
              onClick={() => userMemberships.fetchNext()}
              disabled={userMemberships.isFetching}
              className="cursor-pointer"
            >
              Load more...
            </DropdownMenuItem>
          )}

          <DropdownMenuSeparator />

          {/* Create Organization */}
          <DropdownMenuItem
            onClick={() => setIsCreateDialogOpen(true)}
            className="cursor-pointer"
          >
            <div className="flex items-center gap-2">
              <Plus className="h-4 w-4" />
              <span>Create Organization</span>
            </div>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <CreateOrganizationDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
      />
    </>
  );
}

