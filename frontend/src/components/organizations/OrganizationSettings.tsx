import { useEffect, useState } from "react";
import { useUser } from "@clerk/clerk-react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TagsInput } from "@/components/ui/tags-input";
import { useOrganizationContext } from "@/contexts/OrganizationContext";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { 
  Users, 
  Mail, 
  Shield, 
  User, 
  UserMinus, 
  Loader2,
  Plus,
  X,
  Building2,
  Key,
  AlertTriangle
} from "lucide-react";
import { toast } from "sonner";
import api from "@/utils/api";
import {
  getOrgMembers,
  updateMemberRole,
  removeMember,
  getOrgInvitations,
  revokeInvitation,
  inviteMember,
} from "@/utils/organization-members";
import type { OrganizationMember, OrganizationInvitation } from "@/types/backend";
import type { Organization } from "@/types/backend";
import { OrganizationAPIKeySettings } from "./OrganizationAPIKeySettings";

interface OrganizationSettingsProps {
  organization: Organization;
  userRole?: string;
}

export function OrganizationSettings({
  organization,
  userRole,
}: OrganizationSettingsProps) {
  const { user } = useUser();
  const navigate = useNavigate();
  const { refreshOrganizations, setActiveOrg } = useOrganizationContext();
  const [orgName, setOrgName] = useState(organization.name);
  const [isSaving, setIsSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<'general' | 'members' | 'invitations' | 'api-keys'>('general');

  // Members and invitations state
  const [members, setMembers] = useState<OrganizationMember[]>([]);
  const [invitations, setInvitations] = useState<OrganizationInvitation[]>([]);
  const [isLoadingMembers, setIsLoadingMembers] = useState(true);
  const [isLoadingInvitations, setIsLoadingInvitations] = useState(true);
  
  // Dialogs state
  const [bulkInviteDialog, setBulkInviteDialog] = useState(false);
  const [removeMemberDialog, setRemoveMemberDialog] = useState<OrganizationMember | null>(null);
  const [revokeInviteDialog, setRevokeInviteDialog] = useState<OrganizationInvitation | null>(null);
  const [deleteOrgDialog, setDeleteOrgDialog] = useState(false);
  const [isDeletingOrg, setIsDeletingOrg] = useState(false);
  
  // Bulk invite state
  const [bulkEmails, setBulkEmails] = useState<string[]>([]);
  const [bulkRole, setBulkRole] = useState<"org:admin" | "org:member">("org:member");
  const [isSendingBulk, setIsSendingBulk] = useState(false);

  const isAdmin = userRole === "admin";

  useEffect(() => {
    if (organization?.id) {
      fetchMembers();
      fetchInvitations();
    }
  }, [organization?.id]);

  const fetchMembers = async () => {
    if (!organization?.id) return;
    
    setIsLoadingMembers(true);
    try {
      const data = await getOrgMembers(organization.id);
      setMembers(data);
    } catch (error: any) {
      console.error("Failed to fetch members:", error);
      toast.error(error.message || "Failed to load members");
    } finally {
      setIsLoadingMembers(false);
    }
  };

  const fetchInvitations = async () => {
    if (!organization?.id) return;
    
    setIsLoadingInvitations(true);
    try {
      const data = await getOrgInvitations(organization.id);
      setInvitations(data);
    } catch (error: any) {
      console.error("Failed to fetch invitations:", error);
      toast.error(error.message || "Failed to load invitations");
    } finally {
      setIsLoadingInvitations(false);
    }
  };

  const handleUpdateRole = async (userId: string, newRole: "org:admin" | "org:member") => {
    if (!organization?.id) return;
    
    try {
      await updateMemberRole(organization.id, userId, { role: newRole });
      toast.success("Member role updated successfully");
      fetchMembers();
    } catch (error: any) {
      console.error("Failed to update role:", error);
      toast.error(error.message || "Failed to update member role");
    }
  };

  const handleRemoveMember = async () => {
    if (!organization?.id || !removeMemberDialog) return;
    
    const isRemovingSelf = removeMemberDialog.userId === user?.id;
    
    try {
      await removeMember(organization.id, removeMemberDialog.userId);
      toast.success(
        isRemovingSelf 
          ? "You have left the organization" 
          : "Member removed successfully"
      );
      setRemoveMemberDialog(null);
      
      // If user removed themselves, refresh org list and switch to personal context
      if (isRemovingSelf) {
        await refreshOrganizations();
        await setActiveOrg(null);
      } else {
        // Just refresh the members list
        fetchMembers();
      }
    } catch (error: any) {
      console.error("Failed to remove member:", error);
      toast.error(error.message || "Failed to remove member");
    }
  };

  const handleRevokeInvitation = async () => {
    if (!organization?.id || !revokeInviteDialog) return;
    
    try {
      await revokeInvitation(organization.id, revokeInviteDialog.id);
      toast.success("Invitation revoked successfully");
      setRevokeInviteDialog(null);
      fetchInvitations();
    } catch (error: any) {
      console.error("Failed to revoke invitation:", error);
      toast.error(error.message || "Failed to revoke invitation");
    }
  };

  const handleBulkInvite = async () => {
    if (!organization?.id || bulkEmails.length === 0) return;
    
    // Validate all emails
    const invalidEmails = bulkEmails.filter(
      email => !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
    );
    
    if (invalidEmails.length > 0) {
      toast.error(`Invalid email addresses: ${invalidEmails.join(", ")}`);
      return;
    }

    setIsSendingBulk(true);
    let successCount = 0;
    let failCount = 0;
    const redirectUrl = `${window.location.origin}/accept-invitation`;

    try {
      // Send invitations in parallel
      const results = await Promise.allSettled(
        bulkEmails.map(email =>
          inviteMember(organization.id!, {
            emailAddress: email,
            role: bulkRole,
            redirectUrl,
          })
        )
      );

      results.forEach((result, index) => {
        if (result.status === "fulfilled") {
          successCount++;
        } else {
          failCount++;
          console.error(`Failed to invite ${bulkEmails[index]}:`, result.reason);
        }
      });

      if (successCount > 0) {
        toast.success(`Successfully sent ${successCount} invitation${successCount > 1 ? "s" : ""}`);
      }
      if (failCount > 0) {
        toast.error(`Failed to send ${failCount} invitation${failCount > 1 ? "s" : ""}`);
      }

      setBulkEmails([]);
      setBulkInviteDialog(false);
      fetchInvitations();
    } catch (error) {
      console.error("Bulk invite error:", error);
      toast.error("Failed to send invitations");
    } finally {
      setIsSendingBulk(false);
    }
  };

  const handleSaveSettings = async () => {
    if (!isAdmin) {
      toast.error("Only admins can update organization settings");
      return;
    }

    setIsSaving(true);
    try {
      // TODO: Implement update organization API call
      toast.info("Organization settings update - coming soon");
    } catch (error) {
      console.error("Failed to update organization:", error);
      toast.error("Failed to update organization settings");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteOrganization = async () => {
    if (!organization?.id) return;
    
    setIsDeletingOrg(true);
    try {
      await api.delete(`/organizations/${organization.id}`);
      toast.success("Organization deleted successfully");
      setDeleteOrgDialog(false);
      
      // Navigate away immediately to unmount this component
      // This prevents the useEffect hooks from trying to fetch data for a deleted org
      navigate("/");
      
      // Update state in the background after navigation
      // This ensures the org list is refreshed when user revisits
      setTimeout(async () => {
        setActiveOrg(null);
        await refreshOrganizations();
      }, 100);
    } catch (error: any) {
      console.error("Failed to delete organization:", error);
      const errorMessage = error.response?.data?.error || "Failed to delete organization";
      toast.error(errorMessage);
      setIsDeletingOrg(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Organization Settings</h2>
        <p className="text-muted-foreground">
          Manage your organization settings and members
        </p>
      </div>

      {/* Tabs Navigation - matching Profile page style */}
      <div className="flex items-center gap-1 border-b border-border">
        <button
          onClick={() => setActiveTab('general')}
          className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
            activeTab === 'general'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
            <Building2 className="h-4 w-4" />
            General
          {activeTab === 'general' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
          )}
        </button>
        
        <button
          onClick={() => setActiveTab('members')}
          className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
            activeTab === 'members'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
            <Users className="h-4 w-4" />
            Members
          {!isLoadingMembers && (
            <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
              {members.length}
            </Badge>
          )}
          {activeTab === 'members' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
          )}
        </button>

        <button
          onClick={() => setActiveTab('invitations')}
          className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
            activeTab === 'invitations'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Mail className="h-4 w-4" />
          Invitations
          {!isLoadingInvitations && invitations.length > 0 && (
            <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
              {invitations.length}
            </Badge>
          )}
          {activeTab === 'invitations' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
          )}
        </button>

        <button
          onClick={() => setActiveTab('api-keys')}
          className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
            activeTab === 'api-keys'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Key className="h-4 w-4" />
          API Keys
          {activeTab === 'api-keys' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
          )}
        </button>
      </div>

      {/* General Tab */}
      {activeTab === 'general' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Organization Details</CardTitle>
              <CardDescription>
                Update your organization's basic information
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="orgName">Organization Name</Label>
                <Input
                  id="orgName"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  disabled={!isAdmin || isSaving}
                  placeholder="Acme Inc"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="orgSlug">Slug</Label>
                <Input
                  id="orgSlug"
                  value={organization.slug}
                  disabled
                  placeholder="acme-inc"
                />
                <p className="text-sm text-muted-foreground">
                  The slug cannot be changed after creation
                </p>
              </div>

              <div className="space-y-2">
                <Label>Created</Label>
                <p className="text-sm text-muted-foreground">
                  {new Date(organization.createdAt).toLocaleDateString()}
                </p>
              </div>

              {isAdmin && (
                <div className="pt-4">
                  <Button
                    onClick={handleSaveSettings}
                    disabled={isSaving || orgName === organization.name}
                  >
                    {isSaving ? "Saving..." : "Save Changes"}
                  </Button>
                </div>
              )}

              {!isAdmin && (
                <p className="text-sm text-muted-foreground">
                  Only organization admins can update these settings
                </p>
              )}
            </CardContent>
          </Card>

          {isAdmin && (
            <Card className="border-destructive">
              <CardHeader>
                <CardTitle className="text-destructive">Danger Zone</CardTitle>
                <CardDescription>
                  Irreversible actions that affect your organization
                </CardDescription>
              </CardHeader>
              <CardContent>
                <Button
                  variant="destructive"
                  onClick={() => setDeleteOrgDialog(true)}
                  disabled={isDeletingOrg}
                >
                  Delete Organization
                </Button>
                <p className="text-sm text-muted-foreground mt-2">
                  Once deleted, all organization data will be permanently removed
                </p>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Members Tab */}
      {activeTab === 'members' && (
        <div className="space-y-6">
          {isAdmin && (
            <div className="flex justify-end">
              <Button onClick={() => setBulkInviteDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Bulk Invite
              </Button>
            </div>
          )}

          <Card>
            <CardHeader>
              <CardTitle>Team Members</CardTitle>
              <CardDescription>
                Manage your organization members and their roles
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoadingMembers ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-primary" />
                </div>
              ) : members.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No members found
                </div>
              ) : (
                <div className="space-y-4">
                  {members.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center justify-between p-4 rounded-lg border border-border hover:border-primary/50 transition-colors"
                    >
                      <div className="flex items-center gap-4">
                        <Avatar>
                          <AvatarImage src={member.publicUserData.imageUrl} />
                          <AvatarFallback>
                            <User className="h-5 w-5" />
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            {member.publicUserData.firstName || "Unknown"}{" "}
                            {member.publicUserData.lastName || "User"}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {member.publicUserData.identifier}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        {isAdmin ? (
                          <Select
                            value={member.role}
                            onValueChange={(value: "org:admin" | "org:member") =>
                              handleUpdateRole(member.userId, value)
                            }
                          >
                            <SelectTrigger className="w-32">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="org:admin">
                                <div className="flex items-center gap-2">
                                  <Shield className="h-4 w-4" />
                                  Admin
                                </div>
                              </SelectItem>
                              <SelectItem value="org:member">
                                <div className="flex items-center gap-2">
                                  <User className="h-4 w-4" />
                                  Member
                                </div>
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        ) : (
                          <Badge variant={member.role === "org:admin" ? "default" : "secondary"}>
                            {member.role === "org:admin" ? "Admin" : "Member"}
                          </Badge>
                        )}
                        {member.userId === user?.id && (
                          <Badge variant="outline" className="text-xs">
                            You
                          </Badge>
                        )}
                        {isAdmin && member.userId !== user?.id && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setRemoveMemberDialog(member)}
                          >
                            <UserMinus className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Invitations Tab */}
      {activeTab === 'invitations' && (
        <div className="space-y-6">
          {isAdmin && (
            <div className="flex justify-end">
              <Button onClick={() => setBulkInviteDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Bulk Invite
              </Button>
            </div>
          )}

          <Card>
            <CardHeader>
              <CardTitle>Pending Invitations</CardTitle>
              <CardDescription>
                Manage pending organization invitations
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoadingInvitations ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-primary" />
                </div>
              ) : invitations.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No pending invitations
                </div>
              ) : (
                <div className="space-y-4">
                  {invitations.map((invitation) => (
                    <div
                      key={invitation.id}
                      className="flex items-center justify-between p-4 rounded-lg border border-border hover:border-primary/50 transition-colors"
                    >
                      <div className="flex items-center gap-4">
                        <Mail className="h-5 w-5 text-muted-foreground" />
                        <div>
                          <div className="font-medium">{invitation.emailAddress}</div>
                          <div className="text-sm text-muted-foreground">
                            Invited {new Date(invitation.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <Badge variant="outline">Pending</Badge>
                        <Badge variant={invitation.role === "org:admin" ? "default" : "secondary"}>
                          {invitation.role === "org:admin" ? "Admin" : "Member"}
                        </Badge>
                        {isAdmin && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setRevokeInviteDialog(invitation)}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* API Keys Tab */}
      {activeTab === 'api-keys' && (
        <OrganizationAPIKeySettings 
          organizationId={organization.id}
          isAdmin={isAdmin}
        />
      )}

      {/* Bulk Invite Dialog */}
      <Dialog open={bulkInviteDialog} onOpenChange={setBulkInviteDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Bulk Invite Members</DialogTitle>
            <DialogDescription>
              Enter multiple email addresses to invite them all at once
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="bulk-emails">Email Addresses</Label>
              <TagsInput
                value={bulkEmails}
                onValueChange={setBulkEmails}
                placeholder="Type or paste email addresses (press Enter or comma to add)"
              />
              <p className="text-xs text-muted-foreground">
                Press Enter or comma after each email. You can also paste multiple emails at once.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="bulk-role">Role</Label>
              <Select
                value={bulkRole}
                onValueChange={(value: "org:admin" | "org:member") => setBulkRole(value)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="org:member">
                    <div className="flex items-center gap-2">
                      <User className="h-4 w-4" />
                      Member
                    </div>
                  </SelectItem>
                  <SelectItem value="org:admin">
                    <div className="flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Admin
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setBulkInviteDialog(false)}
              disabled={isSendingBulk}
            >
              Cancel
            </Button>
            <Button onClick={handleBulkInvite} disabled={isSendingBulk || bulkEmails.length === 0}>
              {isSendingBulk ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Sending to {bulkEmails.length} {bulkEmails.length === 1 ? "person" : "people"}...
                </>
              ) : (
                `Send ${bulkEmails.length} Invitation${bulkEmails.length !== 1 ? "s" : ""}`
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Remove Member Confirmation */}
      <Dialog open={!!removeMemberDialog} onOpenChange={() => setRemoveMemberDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove Member</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove {removeMemberDialog?.publicUserData.identifier} from the organization?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveMemberDialog(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRemoveMember}>
              Remove Member
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Revoke Invitation Confirmation */}
      <Dialog open={!!revokeInviteDialog} onOpenChange={() => setRevokeInviteDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke Invitation</DialogTitle>
            <DialogDescription>
              Are you sure you want to revoke the invitation for {revokeInviteDialog?.emailAddress}?
              They will no longer be able to join using this invitation link.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRevokeInviteDialog(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRevokeInvitation}>
              Revoke Invitation
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Organization Confirmation */}
      <Dialog open={deleteOrgDialog} onOpenChange={setDeleteOrgDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-destructive">
              <AlertTriangle className="h-5 w-5" />
              Delete Organization
            </DialogTitle>
            <DialogDescription>
              This action cannot be undone. This will permanently delete the <strong>{organization.name}</strong> organization and remove all associated data including:
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
              <li>All notebooks and notes</li>
              <li>All chapters and content</li>
              <li>All member access</li>
              <li>All pending invitations</li>
              <li>All organization settings</li>
            </ul>
          </div>
          <DialogFooter>
            <Button 
              variant="outline" 
              onClick={() => setDeleteOrgDialog(false)}
              disabled={isDeletingOrg}
            >
              Cancel
            </Button>
            <Button 
              variant="destructive" 
              onClick={handleDeleteOrganization}
              disabled={isDeletingOrg}
            >
              {isDeletingOrg ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                "Delete Organization"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

