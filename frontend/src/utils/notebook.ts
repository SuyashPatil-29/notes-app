import type { Notebook } from "@/types/backend";
import api from "@/utils/api";

export const getUserNotebooks = async (): Promise<Notebook[]> => {
    const response = await api.get("/notebooks");
    return response.data;
};

export const createNotebook = async (data: Notebook) => {
    return await api.post("/notebook", data);
};

export const getNotebook = async (id: string) => {
    return await api.get(`/notebook/${id}`);
};

export const updateNotebook = async (id: string, data: Partial<Notebook>) => {
    return await api.put(`/notebook/${id}`, data);
};

export const deleteNotebook = async (id: string) => {
    return await api.delete(`/notebook/${id}`);
};