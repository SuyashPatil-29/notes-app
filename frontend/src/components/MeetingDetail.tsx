import { useQuery } from "@tanstack/react-query";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Loader2,
  ArrowLeft,
  FileText,
  Clock,
  Video,
  ExternalLink,
  XCircle,
} from "lucide-react";
import {
  getUserMeetings,
  getMeetingStatusLabel,
  getMeetingStatusVariant,
  getMeetingPlatform,
} from "@/utils/meeting";
import { MeetingTranscript } from "./MeetingTranscript";
import { formatDistanceToNow, format } from "date-fns";
import { Header } from "./Header";
import { useUser } from "@/hooks/auth";

export function MeetingDetail() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const { user } = useUser();
  const navigate = useNavigate();

  const { data: meetings, isLoading } = useQuery({
    queryKey: ["userMeetings"],
    queryFn: getUserMeetings,
    refetchInterval: (query) => {
      const meeting = query.state.data?.find(
        (m: any) => m.id === meetingId
      );
      const needsRefetch =
        meeting &&
        (meeting.status === "pending" ||
          meeting.status === "recording" ||
          meeting.status === "processing" ||
          (meeting.status === "completed" && !meeting.generatedNote));
      return needsRefetch ? 5000 : false;
    },
  });

  const meeting = meetings?.find((m) => m.id === meetingId);

  if (isLoading) {
    return (
      <>
        <Header user={user} />
        <div className="flex items-center justify-center min-h-[calc(100vh-64px)]">
          <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
        </div>
      </>
    );
  }

  if (!meeting) {
    return (
      <>
        <Header user={user} /> 
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <Button
            variant="ghost"
            onClick={() => navigate("/")}
            className="mb-6"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </Button>
          <div className="flex flex-col items-center justify-center p-12 text-center">
            <XCircle className="w-12 h-12 text-red-500 mb-4" />
            <h2 className="text-xl font-semibold mb-2">Meeting not found</h2>
            <p className="text-muted-foreground">
              This meeting may have been deleted or you don't have access to it.
            </p>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Header user={user} />
      <div className="container mx-auto px-4 py-8 max-w-5xl">
        {/* Back Button */}
        <Button
          variant="ghost"
          onClick={() => navigate("/")}
          className="mb-6"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>

        {/* Meeting Header */}
        <div className="mb-8">
          <div className="flex items-start justify-between gap-4 mb-4">
            <div className="flex-1">
              <h1 className="text-3xl font-bold mb-2">
                {meeting.generatedNote?.name ||
                  `${getMeetingPlatform(meeting.meetingUrl)} Meeting`}
              </h1>
              <div className="flex items-center gap-4 text-muted-foreground">
                <div className="flex items-center gap-2">
                  <Clock className="w-4 h-4" />
                  <span>
                    {formatDistanceToNow(new Date(meeting.createdAt), {
                      addSuffix: true,
                    })}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Video className="w-4 h-4" />
                  <span>{getMeetingPlatform(meeting.meetingUrl)}</span>
                </div>
              </div>
            </div>
            <Badge variant={getMeetingStatusVariant(meeting.status)} className="text-sm">
              {getMeetingStatusLabel(meeting.status)}
            </Badge>
          </div>

          {/* Meeting URL */}
          <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted/50 p-3 rounded-lg">
            <ExternalLink className="w-4 h-4" />
            <a
              href={meeting.meetingUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:underline font-mono"
            >
              {meeting.meetingUrl}
            </a>
          </div>
        </div>

        {/* AI Summary Section */}
        {meeting.generatedNote?.aiSummary && (
          <div className="mb-8 border rounded-lg p-6 bg-card">
            <h2 className="text-xl font-semibold mb-3 flex items-center gap-2">
              <FileText className="w-5 h-5" />
              AI Summary
            </h2>
            <p className="text-muted-foreground leading-relaxed">
              {meeting.generatedNote.aiSummary}
            </p>
          </div>
        )}

        {/* Action Buttons */}
        {meeting.generatedNote && meeting.generatedNote.chapter && (
          <div className="mb-8">
            <Button
              size="lg"
              className="w-full"
              onClick={() => {
                const note = meeting.generatedNote!;
                const notebookId =
                  note.chapter.notebook?.id || note.chapter.notebookId;
                const chapterId = note.chapterId;
                const noteId = note.id;
                navigate(`/${notebookId}/${chapterId}/${noteId}`);
              }}
            >
              <FileText className="w-5 h-5 mr-2" />
              View Full Notes
            </Button>
          </div>
        )}

        {/* Status Messages */}
        {meeting.status === "failed" && (
          <div className="mb-8 border border-red-200 dark:border-red-800 rounded-lg p-6 bg-red-50 dark:bg-red-950/20">
            <div className="flex items-center gap-3 text-red-600 dark:text-red-400">
              <XCircle className="w-5 h-5" />
              <div>
                <h3 className="font-semibold mb-1">Processing Failed</h3>
                <p className="text-sm">
                  There was an error processing this meeting. Please try again or contact support.
                </p>
              </div>
            </div>
          </div>
        )}

        {(meeting.status === "pending" ||
          meeting.status === "recording" ||
          meeting.status === "processing") && (
            <div className="mb-8 border border-blue-200 dark:border-blue-800 rounded-lg p-6 bg-blue-50 dark:bg-blue-950/20">
              <div className="flex items-center gap-3 text-blue-600 dark:text-blue-400">
                <Loader2 className="w-5 h-5 animate-spin" />
                <div>
                  <h3 className="font-semibold mb-1">
                    {meeting.status === "recording"
                      ? "Recording in Progress"
                      : meeting.status === "processing"
                        ? "Processing Transcript"
                        : "Preparing Bot"}
                  </h3>
                  <p className="text-sm">
                    {meeting.status === "recording"
                      ? "The bot is currently in your meeting and recording the conversation."
                      : meeting.status === "processing"
                        ? "We're processing the transcript and generating your notes. This may take a few minutes."
                        : "The bot is preparing to join your meeting."}
                  </p>
                </div>
              </div>
            </div>
          )}

        {/* Video Recording Section */}
        {meeting.status === "completed" && meeting.videoDownloadUrl && (
          <div className="mb-8 border rounded-lg overflow-hidden bg-card">
            <div className="p-4 border-b bg-muted/50">
              <h2 className="text-xl font-semibold flex items-center gap-2">
                <Video className="w-5 h-5" />
                Meeting Recording
              </h2>
            </div>
            <div className="p-4">
              <video
                controls
                className="w-full rounded-lg bg-black"
                style={{ maxHeight: "600px" }}
                preload="metadata"
              >
                <source src={meeting.videoDownloadUrl} type="video/mp4" />
                Your browser does not support the video player.
              </video>
              <p className="text-sm text-muted-foreground mt-3 text-center">
                Full meeting recording • {getMeetingPlatform(meeting.meetingUrl)}
              </p>
            </div>
          </div>
        )}

        {/* Transcript Section */}
        {meeting.status === "completed" && (
          <MeetingTranscript meetingId={meeting.id} />
        )}

        {/* Meeting Metadata */}
        {meeting.completedAt && (
          <div className="mt-6 text-sm text-muted-foreground text-center">
            Completed on {format(new Date(meeting.completedAt), "PPpp")}
          </div>
        )}
      </div>
    </>
  );
}