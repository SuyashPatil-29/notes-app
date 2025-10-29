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
  createdAt: string;
  updatedAt: string;
};

export type Chapter = {
  id: string;
  name: string;
  notebookId: string;
  notebook: Notebook;
  notes: Notes[];
  createdAt: string;
  updatedAt: string;
};

export type Notes = {
  id: string;
  name: string;
  content: string;
  chapterId: string;
  chapter: Chapter;
  createdAt: string;
  updatedAt: string;
};

export type AuthenticatedUser = Pick<User, 'id' | 'name' | 'email' | 'imageUrl'>;
