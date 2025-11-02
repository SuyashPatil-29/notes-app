import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Building2, Users } from "lucide-react";
import { OrganizationMembersList } from "./OrganizationMembersList";
import type { Organization } from "@/types/backend";
import { toast } from "sonner";

interface OrganizationSettingsProps {
  organization: Organization;
  userRole?: string;
}

export function OrganizationSettings({
  organization,
  userRole,
}: OrganizationSettingsProps) {
  const [orgName, setOrgName] = useState(organization.name);
  const [isSaving, setIsSaving] = useState(false);

  const isAdmin = userRole === "admin";

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

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Organization Settings</h2>
        <p className="text-muted-foreground">
          Manage your organization settings and members
        </p>
      </div>

      <Tabs defaultValue="general" className="space-y-4">
        <TabsList>
          <TabsTrigger value="general" className="gap-2">
            <Building2 className="h-4 w-4" />
            General
          </TabsTrigger>
          <TabsTrigger value="members" className="gap-2">
            <Users className="h-4 w-4" />
            Members
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-4">
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
                  onClick={() => toast.info("Delete organization - coming soon")}
                >
                  Delete Organization
                </Button>
                <p className="text-sm text-muted-foreground mt-2">
                  Once deleted, all organization data will be permanently removed
                </p>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="members" className="space-y-4">
          <OrganizationMembersList
            organizationId={organization.id}
            currentUserRole={userRole}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

