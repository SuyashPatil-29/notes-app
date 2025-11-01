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
  meetingRecordingId?: string;
  aiSummary?: string;
  transcriptRaw?: string;
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

export type StartMeetingRecordingRequest = {
  meeting_url: string;
};

export type StartMeetingRecordingResponse = {
  data: MeetingRecording;
  message: string;
};

export type GetMeetingsResponse = {
  data: MeetingRecording[];
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
