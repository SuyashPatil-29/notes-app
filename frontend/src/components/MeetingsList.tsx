import { useQuery, useQueryClient, useMutation } from "@tanstack/react-query";
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Video,
  Clock,
  Download,
} from "lucide-react";
import {
  getUserMeetings,
  getMeetingPlatform,
  backfillVideoURLs,
} from "@/utils/meeting";
import type { MeetingListItem } from "@/types/backend";
import { useNavigate } from "react-router-dom";
import { formatDistanceToNow } from "date-fns";
import { useEffect, useRef, useState } from "react";
import { MeetingRecorder } from "./MeetingRecorder";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { toast } from "sonner";

export function MeetingsList() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const previousMeetingsRef = useRef<MeetingListItem[]>([]);
  const [isRecordDialogOpen, setIsRecordDialogOpen] = useState(false);

  const {
    data: meetings,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["userMeetings"],
    queryFn: getUserMeetings,
    refetchInterval: (query) => {
      // Refetch every 5 seconds if there are pending/processing meetings
      // OR if there are completed meetings without generated notes (note generation in progress)
      const hasActiveMeetings = query.state.data?.some(
        (meeting: MeetingListItem) =>
          meeting.status === "pending" ||
          meeting.status === "recording" ||
          meeting.status === "processing" ||
          (meeting.status === "completed" && !meeting.generatedNoteId)
      );
      return hasActiveMeetings ? 5000 : false;
    },
  });

  // Mutation for backfilling video URLs
  const backfillMutation = useMutation({
    mutationFn: backfillVideoURLs,
    onSuccess: (data) => {
      // Invalidate meetings query to refresh the list
      queryClient.invalidateQueries({ queryKey: ["userMeetings"] });
      
      // Show success toast
      if (data.updated_count === 0) {
        toast.success("All up to date - " + data.message);
      } else {
        toast.success(
          `Videos fetched successfully! Updated ${data.updated_count} meeting${data.updated_count !== 1 ? 's' : ''} with video recordings`
        );
      }

      // Show error details if any
      if (data.failed_count > 0 && data.errors) {
        console.error('Backfill errors:', data.errors);
        toast.error(
          `Some meetings couldn't be updated: ${data.failed_count} meeting${data.failed_count !== 1 ? 's' : ''} failed. Check console for details.`
        );
      }
    },
    onError: (error) => {
      toast.error(
        error.message || "Failed to fetch video URLs - An error occurred"
      );
    },
  });

  // Watch for newly completed meetings with generated notes and invalidate related queries
  useEffect(() => {
    if (!meetings || !previousMeetingsRef.current.length) {
      if (meetings) {
        previousMeetingsRef.current = meetings;
      }
      return;
    }

    const previousMeetings = previousMeetingsRef.current;
    const newlyCompletedWithNotes = meetings.filter((meeting) => {
      const previous = previousMeetings.find((m) => m.id === meeting.id);
      // Check if this meeting just got a generated note
      return (
        meeting.generatedNoteId &&
        (!previous || !previous.generatedNoteId)
      );
    });

    if (newlyCompletedWithNotes.length > 0) {
      
      // Invalidate all notebook-related queries to refresh sidebar
      // Since we don't have the full note details in the list item,
      // we invalidate all notebooks to be safe
      queryClient.invalidateQueries({ queryKey: ["notebooks"] });
      queryClient.invalidateQueries({ queryKey: ["userNotebooks"] });
    }

    previousMeetingsRef.current = meetings;
  }, [meetings, queryClient]);


  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-12">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center p-12 text-center">
        <XCircle className="w-12 h-12 text-red-500 mb-4" />
        <p className="text-muted-foreground">Failed to load meetings</p>
        <p className="text-sm text-muted-foreground mt-1">
          Please try refreshing the page
        </p>
      </div>
    );
  }

  if (!meetings || meetings.length === 0) {
    return (
      <>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {/* Record New Meeting Card */}
          <button
            onClick={() => setIsRecordDialogOpen(true)}
            className="border-2 border-dashed border-border/80 rounded-2xl p-6 h-40 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
            title="Record new meeting"
          >
            <span className="text-4xl leading-none">+</span>
          </button>
        </div>

        <div className="flex flex-col items-center justify-center p-12 text-center mt-8">
          <Video className="w-12 h-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No meeting recordings yet</p>
          <p className="text-sm text-muted-foreground mt-1">
            Click the + button above to start recording your first meeting
          </p>
        </div>

        {/* Record Meeting Dialog */}
        <Dialog open={isRecordDialogOpen} onOpenChange={setIsRecordDialogOpen}>
          <DialogContent className="sm:max-w-[600px]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <Video className="w-5 h-5" />
                Record Meeting
              </DialogTitle>
              <DialogDescription>
                Start recording a meeting from Zoom or Google Meet
              </DialogDescription>
            </DialogHeader>
            <MeetingRecorder onSuccess={() => setIsRecordDialogOpen(false)} />
          </DialogContent>
        </Dialog>
      </>
    );
  }

  // Check if any completed meetings are missing video URLs
  const meetingsNeedingVideoBackfill = meetings.filter(
    (m) => m.status === "completed" && m.recallRecordingId && !m.videoDownloadUrl
  );

  return (
    <>
      {/* Backfill Video URLs Banner */}
      {meetingsNeedingVideoBackfill.length > 0 && (
        <div className="mb-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Download className="w-5 h-5 text-blue-500" />
            <div>
              <p className="text-sm font-medium text-foreground">
                {meetingsNeedingVideoBackfill.length} meeting{meetingsNeedingVideoBackfill.length !== 1 ? 's' : ''} missing video recordings
              </p>
              <p className="text-xs text-muted-foreground">
                Fetch video recordings for your previous meetings
              </p>
            </div>
          </div>
          <button
            onClick={() => backfillMutation.mutate()}
            disabled={backfillMutation.isPending}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium flex items-center gap-2"
          >
            {backfillMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Fetching...
              </>
            ) : (
              <>
                <Download className="w-4 h-4" />
                Fetch Videos
              </>
            )}
          </button>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {/* Record New Meeting Card */}
        <button
          onClick={() => setIsRecordDialogOpen(true)}
          className="border-2 border-dashed border-border/80 rounded-2xl p-6 h-40 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-foreground/30 transition-colors"
          title="Record new meeting"
        >
          <span className="text-4xl leading-none">+</span>
        </button>

        {/* Existing Meetings */}
        {meetings.map((meeting) => (
        <button
          key={meeting.id}
          onClick={() => navigate(`/meetings/${meeting.id}`)}
          className="bg-card border border-border rounded-lg p-6 space-y-3 hover:border-primary/50 transition-colors text-left"
        >
          {/* Title with Status */}
          <div className="flex items-start justify-between gap-3">
            <h3 className="text-lg font-semibold text-card-foreground truncate flex-1">
              {getMeetingPlatform(meeting.meetingUrl)} Meeting
            </h3>
            {meeting.status === "completed" && (
              <CheckCircle2 className="w-5 h-5 text-green-500 shrink-0" />
            )}
            {meeting.status === "processing" && (
              <Loader2 className="w-5 h-5 animate-spin text-blue-500 shrink-0" />
            )}
            {meeting.status === "recording" && (
              <Video className="w-5 h-5 text-orange-500 shrink-0" />
            )}
            {meeting.status === "failed" && (
              <XCircle className="w-5 h-5 text-red-500 shrink-0" />
            )}
            {meeting.status === "pending" && (
              <Clock className="w-5 h-5 text-gray-500 shrink-0" />
            )}
          </div>

          {/* Metadata */}
          <p className="text-sm text-muted-foreground">
            {formatDistanceToNow(new Date(meeting.createdAt), {
              addSuffix: true,
            })}{" "}
            • {getMeetingPlatform(meeting.meetingUrl)}
            {meeting.generatedNoteId && " • Notes generated"}
          </p>
        </button>
        ))}
      </div>

      {/* Record Meeting Dialog */}
      <Dialog open={isRecordDialogOpen} onOpenChange={setIsRecordDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Video className="w-5 h-5" />
              Record Meeting
            </DialogTitle>
            <DialogDescription>
              Join a meeting and automatically transcribe it to your notes
            </DialogDescription>
          </DialogHeader>
          <MeetingRecorder onSuccess={() => setIsRecordDialogOpen(false)} />
        </DialogContent>
      </Dialog>
    </>
  );
}
