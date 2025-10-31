export type User = {
  id: number;
  name: string;
  email: string;
  imageUrl: string | null;
  notebooks: Notebook[];
  createdAt: string;
  updatedAt: string;
};

export type Notebook = {
  id: string;
  name: string;
  userId: number;
  chapters: Chapter[];
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
};

export type Chapter = {
  id: string;
  name: string;
  notebookId: string;
  notebook: Notebook;
  notes: Notes[];
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
};

export type Notes = {
  id: string;
  name: string;
  content: string;
  chapterId: string;
  chapter: Chapter;
  isPublic: boolean;
  videoData?: string;
  hasVideo?: boolean;
  createdAt: string;
  updatedAt: string;
};

export type AuthenticatedUser = Pick<User, 'id' | 'name' | 'email' | 'imageUrl'> & {
  onboardingCompleted: boolean;
  hasApiKey: boolean;
};

export type PublishSettings = {
  notebookId: string;
  selectedNoteIds: string[];
};
