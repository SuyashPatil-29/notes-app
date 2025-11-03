import { useState, useEffect } from "react";
import { useOrganizationList } from "@clerk/clerk-react";
import { useOrganizationContext } from "@/contexts/OrganizationContext";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Building2, Check, X, Loader2, Mail, UserCheck } from "lucide-react";
import { toast } from "sonner";

export function OrganizationsMemberView() {
  const { isLoaded, setActive, userInvitations, userMemberships } = useOrganizationList({
    userInvitations: {
      infinite: true,
    },
    userMemberships: {
      infinite: true,
    },
  });
  const { organizations, isLoadingOrgs, activeOrg, setActiveOrg, refreshOrganizations } = useOrganizationContext();
  const [acceptingInvite, setAcceptingInvite] = useState<string | null>(null);
  const [decliningInvite, setDecliningInvite] = useState<string | null>(null);
  const [switchingOrg, setSwitchingOrg] = useState<string | null>(null);

  useEffect(() => {
    if (isLoaded) {
      userInvitations.revalidate?.();
      userMemberships.revalidate?.();
      // Also refresh organizations from context
      refreshOrganizations();
    }
  }, [isLoaded]);

  const handleAcceptInvitation = async (invitationId: string, _orgId: string, orgName: string) => {
    setAcceptingInvite(invitationId);
    try {
      const invitation = userInvitations.data?.find((inv) => inv.id === invitationId);
      if (invitation) {
        await invitation.accept();
        await userMemberships.revalidate?.();
        await userInvitations.revalidate?.();
        
        // Refresh the organization context so it shows up in the list
        await refreshOrganizations();
        
        toast.success(`Successfully joined ${orgName}`);
      }
    } catch (error) {
      console.error("Failed to accept invitation:", error);
      toast.error("Failed to accept invitation");
    } finally {
      setAcceptingInvite(null);
    }
  };

  const handleDeclineInvitation = async (invitationId: string) => {
    setDecliningInvite(invitationId);
    try {
      const invitation = userInvitations.data?.find((inv) => inv.id === invitationId);
      if (invitation) {
        // Decline is not directly available in Clerk v5, but we can use revoke if we have admin access
        // For now, we'll just show a message
        toast.info("You can ignore invitations - they will expire automatically");
      }
    } catch (error) {
      console.error("Failed to decline invitation:", error);
      toast.error("Failed to decline invitation");
    } finally {
      setDecliningInvite(null);
    }
  };

  const handleSwitchOrganization = async (orgId: string) => {
    if (orgId === activeOrg?.id) return;
    
    setSwitchingOrg(orgId);
    try {
      // Find the organization from displayOrgs
      const clerkOrgs = userMemberships.data?.map((membership) => ({
        id: membership.organization.id,
        name: membership.organization.name,
        slug: membership.organization.slug || '',
        imageUrl: membership.organization.imageUrl,
        membersCount: membership.organization.membersCount,
        role: (membership.role === 'org:admin' || membership.role === 'admin' ? 'admin' : 'member') as 'admin' | 'member',
        createdAt: (typeof membership.organization.createdAt === 'string' 
          ? membership.organization.createdAt 
          : membership.organization.createdAt?.toISOString()) || new Date().toISOString(),
      })) || [];
      
      const displayOrgs = clerkOrgs.length > 0 ? clerkOrgs : organizations;
      const org = displayOrgs.find((o) => o.id === orgId);
      
      if (org && setActive) {
        // Set active in Clerk
        await setActive({ organization: orgId });
        
        // Update context
        setActiveOrg(org);
        
        toast.success(`Switched to ${org.name}`);
      }
    } catch (error) {
      console.error("Failed to switch organization:", error);
      toast.error("Failed to switch organization");
    } finally {
      setSwitchingOrg(null);
    }
  };

  const handleSwitchToPersonal = async () => {
    if (!activeOrg) return;
    
    setSwitchingOrg("personal");
    try {
      // Set active to null (personal account) in Clerk
      if (setActive) {
        await setActive({ organization: null });
      }
      
      // Update context
      await setActiveOrg(null);
      
      toast.success("Switched to personal workspace");
    } catch (error) {
      console.error("Failed to switch to personal:", error);
      toast.error("Failed to switch workspace");
    } finally {
      setSwitchingOrg(null);
    }
  };

  if (!isLoaded || isLoadingOrgs) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  const pendingInvitations = userInvitations.data || [];
  
  // Get organizations from both context and Clerk memberships
  const clerkOrgs = userMemberships.data?.map((membership) => ({
    id: membership.organization.id,
    name: membership.organization.name,
    slug: membership.organization.slug || '',
    imageUrl: membership.organization.imageUrl,
    membersCount: membership.organization.membersCount,
    role: (membership.role === 'org:admin' || membership.role === 'admin' ? 'admin' : 'member') as 'admin' | 'member',
    createdAt: (typeof membership.organization.createdAt === 'string' 
      ? membership.organization.createdAt 
      : membership.organization.createdAt?.toISOString()) || new Date().toISOString(),
  })) || [];
  
  // Use Clerk orgs as the source of truth, fallback to context
  const displayOrgs = clerkOrgs.length > 0 ? clerkOrgs : organizations;

  return (
    <div className="space-y-8">
      {/* Pending Invitations Section - Always show */}
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-semibold mb-1">Pending Invitations</h3>
          <p className="text-sm text-muted-foreground">
            {pendingInvitations.length > 0 
              ? `You have ${pendingInvitations.length} pending ${pendingInvitations.length === 1 ? "invitation" : "invitations"}`
              : "You have no pending invitations"
            }
          </p>
        </div>

        {pendingInvitations.length > 0 ? (
          <div className="space-y-3">
            {pendingInvitations.map((invitation) => {
              const orgName = invitation.publicOrganizationData.name;
              const orgImageUrl = invitation.publicOrganizationData.imageUrl;
              const isAccepting = acceptingInvite === invitation.id;
              const isDeclining = decliningInvite === invitation.id;

              return (
                <Card key={invitation.id} className="p-4 border-primary/20">
                  <div className="flex items-center gap-4">
                    {/* Organization Logo */}
                    {orgImageUrl ? (
                      <img
                        src={orgImageUrl}
                        alt={orgName}
                        className="h-12 w-12 rounded"
                      />
                    ) : (
                      <div className="h-12 w-12 rounded bg-primary/10 flex items-center justify-center">
                        <Building2 className="h-6 w-6 text-primary" />
                      </div>
                    )}

                    {/* Invitation Details */}
                    <div className="flex-1 min-w-0">
                      <h4 className="font-semibold truncate">{orgName}</h4>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground mt-0.5">
                        <Mail className="h-3 w-3" />
                        <span className="truncate">Invitation to join</span>
                      </div>
                    </div>

                    {/* Action Buttons */}
                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        onClick={() =>
                          handleAcceptInvitation(
                            invitation.id,
                            invitation.publicOrganizationData.id,
                            orgName
                          )
                        }
                        disabled={isAccepting || isDeclining}
                      >
                        {isAccepting ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <>
                            <Check className="h-4 w-4 mr-1" />
                            Accept
                          </>
                        )}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleDeclineInvitation(invitation.id)}
                        disabled={isAccepting || isDeclining}
                      >
                        {isDeclining ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <X className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>
                </Card>
              );
            })}
          </div>
        ) : (
          <Card className="p-6 text-center border-dashed">
            <p className="text-sm text-muted-foreground">
              No pending invitations at the moment.
            </p>
          </Card>
        )}
      </div>

      {/* Current Organizations Section */}
      <div className="space-y-4">
        <div>
          <h3 className="text-lg font-semibold mb-1">Your Workspaces</h3>
          <p className="text-sm text-muted-foreground">
            Switch between your personal workspace and organizations
          </p>
        </div>

        <div className="space-y-3">
          {/* Personal Workspace */}
          <Card
            className={`p-4 cursor-pointer transition-all hover:border-primary/50 ${
              !activeOrg ? "border-primary ring-2 ring-primary/20" : ""
            }`}
            onClick={handleSwitchToPersonal}
          >
            <div className="flex items-center gap-4">
              <div className="h-12 w-12 rounded bg-linear-to-br from-primary/20 to-primary/5 flex items-center justify-center">
                <UserCheck className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold">Personal Workspace</h4>
                <p className="text-sm text-muted-foreground">Your private notes and content</p>
              </div>
              {!activeOrg && (
                <Badge variant="default" className="ml-2">
                  Active
                </Badge>
              )}
              {switchingOrg === "personal" && (
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
              )}
            </div>
          </Card>

          {/* Organizations */}
          {displayOrgs.length > 0 ? (
            displayOrgs.map((org) => {
              const isActive = activeOrg?.id === org.id;
              const isSwitching = switchingOrg === org.id;

              return (
                <Card
                  key={org.id}
                  className={`p-4 cursor-pointer transition-all hover:border-primary/50 ${
                    isActive ? "border-primary ring-2 ring-primary/20" : ""
                  }`}
                  onClick={() => handleSwitchOrganization(org.id)}
                >
                  <div className="flex items-center gap-4">
                    {org.imageUrl ? (
                      <img
                        src={org.imageUrl}
                        alt={org.name}
                        className="h-12 w-12 rounded"
                      />
                    ) : (
                      <div className="h-12 w-12 rounded bg-primary/10 flex items-center justify-center">
                        <Building2 className="h-6 w-6 text-primary" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <h4 className="font-semibold truncate">{org.name}</h4>
                      <div className="flex items-center gap-2 mt-0.5">
                        <p className="text-sm text-muted-foreground">
                          {org.membersCount || "?"}{" "}
                          {org.membersCount === 1 ? "member" : "members"}
                        </p>
                        <span className="text-muted-foreground">•</span>
                        <Badge variant="outline" className="text-xs">
                          {org.role === "admin" ? "Admin" : "Member"}
                        </Badge>
                      </div>
                    </div>
                    {isActive && (
                      <Badge variant="default" className="ml-2">
                        Active
                      </Badge>
                    )}
                    {isSwitching && (
                      <Loader2 className="h-4 w-4 animate-spin text-primary" />
                    )}
                  </div>
                </Card>
              );
            })
          ) : (
            <Card className="p-8 text-center border-dashed">
              <Building2 className="h-12 w-12 mx-auto text-muted-foreground mb-3" />
              <p className="text-muted-foreground">
                You're not a member of any organizations yet.
              </p>
              {pendingInvitations.length > 0 && (
                <p className="text-sm text-muted-foreground mt-2">
                  Accept an invitation above to join an organization.
                </p>
              )}
            </Card>
          )}
        </div>
      </div>

      {/* Load More Button for Invitations */}
      {userInvitations.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => userInvitations.fetchNext?.()}
            disabled={userInvitations.isFetching}
          >
            {userInvitations.isFetching ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              "Load More Invitations"
            )}
          </Button>
        </div>
      )}
    </div>
  );
}

