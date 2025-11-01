import api from './api';
import type { 
  MeetingRecording, 
  StartMeetingRecordingRequest, 
  StartMeetingRecordingResponse, 
  GetMeetingsResponse 
} from '../types/backend';

/**
 * Start a new meeting recording by creating a Recall.ai bot
 */
export const startMeetingRecording = async (meetingUrl: string): Promise<MeetingRecording> => {
  try {
    const response = await api.post<StartMeetingRecordingResponse>('/meeting/start', {
      meeting_url: meetingUrl,
    } as StartMeetingRecordingRequest);
    
    return response.data.data;
  } catch (error) {
    console.error('Failed to start meeting recording:', error);
    throw new Error('Failed to start meeting recording');
  }
};

/**
 * Get all meeting recordings for the authenticated user
 */
export const getUserMeetings = async (): Promise<MeetingRecording[]> => {
  try {
    const response = await api.get<GetMeetingsResponse>('/meetings');
    return response.data.data;
  } catch (error) {
    console.error('Failed to fetch user meetings:', error);
    throw new Error('Failed to fetch meetings');
  }
};

/**
 * Validate if a URL is a valid meeting URL format
 */
export const isValidMeetingUrl = (url: string): boolean => {
  try {
    const parsedUrl = new URL(url);
    
    // Check if it's HTTP/HTTPS
    if (parsedUrl.protocol !== 'http:' && parsedUrl.protocol !== 'https:') {
      return false;
    }
    
    // Check for supported meeting platforms
    const supportedDomains = [
      'meet.google.com',
      'zoom.us',
      'teams.microsoft.com',
      'webex.com',
      'gotomeeting.com',
    ];
    
    const hostname = parsedUrl.hostname.toLowerCase();
    
    // Check if it's a supported domain or subdomain
    for (const domain of supportedDomains) {
      if (hostname === domain || hostname.endsWith('.' + domain)) {
        return true;
      }
    }
    
    // Allow any domain for now (Recall.ai supports many platforms)
    return hostname.length > 0;
  } catch {
    return false;
  }
};

/**
 * Get a human-readable status label for meeting recording status
 */
export const getMeetingStatusLabel = (status: MeetingRecording['status']): string => {
  switch (status) {
    case 'pending':
      return 'Pending';
    case 'recording':
      return 'Recording';
    case 'processing':
      return 'Processing';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    default:
      return 'Unknown';
  }
};

/**
 * Get the appropriate CSS class for meeting status badge
 */
export const getMeetingStatusVariant = (status: MeetingRecording['status']): 'default' | 'secondary' | 'destructive' | 'outline' => {
  switch (status) {
    case 'completed':
      return 'default';
    case 'failed':
      return 'destructive';
    case 'processing':
    case 'recording':
      return 'secondary';
    case 'pending':
    default:
      return 'outline';
  }
};

/**
 * Extract platform name from meeting URL
 */
export const getMeetingPlatform = (url: string): string => {
  try {
    const parsedUrl = new URL(url);
    const hostname = parsedUrl.hostname.toLowerCase();
    
    if (hostname.includes('meet.google.com')) {
      return 'Google Meet';
    } else if (hostname.includes('zoom.us')) {
      return 'Zoom';
    } else if (hostname.includes('teams.microsoft.com')) {
      return 'Microsoft Teams';
    } else if (hostname.includes('webex.com')) {
      return 'Webex';
    } else if (hostname.includes('gotomeeting.com')) {
      return 'GoToMeeting';
    }
    
    return 'Unknown Platform';
  } catch {
    return 'Unknown Platform';
  }
};

/**
 * Get transcript for a specific meeting
 */
export interface TranscriptEntry {
  speaker: string;
  text: string;
  timestamp: string;
}

export interface MeetingTranscriptResponse {
  transcript: TranscriptEntry[] | null;
  meetingUrl: string;
  status: string;
  createdAt: string;
  completedAt: string;
  message?: string;
}

export const getMeetingTranscript = async (meetingId: string): Promise<MeetingTranscriptResponse> => {
  try {
    const response = await api.get<MeetingTranscriptResponse>(`/meeting/${meetingId}/transcript`);
    return response.data;
  } catch (error) {
    console.error('Failed to fetch meeting transcript:', error);
    throw new Error('Failed to fetch transcript');
  }
};