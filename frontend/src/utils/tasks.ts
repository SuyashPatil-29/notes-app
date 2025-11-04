import type {
  Task,
  TaskBoard,
  CreateTaskRequest,
  UpdateTaskRequest,
  CreateTaskBoardRequest,
  UpdateTaskBoardRequest,
  GenerateTasksResponse,
  GetUserTaskBoardsResponse,
  GetTasksForNoteResponse,
  GetOrganizationMembersResponse,
} from "@/types/backend";
import api from "@/utils/api";

// Task Board API functions
export const getUserTaskBoards = async (
  organizationId?: string | null,
  page: number = 1,
  pageSize: number = 50
): Promise<GetUserTaskBoardsResponse> => {
  const params: any = { page, page_size: pageSize };
  if (organizationId) {
    params.organizationId = organizationId;
  }
  const response = await api.get("/user/kanban", { params });
  return response.data;
};

export const getTaskBoard = async (boardId: string): Promise<TaskBoard> => {
  const response = await api.get(`/kanban/${boardId}`);
  return response.data;
};

export const createTaskBoard = async (data: CreateTaskBoardRequest): Promise<TaskBoard> => {
  const response = await api.post("/kanban", data);
  return response.data;
};

export const updateTaskBoard = async (
  boardId: string,
  data: UpdateTaskBoardRequest
): Promise<TaskBoard> => {
  const response = await api.put(`/kanban/${boardId}`, data);
  return response.data;
};

export const deleteTaskBoard = async (boardId: string): Promise<void> => {
  await api.delete(`/kanban/${boardId}`);
};

// Task API functions
export const createTask = async (
  boardId: string,
  data: CreateTaskRequest
): Promise<Task> => {
  const response = await api.post(`/kanban/${boardId}/tasks`, data);
  return response.data;
};

export const updateTask = async (
  taskId: string,
  data: UpdateTaskRequest
): Promise<Task> => {
  const response = await api.put(`/tasks/${taskId}`, data);
  return response.data;
};

export const deleteTask = async (taskId: string): Promise<void> => {
  await api.delete(`/tasks/${taskId}`);
};

// Note-associated task functions
export const getTasksForNote = async (noteId: string): Promise<GetTasksForNoteResponse> => {
  const response = await api.get(`/notes/${noteId}/tasks`);
  return response.data;
};

export const generateTasksFromNote = async (noteId: string): Promise<GenerateTasksResponse> => {
  const response = await api.post(`/notes/${noteId}/tasks/generate`);
  return response.data;
};

// Task assignment functions
export const assignTaskToUsers = async (
  taskId: string,
  userIds: string[]
): Promise<void> => {
  await api.post(`/tasks/${taskId}/assign`, { userIds });
};

export const unassignUserFromTask = async (
  taskId: string,
  userId: string
): Promise<void> => {
  await api.delete(`/tasks/${taskId}/assign/${userId}`);
};

// Organization member functions for task assignment
export const getOrganizationMembers = async (organizationId: string): Promise<GetOrganizationMembersResponse> => {
  const response = await api.get(`/organization/${organizationId}/members`);
  return response.data;
};

// Task utility functions
export const getTasksByStatus = (tasks: Task[], status: Task['status']): Task[] => {
  return tasks.filter(task => task.status === status);
};

export const getTaskCountByStatus = (tasks: Task[]): Record<Task['status'], number> => {
  return tasks.reduce((acc, task) => {
    acc[task.status] = (acc[task.status] || 0) + 1;
    return acc;
  }, {} as Record<Task['status'], number>);
};

export const sortTasksByPosition = (tasks: Task[]): Task[] => {
  return [...tasks].sort((a, b) => a.position - b.position);
};

export const getTaskPriorityColor = (priority: Task['priority']): string => {
  switch (priority) {
    case 'high':
      return 'bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/30';
    case 'medium':
      return 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30';
    case 'low':
      return 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30';
    default:
      return 'bg-slate-500/15 text-slate-700 dark:text-slate-400 border-slate-500/30';
  }
};

export const getTaskStatusColor = (status: Task['status']): string => {
  switch (status) {
    case 'todo':
      return 'text-blue-600 bg-blue-50 border-blue-200';
    case 'in_progress':
      return 'text-orange-600 bg-orange-50 border-orange-200';
    case 'done':
      return 'text-green-600 bg-green-50 border-green-200';
    default:
      return 'text-gray-600 bg-gray-50 border-gray-200';
  }
};

// Task validation functions
export const validateTaskTitle = (title: string): string | null => {
  if (!title.trim()) {
    return 'Task title is required';
  }
  if (title.length > 255) {
    return 'Task title must be less than 255 characters';
  }
  return null;
};

export const validateTaskDescription = (description: string): string | null => {
  if (description.length > 2000) {
    return 'Task description must be less than 2000 characters';
  }
  return null;
};

export const validateTaskBoardName = (name: string): string | null => {
  if (!name.trim()) {
    return 'Task board name is required';
  }
  if (name.length > 255) {
    return 'Task board name must be less than 255 characters';
  }
  return null;
};