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
  organizationId?: string | null;
  chapters: Chapter[];
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
};

// Lightweight notebook for list views
export type NotebookListItem = {
  id: string;
  name: string;
  clerkUserId: string;
  organizationId?: string | null;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
  chapterCount: number;
};

export type Chapter = {
  id: string;
  name: string;
  notebookId: string;
  organizationId?: string | null;
  notebook: Notebook;
  notes: Notes[];
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
};

// Lightweight chapter for list views
export type ChapterListItem = {
  id: string;
  name: string;
  notebookId: string;
  organizationId?: string | null;
  isPublic: boolean;
  createdAt: string;
  updatedAt: string;
  noteCount: number;
};

export type Notes = {
  id: string;
  name: string;
  content: string;
  chapterId: string;
  organizationId?: string | null;
  chapter: Chapter;
  isPublic: boolean;
  videoData?: string;
  hasVideo?: boolean;
  meetingRecordingId?: string;
  aiSummary?: string;
  transcriptRaw?: string;
  taskBoard?: TaskBoard;
  createdAt: string;
  updatedAt: string;
};

// Lightweight note for list views (without content)
export type NoteListItem = {
  id: string;
  name: string;
  chapterId: string;
  organizationId?: string | null;
  isPublic: boolean;
  hasVideo?: boolean;
  meetingRecordingId?: string;
  createdAt: string;
  updatedAt: string;
};

export type AuthenticatedUser = Pick<User, 'id' | 'name' | 'email' | 'imageUrl'> & {
  clerkUserId: string;
  onboardingCompleted: boolean;
  hasApiKey: boolean;
};

export type PublishSettings = {
  notebookId: string;
  selectedNoteIds: string[];
};

export type MeetingRecording = {
  id: string;
  userId: number;
  botId: string;
  meetingUrl: string;
  status: 'pending' | 'recording' | 'processing' | 'completed' | 'failed';
  recallRecordingId?: string;
  transcriptDownloadUrl?: string;
  videoDownloadUrl?: string;
  generatedNoteId?: string;
  generatedNote?: Notes;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
};

// Lightweight meeting for list views (without generated note details)
export type MeetingListItem = {
  id: string;
  clerkUserId: string;
  botId: string;
  meetingUrl: string;
  status: 'pending' | 'recording' | 'processing' | 'completed' | 'failed';
  recallRecordingId?: string;
  transcriptDownloadUrl?: string;
  videoDownloadUrl?: string;
  generatedNoteId?: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
};

export type StartMeetingRecordingRequest = {
  meeting_url: string;
};

export type StartMeetingRecordingResponse = {
  data: MeetingRecording;
  message: string;
};

export type GetMeetingsResponse = {
  data: MeetingListItem[];
};

// Calendar Integration Types
export type Calendar = {
  id: string;
  userId: number;
  recallCalendarId: string;
  platform: 'google_calendar' | 'microsoft_outlook';
  platformEmail: string;
  status: 'active' | 'inactive' | 'error';
  lastSyncedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CalendarEvent = {
  id: string;
  calendarId: string;
  recallEventId: string;
  meetingPlatform: string;
  meetingUrl: string;
  title: string;
  startTime: string;
  endTime: string;
  botScheduled: boolean;
  botId?: string;
  isDeleted: boolean;
  meetingRecordingId?: string;
  createdAt: string;
  updatedAt: string;
};

export type GetCalendarsResponse = {
  calendars: Calendar[];
};

export type GetCalendarEventsResponse = {
  events: CalendarEvent[];
};

export type InitiateCalendarAuthResponse = {
  authUrl: string;
};

export type SyncCalendarResponse = {
  message: string;
  syncedCount: number;
  lastSyncedAt: string;
};

// Organization Types
export type Organization = {
  id: string;
  name: string;
  slug: string;
  imageUrl?: string;
  createdAt: string;
  membersCount?: number;
  role?: 'admin' | 'member'; // Current user's role in this org
};

export type OrganizationMember = {
  id: string;
  userId: string;
  organizationId: string;
  role: 'org:admin' | 'org:member';
  createdAt: string;
  isCurrentUser?: boolean;
  publicUserData: {
    identifier: string;
    firstName?: string;
    lastName?: string;
    imageUrl?: string;
  };
};

export type OrganizationInvitation = {
  id: string;
  emailAddress: string;
  organizationId: string;
  role: 'org:admin' | 'org:member';
  status: 'pending' | 'accepted' | 'revoked';
  createdAt: string;
  organization?: {
    id: string;
    name: string;
    slug: string;
    imageUrl?: string;
  };
};

export type CreateOrganizationRequest = {
  name: string;
  slug?: string;
};

export type InviteMemberRequest = {
  emailAddress: string;
  role: 'org:admin' | 'org:member';
  redirectUrl?: string; // Optional URL to redirect after accepting invitation
};

export type UpdateMemberRoleRequest = {
  role: 'org:admin' | 'org:member';
};

export type ListOrganizationsResponse = {
  organizations: Organization[];
  total: number;
};

export type ListOrganizationMembersResponse = {
  members: OrganizationMember[];
  total: number;
};

export type ListOrganizationInvitationsResponse = {
  invitations: OrganizationInvitation[];
  total: number;
};

export type ListUserInvitationsResponse = {
  invitations: OrganizationInvitation[];
  total: number;
};
// Organization API Key Types
export type OrganizationAPIKeyStatus = {
  provider: string;
  hasKey: boolean;
  maskedKey?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
};

export type GetOrgAPICredentialsResponse = {
  credentials: OrganizationAPIKeyStatus[];
  total: number;
};

export type SetOrgAPICredentialRequest = {
  provider: string;
  apiKey: string;
};

export type DeleteOrgAPICredentialRequest = {
  provider: string;
};

// Task Management Types
export type TaskAssignment = {
  id: string;
  taskId: string;
  userId: string;
  createdAt: string;
};

export type Task = {
  id: string;
  title: string;
  description: string;
  status: 'backlog' | 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  taskBoardId: string;
  position: number;
  organizationId?: string | null;
  taskBoard?: TaskBoard;
  assignments?: TaskAssignment[];
  createdAt: string;
  updatedAt: string;
};

export type TaskBoard = {
  id: string;
  name: string;
  description: string;
  noteId?: string | null;
  clerkUserId: string;
  organizationId?: string | null;
  isStandalone: boolean;
  tasks: Task[];
  note?: Notes;
  createdAt: string;
  updatedAt: string;
};

// Lightweight task board for list views
export type TaskBoardListItem = {
  id: string;
  name: string;
  description: string;
  noteId?: string | null;
  clerkUserId: string;
  organizationId?: string | null;
  isStandalone: boolean;
  taskCount: number;
  createdAt: string;
  updatedAt: string;
};

// Task creation and update types
export type CreateTaskRequest = {
  title: string;
  description: string;
  status?: 'backlog' | 'todo' | 'in_progress' | 'done';
  priority?: 'low' | 'medium' | 'high';
  position?: number;
};

export type UpdateTaskRequest = {
  title?: string;
  description?: string;
  status?: 'backlog' | 'todo' | 'in_progress' | 'done';
  priority?: 'low' | 'medium' | 'high';
  position?: number;
};

export type CreateTaskBoardRequest = {
  name: string;
  description?: string;
  noteId?: string | null;
  isStandalone?: boolean;
};

export type UpdateTaskBoardRequest = {
  name?: string;
  description?: string;
};

// Task generation types
export type GenerateTasksRequest = {
  noteId: string;
};

export type GenerateTasksResponse = {
  message: string;
  taskBoard: TaskBoard;
};

// API response types with pagination
export type GetUserTaskBoardsResponse = {
  data: TaskBoardListItem[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
};

export type GetTasksForNoteResponse = {
  taskBoard: TaskBoard | null;
  tasks: Task[];
};

// Task Assignment Types
export type OrganizationMemberForAssignment = {
  id: string;
  name: string;
  email: string;
  imageUrl?: string;
  role: 'admin' | 'member';
};

export type AssignTaskRequest = {
  userIds: string[];
};

export type AssignTaskResponse = {
  message: string;
  assignments: TaskAssignment[];
};

export type GetOrganizationMembersResponse = {
  members: OrganizationMemberForAssignment[];
  total: number;
};