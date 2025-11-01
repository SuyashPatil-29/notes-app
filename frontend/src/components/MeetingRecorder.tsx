import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, CheckCircle2, XCircle } from 'lucide-react';
import { startMeetingRecording, isValidMeetingUrl, getMeetingPlatform } from '@/utils/meeting';
import { toast } from 'sonner';

interface MeetingRecorderProps {
  onSuccess?: () => void;
}

export function MeetingRecorder({ onSuccess }: MeetingRecorderProps) {
  const [meetingUrl, setMeetingUrl] = useState('');
  const [showSuccess, setShowSuccess] = useState(false);
  const queryClient = useQueryClient();
  
  const startRecording = useMutation({
    mutationFn: startMeetingRecording,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['userMeetings'] });
      setMeetingUrl('');
      setShowSuccess(true);
      toast.success('Meeting recording started successfully!');
      
      // Hide success message after 3 seconds, then close dialog
      setTimeout(() => {
        setShowSuccess(false);
        onSuccess?.();
      }, 3000);
    },
    onError: (error) => {
      console.error('Failed to start meeting recording:', error);
      toast.error('Failed to start meeting recording. Please try again.');
    },
  });
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const trimmedUrl = meetingUrl.trim();
    if (!trimmedUrl) {
      toast.error('Please enter a meeting URL');
      return;
    }
    
    if (!isValidMeetingUrl(trimmedUrl)) {
      toast.error('Please enter a valid meeting URL');
      return;
    }
    
    startRecording.mutate(trimmedUrl);
  };
  
  const isUrlValid = meetingUrl.trim() ? isValidMeetingUrl(meetingUrl.trim()) : true;
  const platform = meetingUrl.trim() ? getMeetingPlatform(meetingUrl.trim()) : '';
  
  return (
    <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="meeting-url">Meeting URL</Label>
            <Input
              id="meeting-url"
              type="url"
              placeholder="https://meet.google.com/abc-defg-hij"
              value={meetingUrl}
              onChange={(e) => setMeetingUrl(e.target.value)}
              disabled={startRecording.isPending}
              className={!isUrlValid ? 'border-red-500' : ''}
            />
            {meetingUrl.trim() && isUrlValid && platform !== 'Unknown Platform' && (
              <p className="text-sm text-muted-foreground flex items-center gap-1">
                <CheckCircle2 className="w-3 h-3 text-green-500" />
                Detected: {platform}
              </p>
            )}
            {meetingUrl.trim() && !isUrlValid && (
              <p className="text-sm text-red-500 flex items-center gap-1">
                <XCircle className="w-3 h-3" />
                Invalid meeting URL format
              </p>
            )}
            <p className="text-sm text-muted-foreground">
              Supports Google Meet, Zoom, Microsoft Teams, and more
            </p>
          </div>
          
          {startRecording.isError && (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertDescription>
                Failed to start recording. Please check the meeting URL and try again.
              </AlertDescription>
            </Alert>
          )}
          
          {showSuccess && (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>
                Bot is joining your meeting! Transcription will be saved automatically when the meeting ends.
              </AlertDescription>
            </Alert>
          )}
          
          <Button
            type="submit"
            disabled={!meetingUrl.trim() || !isUrlValid || startRecording.isPending}
            className="w-full"
          >
            {startRecording.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {startRecording.isPending ? 'Starting Bot...' : 'Start Recording'}
        </Button>
      </form>
  );
}