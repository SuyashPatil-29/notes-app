import api from "./api";
import type {
  Organization,
  CreateOrganizationRequest,
  ListOrganizationsResponse,
} from "../types/backend";

// Create a new organization
export const createOrganization = async (data: CreateOrganizationRequest): Promise<Organization> => {
  const response = await api.post<Organization>("/organizations", data);
  return response.data;
};

// Get all organizations the user is a member of
export const getUserOrganizations = async (): Promise<Organization[]> => {
  const response = await api.get<ListOrganizationsResponse>("/organizations");
  return response.data.organizations;
};

// Get a specific organization by ID
export const getOrganization = async (orgId: string): Promise<Organization> => {
  const response = await api.get<Organization>(`/organizations/${orgId}`);
  return response.data;
};

// Update an organization (admin only)
export const updateOrganization = async (
  orgId: string,
  data: Partial<CreateOrganizationRequest>
): Promise<Organization> => {
  const response = await api.put<Organization>(`/organizations/${orgId}`, data);
  return response.data;
};

// Delete an organization (admin only)
export const deleteOrganization = async (orgId: string): Promise<{ message: string }> => {
  const response = await api.delete<{ message: string }>(`/organizations/${orgId}`);
  return response.data;
};

