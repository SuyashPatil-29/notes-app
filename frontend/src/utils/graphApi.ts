import type { GraphData, NoteLink, CreateNoteLinkRequest } from '@/types/graph';
import { getStoredAuthToken } from './api';

const API_BASE_URL = 'https://notes-app-suyashpatil-295761-0ye3wo77.apn.leapcell.dev';

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = await getStoredAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  // Merge with any existing headers
  if (options.headers) {
    Object.assign(headers, options.headers);
  }

  const response = await fetch(url, {
    ...options,
    credentials: 'include',
    headers,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}

export async function getGraphData(searchQuery?: string, organizationId?: string | null): Promise<GraphData> {
  const url = new URL(`${API_BASE_URL}/api/graph/data`);
  if (searchQuery) {
    url.searchParams.append('q', searchQuery);
  }
  if (organizationId) {
    url.searchParams.append('organizationId', organizationId);
  }
  return fetchWithAuth(url.toString());
}

export async function createNoteLink(data: CreateNoteLinkRequest, organizationId?: string | null): Promise<NoteLink> {
  const url = new URL(`${API_BASE_URL}/api/notes/links`);
  if (organizationId) {
    url.searchParams.append('organizationId', organizationId);
  }
  return fetchWithAuth(url.toString(), {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function deleteNoteLink(linkId: string): Promise<void> {
  return fetchWithAuth(`${API_BASE_URL}/api/notes/links/${linkId}`, {
    method: 'DELETE',
  });
}

export async function updateNoteLink(linkId: string, linkType: string): Promise<NoteLink> {
  return fetchWithAuth(`${API_BASE_URL}/api/notes/links/${linkId}`, {
    method: 'PUT',
    body: JSON.stringify({ linkType }),
  });
}

export async function getNoteLinksByNoteId(noteId: string): Promise<NoteLink[]> {
  return fetchWithAuth(`${API_BASE_URL}/api/notes/${noteId}/links`);
}

export async function getAllLinks(): Promise<NoteLink[]> {
  return fetchWithAuth(`${API_BASE_URL}/api/notes/links`);
}

