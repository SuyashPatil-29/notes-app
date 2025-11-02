import type { Chapter } from "@/types/backend";
import api from "@/utils/api";

export const createChapter = async (data: Chapter) => {
    return await api.post("/chapter", data);
};

export const getChapter = async (id: string) => {
    return await api.get(`/chapter/${id}`);
};

export const getChaptersByNotebook = async (notebookId: string) => {
    return await api.get(`/notebooks/${notebookId}/chapters`);
};

export const updateChapter = async (id: string, data: Partial<Chapter>) => {
    return await api.put(`/chapter/${id}`, data);
};

export const deleteChapter = async (id: string) => {
    return await api.delete(`/chapter/${id}`);
};

export const moveChapter = async (id: string, notebookId: string, organizationId?: string) => {
    return await api.patch(`/chapter/${id}/move`, { 
        notebook_id: notebookId,
        organization_id: organizationId 
    });
};