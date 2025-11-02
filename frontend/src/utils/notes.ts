import type { Notes } from "@/types/backend";
import api from "@/utils/api";

export const createNote = async (data: Notes) => {
    return await api.post("/note", data);
};

export const getNote = async (id: string) => {
    return await api.get(`/note/${id}`);
};

export const getNotesByChapter = async (chapterId: string) => {
    return await api.get(`/chapters/${chapterId}/notes`);
};

export const updateNote = async (id: string, data: Partial<Notes>) => {
    return await api.put(`/note/${id}`, data);
};

export const deleteNote = async (id: string) => {
    return await api.delete(`/note/${id}`);
};

export const moveNote = async (id: string, chapterId: string, organizationId?: string) => {
  return await api.patch(`/note/${id}/move`, { 
    chapter_id: chapterId,
    organization_id: organizationId 
  });
};

export const generateNoteVideo = async (noteId: string) => {
  return await api.post(`/note/${noteId}/generate-video`);
};

export const deleteNoteVideo = async (noteId: string) => {
  return await api.delete(`/note/${noteId}/video`);
};

