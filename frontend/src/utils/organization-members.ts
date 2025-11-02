import api from "./api";
import type {
  OrganizationMember,
  OrganizationInvitation,
  InviteMemberRequest,
  UpdateMemberRoleRequest,
  ListOrganizationMembersResponse,
  ListOrganizationInvitationsResponse,
  ListUserInvitationsResponse,
} from "../types/backend";

// ============= Member Management =============

// Get all members of an organization
export const getOrgMembers = async (orgId: string): Promise<OrganizationMember[]> => {
  const response = await api.get<ListOrganizationMembersResponse>(`/organizations/${orgId}/members`);
  return response.data.members;
};

// Update a member's role (admin only)
export const updateMemberRole = async (
  orgId: string,
  userId: string,
  data: UpdateMemberRoleRequest
): Promise<OrganizationMember> => {
  const response = await api.put<OrganizationMember>(
    `/organizations/${orgId}/members/${userId}`,
    data
  );
  return response.data;
};

// Remove a member from organization (admin only)
export const removeMember = async (
  orgId: string,
  userId: string
): Promise<{ message: string }> => {
  const response = await api.delete<{ message: string }>(
    `/organizations/${orgId}/members/${userId}`
  );
  return response.data;
};

// ============= Invitation Management (Admin) =============

// Send an invitation to join organization (admin only)
export const inviteMember = async (
  orgId: string,
  data: InviteMemberRequest
): Promise<OrganizationInvitation> => {
  const response = await api.post<OrganizationInvitation>(
    `/organizations/${orgId}/invitations`,
    data
  );
  return response.data;
};

// Get all pending invitations for an organization (admin/members)
export const getOrgInvitations = async (orgId: string): Promise<OrganizationInvitation[]> => {
  const response = await api.get<ListOrganizationInvitationsResponse>(
    `/organizations/${orgId}/invitations`
  );
  return response.data.invitations;
};

// Revoke an invitation (admin only)
export const revokeInvitation = async (
  orgId: string,
  invitationId: string
): Promise<{ message: string }> => {
  const response = await api.delete<{ message: string }>(
    `/organizations/${orgId}/invitations/${invitationId}`
  );
  return response.data;
};

// ============= User Invitations (Current User) =============

// Get all pending invitations for the current user
export const getUserInvitations = async (): Promise<OrganizationInvitation[]> => {
  const response = await api.get<ListUserInvitationsResponse>("/user/invitations");
  return response.data.invitations;
};

// Accept an invitation to join an organization
export const acceptInvitation = async (invitationId: string): Promise<{ message: string; organizationId: string }> => {
  const response = await api.post<{ message: string; organizationId: string }>(
    `/user/invitations/${invitationId}/accept`
  );
  return response.data;
};

// Decline an invitation to join an organization
export const declineInvitation = async (invitationId: string): Promise<{ message: string }> => {
  const response = await api.post<{ message: string }>(
    `/user/invitations/${invitationId}/decline`
  );
  return response.data;
};

