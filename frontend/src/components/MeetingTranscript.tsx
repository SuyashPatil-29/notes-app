import { useQuery } from "@tanstack/react-query";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Loader2,
  MessageSquare,
  Download,
  Copy,
  Check,
  Clock,
} from "lucide-react";
import { getMeetingTranscript } from "@/utils/meeting";
import { formatDistanceToNow, format } from "date-fns";
import { useState } from "react";

interface MeetingTranscriptProps {
  meetingId: string;
}

interface TranscriptEntry {
  speaker: string;
  text: string;
  timestamp: string;
}

export function MeetingTranscript({ meetingId }: MeetingTranscriptProps) {
  const [copied, setCopied] = useState(false);

  const {
    data: transcriptData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["meetingTranscript", meetingId],
    queryFn: () => getMeetingTranscript(meetingId),
    enabled: !!meetingId,
  });

  const copyTranscript = async () => {
    if (!transcriptData?.transcript) return;

    const text = transcriptData.transcript
      .map((entry: TranscriptEntry) => `${entry.speaker}: ${entry.text}`)
      .join("\n\n");

    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const downloadTranscript = () => {
    if (!transcriptData?.transcript) return;

    const text = transcriptData.transcript
      .map(
        (entry: TranscriptEntry) =>
          `[${format(new Date(entry.timestamp), "HH:mm:ss")}] ${entry.speaker}: ${entry.text}`
      )
      .join("\n\n");

    const blob = new Blob([text], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `meeting-transcript-${meetingId}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-8">
          <Loader2 className="w-6 h-6 animate-spin" />
          <span className="ml-2">Loading transcript...</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center p-8 text-center">
          <MessageSquare className="w-12 h-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">
            Failed to load transcript
          </p>
          <p className="text-sm text-muted-foreground mt-1">
            {(error as Error).message}
          </p>
        </CardContent>
      </Card>
    );
  }

  if (!transcriptData?.transcript || transcriptData.transcript.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center p-8 text-center">
          <MessageSquare className="w-12 h-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">
            {transcriptData?.status === "completed"
              ? "No transcript available"
              : "Transcript not yet available"}
          </p>
          <p className="text-sm text-muted-foreground mt-1">
            {transcriptData?.status === "pending" &&
              "The bot is preparing to join the meeting"}
            {transcriptData?.status === "recording" &&
              "Recording in progress..."}
            {transcriptData?.status === "processing" &&
              "Processing transcript..."}
          </p>
        </CardContent>
      </Card>
    );
  }

  const transcript = transcriptData.transcript as TranscriptEntry[];

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2">
              <MessageSquare className="w-5 h-5" />
              Meeting Transcript
            </CardTitle>
            <CardDescription className="mt-2">
              <div className="flex items-center gap-2 text-xs">
                <Clock className="w-3 h-3" />
                {transcriptData.createdAt && (
                  <span>
                    {formatDistanceToNow(new Date(transcriptData.createdAt), {
                      addSuffix: true,
                    })}
                  </span>
                )}
              </div>
            </CardDescription>
          </div>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={copyTranscript}
              disabled={copied}
            >
              {copied ? (
                <>
                  <Check className="w-4 h-4 mr-2" />
                  Copied
                </>
              ) : (
                <>
                  <Copy className="w-4 h-4 mr-2" />
                  Copy
                </>
              )}
            </Button>
            <Button size="sm" variant="outline" onClick={downloadTranscript}>
              <Download className="w-4 h-4 mr-2" />
              Download
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4 max-h-[70vh] overflow-y-auto rounded-md border p-4">
          {transcript.map((entry, index) => (
            <div
              key={index}
              className="flex gap-3 pb-4 border-b last:border-b-0 last:pb-0"
            >
              <div className="shrink-0">
                <Badge variant="outline" className="font-mono text-xs">
                  {entry.timestamp &&
                    format(new Date(entry.timestamp), "HH:mm:ss")}
                </Badge>
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-sm mb-1">
                  {entry.speaker}
                </div>
                <p className="text-sm text-muted-foreground leading-relaxed">
                  {entry.text}
                </p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

