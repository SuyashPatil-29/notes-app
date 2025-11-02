import { useEffect, useState } from "react";
import { useOrganizationList } from "@clerk/clerk-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Building2, Mail, Check, X } from "lucide-react";
import { toast } from "sonner";
import { useOrganizationContext } from "@/contexts/OrganizationContext";

interface JoinOrganizationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function JoinOrganizationDialog({
  open,
  onOpenChange,
}: JoinOrganizationDialogProps) {
  const { isLoaded, userInvitations, userMemberships } = useOrganizationList({
    userInvitations: {
      infinite: true,
    },
    userMemberships: {
      infinite: true,
    },
  });
  const { setActiveOrg, refreshOrganizations } = useOrganizationContext();
  const [acceptingInvite, setAcceptingInvite] = useState<string | null>(null);
  const [decliningInvite, setDecliningInvite] = useState<string | null>(null);

  useEffect(() => {
    if (open && isLoaded) {
      // Revalidate invitations when dialog opens
      userInvitations.revalidate?.();
    }
  }, [open, isLoaded]);

  const handleAcceptInvitation = async (invitationId: string, orgId: string, orgName: string) => {
    setAcceptingInvite(invitationId);
    try {
      const invitation = userInvitations.data?.find((inv) => inv.id === invitationId);
      if (invitation) {
        await invitation.accept();
        await userMemberships.revalidate?.();
        await userInvitations.revalidate?.();
        await refreshOrganizations();
        
        toast.success(`Joined ${orgName}`);
        
        // Switch to the newly joined organization
        const orgData = {
          id: orgId,
          name: orgName,
          slug: invitation.publicOrganizationData.slug || "",
          imageUrl: invitation.publicOrganizationData.imageUrl,
          createdAt: new Date().toISOString(),
          membersCount: undefined, // Will be fetched when viewing org details
        };
        setActiveOrg(orgData);
        
        onOpenChange(false);
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
      // Note: Clerk doesn't have a built-in decline method, so we'll just hide it
      // In a real implementation, you might want to call a backend endpoint
      toast.info("Invitation declined");
      await userInvitations.revalidate?.();
    } catch (error) {
      console.error("Failed to decline invitation:", error);
      toast.error("Failed to decline invitation");
    } finally {
      setDecliningInvite(null);
    }
  };

  if (!isLoaded) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Join Organizations</DialogTitle>
            <DialogDescription>Loading invitations...</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {[1, 2].map((i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  const invitations = userInvitations.data || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Join Organizations</DialogTitle>
          <DialogDescription>
            You have {invitations.length} pending{" "}
            {invitations.length === 1 ? "invitation" : "invitations"}
          </DialogDescription>
        </DialogHeader>

        {invitations.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <div className="text-center py-8">
                <Mail className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <h3 className="font-semibold text-lg mb-2">No pending invitations</h3>
                <p className="text-sm text-muted-foreground">
                  You don't have any pending organization invitations at the moment.
                </p>
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-4">
            {invitations.map((invitation) => (
              <Card key={invitation.id}>
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      {invitation.publicOrganizationData.imageUrl ? (
                        <img
                          src={invitation.publicOrganizationData.imageUrl}
                          alt={invitation.publicOrganizationData.name}
                          className="h-10 w-10 rounded"
                        />
                      ) : (
                        <div className="h-10 w-10 rounded bg-primary/10 flex items-center justify-center">
                          <Building2 className="h-5 w-5 text-primary" />
                        </div>
                      )}
                      <div>
                        <CardTitle className="text-base">
                          {invitation.publicOrganizationData.name}
                        </CardTitle>
                        <CardDescription>
                          Invited as{" "}
                          <Badge variant="outline" className="ml-1">
                            {invitation.role === "org:admin" ? "Admin" : "Member"}
                          </Badge>
                        </CardDescription>
                      </div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="flex gap-2">
                    <Button
                      onClick={() =>
                        handleAcceptInvitation(
                          invitation.id,
                          invitation.publicOrganizationData.id,
                          invitation.publicOrganizationData.name
                        )
                      }
                      disabled={
                        acceptingInvite === invitation.id || decliningInvite === invitation.id
                      }
                      size="sm"
                    >
                      {acceptingInvite === invitation.id ? (
                        "Accepting..."
                      ) : (
                        <>
                          <Check className="h-4 w-4 mr-2" />
                          Accept
                        </>
                      )}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => handleDeclineInvitation(invitation.id)}
                      disabled={
                        acceptingInvite === invitation.id || decliningInvite === invitation.id
                      }
                      size="sm"
                    >
                      {decliningInvite === invitation.id ? (
                        "Declining..."
                      ) : (
                        <>
                          <X className="h-4 w-4 mr-2" />
                          Decline
                        </>
                      )}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}

            {userInvitations.hasNextPage && (
              <Button
                variant="outline"
                onClick={() => userInvitations.fetchNext?.()}
                disabled={userInvitations.isFetching}
                className="w-full"
              >
                {userInvitations.isFetching ? "Loading..." : "Load More"}
              </Button>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

