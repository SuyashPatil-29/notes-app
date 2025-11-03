import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { inviteMember } from "@/utils/organization-members";
import { toast } from "sonner";

interface InviteMembersDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
}

export function InviteMembersDialog({
  open,
  onOpenChange,
  organizationId,
}: InviteMembersDialogProps) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<"org:admin" | "org:member">("org:member");
  // Default redirect to the invitations page
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) {
      toast.error("Email address is required");
      return;
    }

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      toast.error("Please enter a valid email address");
      return;
    }

    setIsLoading(true);
    try {
      await inviteMember(organizationId, {
        emailAddress: email.trim(),
        role,
          redirectUrl: `${window.location.origin}/accept-invitation`, // Redirect to public acceptance page
      });
      toast.success(`Invitation sent to ${email}`);
      setEmail("");
      setRole("org:member");
      onOpenChange(false);
    } catch (error: any) {
      console.error("Failed to invite member:", error);
      toast.error(error.message || "Failed to send invitation");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite Members</DialogTitle>
          <DialogDescription>
            Send an invitation to join this organization.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="colleague@example.com"
                disabled={isLoading}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role">Role</Label>
              <Select
                value={role}
                onValueChange={(value) => setRole(value as "org:admin" | "org:member")}
                disabled={isLoading}
              >
                <SelectTrigger id="role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="org:member">Member</SelectItem>
                  <SelectItem value="org:admin">Admin</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-muted-foreground">
                {role === "org:admin"
                  ? "Admins can manage organization settings and members."
                  : "Members can view and edit shared content."}
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || !email.trim()}>
              {isLoading ? "Sending..." : "Send Invitation"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

