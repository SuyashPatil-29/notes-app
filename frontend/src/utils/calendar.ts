import api from "./api";
import type {
  GetCalendarsResponse,
  GetCalendarEventsResponse,
  InitiateCalendarAuthResponse,
  SyncCalendarResponse,
} from "@/types/backend";

/**
 * Initiate OAuth flow for calendar connection
 * Opens OAuth URL in the current window
 */
export const initiateCalendarAuth = async (provider: 'google' | 'microsoft') => {
  const response = await api.post<InitiateCalendarAuthResponse>(
    `/api/calendar-auth/${provider}`
  );
  // Redirect to OAuth URL
  window.location.href = response.data.authUrl;
};

/**
 * Get all calendars connected by the user
 */
export const getUserCalendars = async () => {
  const response = await api.get<GetCalendarsResponse>('/api/calendars');
  return response.data.calendars;
};

/**
 * Disconnect a calendar
 */
export const disconnectCalendar = async (calendarId: string) => {
  await api.delete(`/api/calendars/${calendarId}`);
};

/**
 * Get events for a specific calendar
 */
export const getCalendarEvents = async (
  calendarId: string,
  upcomingOnly: boolean = true
) => {
  const response = await api.get<GetCalendarEventsResponse>(
    `/api/calendars/${calendarId}/events`,
    {
      params: { upcoming: upcomingOnly },
    }
  );
  return response.data.events;
};

/**
 * Manually sync calendar events from Recall.ai
 */
export const syncCalendar = async (calendarId: string) => {
  const response = await api.post<SyncCalendarResponse>(
    `/api/calendars/${calendarId}/sync`
  );
  return response.data;
};

/**
 * Sync missing calendars from Recall.ai
 * This fetches all calendars from Recall for existing connections
 * and saves any that are missing in the database
 */
export const syncMissingCalendars = async () => {
  const response = await api.post<{ message: string; added: number }>(
    '/api/calendars/sync-missing'
  );
  return response.data;
};

/**
 * Schedule a bot to join a specific event
 */
export const scheduleBotForEvent = async (eventId: string) => {
  const response = await api.post(`/api/calendar-events/${eventId}/schedule-bot`);
  return response.data;
};

/**
 * Cancel a scheduled bot for an event
 */
export const cancelBotForEvent = async (eventId: string) => {
  await api.delete(`/api/calendar-events/${eventId}/cancel-bot`);
};

/**
 * Format calendar platform name for display
 */
export const formatPlatformName = (platform: string): string => {
  switch (platform) {
    case 'google_calendar':
      return 'Google Calendar';
    case 'microsoft_outlook':
      return 'Microsoft Outlook';
    default:
      return platform;
  }
};

/**
 * Format meeting platform name for display
 */
export const formatMeetingPlatform = (platform: string): string => {
  switch (platform) {
    case 'zoom':
      return 'Zoom';
    case 'google_meet':
      return 'Google Meet';
    case 'microsoft_teams':
      return 'Microsoft Teams';
    case 'webex':
      return 'Webex';
    default:
      return platform;
  }
};

/**
 * Get platform icon/emoji
 */
export const getPlatformIcon = (platform: string): string => {
  switch (platform) {
    case 'google_calendar':
      return '📅';
    case 'microsoft_outlook':
      return '📆';
    default:
      return '📅';
  }
};

/**
 * Get meeting platform icon/emoji
 */
export const getMeetingPlatformIcon = (platform: string): string => {
  switch (platform) {
    case 'zoom':
      return '🎥';
    case 'google_meet':
      return '📹';
    case 'microsoft_teams':
      return '👥';
    case 'webex':
      return '🎦';
    default:
      return '🎥';
  }
};

