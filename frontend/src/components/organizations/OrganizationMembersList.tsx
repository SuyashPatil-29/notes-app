import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { UserMinus, Mail } from "lucide-react";
import { getOrgMembers } from "@/utils/organization-members";
import type { OrganizationMember } from "@/types/backend";
import { toast } from "sonner";
import { InviteMembersDialog } from "./InviteMembersDialog";

interface OrganizationMembersListProps {
  organizationId: string;
  currentUserRole?: string;
}

export function OrganizationMembersList({
  organizationId,
  currentUserRole,
}: OrganizationMembersListProps) {
  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

  const isAdmin = currentUserRole === "admin";

  const loadMembers = async () => {
    try {
      setIsLoading(true);
      const data = await getOrgMembers(organizationId);
      setMembers(data);
    } catch (error) {
      console.error("Failed to load members:", error);
      toast.error("Failed to load organization members");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadMembers();
  }, [organizationId]);

  const getInitials = (firstName?: string, lastName?: string) => {
    const first = firstName?.charAt(0) || "";
    const last = lastName?.charAt(0) || "";
    return (first + last).toUpperCase() || "?";
  };

  const getRoleLabel = (role: string) => {
    switch (role) {
      case "org:admin":
        return "Admin";
      case "org:member":
        return "Member";
      default:
        return role;
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
          <CardDescription>Loading members...</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center gap-4">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="space-y-2 flex-1">
                <Skeleton className="h-4 w-[200px]" />
                <Skeleton className="h-3 w-[100px]" />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Members</CardTitle>
              <CardDescription>
                {members.length} {members.length === 1 ? "member" : "members"} in this
                organization
              </CardDescription>
            </div>
            {isAdmin && (
              <Button onClick={() => setIsInviteDialogOpen(true)} size="sm">
                <Mail className="h-4 w-4 mr-2" />
                Invite
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {members.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between p-4 border rounded-lg"
              >
                <div className="flex items-center gap-4">
                  <Avatar>
                    <AvatarImage
                      src={member.publicUserData.imageUrl}
                      alt={member.publicUserData.identifier}
                    />
                    <AvatarFallback>
                      {getInitials(
                        member.publicUserData.firstName,
                        member.publicUserData.lastName
                      )}
                    </AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="font-medium">
                        {member.publicUserData.firstName}{" "}
                        {member.publicUserData.lastName}
                      </p>
                      {member.isCurrentUser && (
                        <Badge variant="secondary" className="text-xs">
                          You
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {member.publicUserData.identifier}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={member.role === "org:admin" ? "default" : "outline"}>
                    {getRoleLabel(member.role)}
                  </Badge>
                  {isAdmin && !member.isCurrentUser && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        toast.info("Member management temporarily unavailable");
                      }}
                      title="Remove member"
                    >
                      <UserMinus className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <InviteMembersDialog
        open={isInviteDialogOpen}
        onOpenChange={(open) => {
          setIsInviteDialogOpen(open);
          if (!open) {
            loadMembers(); // Refresh after closing in case invitations were sent
          }
        }}
        organizationId={organizationId}
      />
    </>
  );
}

