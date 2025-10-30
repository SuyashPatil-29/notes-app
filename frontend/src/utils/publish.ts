import api from './api';
import type { Notebook, Chapter, Notes, PublishSettings, User } from '@/types/backend';

// Publishing API functions

export const publishNotebook = async (notebookId: string, noteIds: string[]) => {
  const response = await api.post(`/notebook/${notebookId}/publish`, {
    noteIds,
  });
  return response.data;
};

export const updatePublishedNotes = async (notebookId: string, noteIds: string[]) => {
  const response = await api.put(`/notebook/${notebookId}/published-notes`, {
    noteIds,
  });
  return response.data;
};

export const unpublishNotebook = async (notebookId: string) => {
  const response = await api.post(`/notebook/${notebookId}/unpublish`);
  return response.data;
};

export const toggleNotePublish = async (noteId: string) => {
  const response = await api.patch(`/note/${noteId}/publish`);
  return response.data;
};

// Public content API functions

export const getPublicNotebook = async (notebookId: string): Promise<Notebook> => {
  const response = await api.get(`/public/${notebookId}`);
  return response.data;
};

export const getPublicChapter = async (notebookId: string, chapterId: string): Promise<Chapter> => {
  const response = await api.get(`/public/${notebookId}/${chapterId}`);
  return response.data;
};

export const getPublicNote = async (notebookId: string, chapterId: string, noteId: string): Promise<Notes> => {
  const response = await api.get(`/public/${notebookId}/${chapterId}/${noteId}`);
  return response.data;
};

export const getPublicUserProfile = async (email: string): Promise<User> => {
  const response = await api.get(`/public/user/${email}`);
  return response.data;
};

// Utility functions for publish settings

export const getPublishSettings = (notebook: Notebook): PublishSettings => {
  const selectedNoteIds: string[] = [];

  notebook.chapters?.forEach(chapter => {
    chapter.notes?.forEach(note => {
      if (note.isPublic) {
        selectedNoteIds.push(note.id);
      }
    });
  });

  return {
    notebookId: notebook.id,
    selectedNoteIds,
  };
};

export const isNotebookPublished = (notebook: Notebook): boolean => {
  return notebook.isPublic;
};

export const getPublishedNoteCount = (notebook: Notebook): number => {
  let count = 0;
  notebook.chapters?.forEach(chapter => {
    chapter.notes?.forEach(note => {
      if (note.isPublic) {
        count++;
      }
    });
  });
  return count;
};

export const getPublishedChapterCount = (notebook: Notebook): number => {
  let count = 0;
  notebook.chapters?.forEach(chapter => {
    if (chapter.isPublic) {
      count++;
    }
  });
  return count;
};
