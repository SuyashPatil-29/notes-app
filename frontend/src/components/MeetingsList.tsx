import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Video,
  Clock,
} from "lucide-react";
import {
  getUserMeetings,
  getMeetingPlatform,
} from "@/utils/meeting";
import type { MeetingRecording } from "@/types/backend";
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

export function MeetingsList() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const previousMeetingsRef = useRef<MeetingRecording[]>([]);
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
        (meeting: MeetingRecording) =>
          meeting.status === "pending" ||
          meeting.status === "recording" ||
          meeting.status === "processing" ||
          (meeting.status === "completed" && !meeting.generatedNote)
      );
      return hasActiveMeetings ? 5000 : false;
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
        meeting.generatedNote &&
        (!previous || !previous.generatedNote)
      );
    });

    if (newlyCompletedWithNotes.length > 0) {
      console.log("Detected newly generated notes, invalidating queries...");
      
      // Invalidate all notebook-related queries to refresh sidebar
      queryClient.invalidateQueries({ queryKey: ["notebooks"] });
      queryClient.invalidateQueries({ queryKey: ["userNotebooks"] });
      
      // Invalidate chapter queries for the affected notebooks
      newlyCompletedWithNotes.forEach((meeting) => {
        if (meeting.generatedNote?.chapter?.notebookId) {
          queryClient.invalidateQueries({
            queryKey: ["chapters", meeting.generatedNote.chapter.notebookId],
          });
        }
      });
      
      // Invalidate notes queries for the affected chapters
      newlyCompletedWithNotes.forEach((meeting) => {
        if (meeting.generatedNote?.chapterId) {
          queryClient.invalidateQueries({
            queryKey: ["notes", meeting.generatedNote.chapterId],
          });
        }
      });
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
      <div className="flex flex-col items-center justify-center p-12 text-center">
        <Video className="w-12 h-12 text-muted-foreground mb-4" />
        <p className="text-muted-foreground">No meeting recordings yet</p>
        <p className="text-sm text-muted-foreground mt-1">
          Start recording a meeting to see it here
        </p>
      </div>
    );
  }

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
              {meeting.generatedNote?.name ||
                `${getMeetingPlatform(meeting.meetingUrl)} Meeting`}
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
          </p>

          {/* View Notes Button */}
          {meeting.generatedNote && meeting.generatedNote.chapter && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                const note = meeting.generatedNote!;
                const notebookId = note.chapter.notebook?.id || note.chapter.notebookId;
                const chapterId = note.chapterId;
                const noteId = note.id;
                navigate(`/${notebookId}/${chapterId}/${noteId}`);
              }}
              className="w-full mt-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors text-sm font-medium"
            >
              View Notes
            </button>
          )}
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
