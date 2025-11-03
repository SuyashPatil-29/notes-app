import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Calendar as CalendarIcon, Bot, BotOff, ExternalLink, RefreshCw, Video } from "lucide-react";
import { toast } from "sonner";
import type { Calendar, CalendarEvent } from "@/types/backend";
import {
  getUserCalendars,
  getCalendarEvents,
  scheduleBotForEvent,
  cancelBotForEvent,
  formatMeetingPlatform,
  getMeetingPlatformIcon,
} from "@/utils/calendar";

export function CalendarEventsList() {
  const navigate = useNavigate();
  const [calendars, setCalendars] = useState<Calendar[]>([]);
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSchedulingId, setIsSchedulingId] = useState<string | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const calendarsData = await getUserCalendars();
      setCalendars(calendarsData);

      // Fetch events from all calendars
      const allEvents: CalendarEvent[] = [];
      for (const calendar of calendarsData) {
        try {
          const calendarEvents = await getCalendarEvents(calendar.id, true);
          allEvents.push(...calendarEvents);
        } catch (error) {
          console.error(`Failed to fetch events for calendar ${calendar.id}:`, error);
        }
      }

      // Filter to only show meetings with meeting platform and URL (virtual meetings)
      const virtualMeetings = allEvents.filter(event => 
        event.meetingPlatform && event.meetingUrl
      );

      // Sort by start time
      virtualMeetings.sort((a, b) => 
        new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
      );

      setEvents(virtualMeetings);
    } catch (error) {
      console.error("Failed to fetch calendar data:", error);
      toast.error("Failed to load calendar events");
    } finally {
      setIsLoading(false);
    }
  };

  const handleScheduleBot = async (eventId: string) => {
    setIsSchedulingId(eventId);
    try {
      await scheduleBotForEvent(eventId);
      toast.success("Bot scheduled for meeting");
      // Refresh data from backend to ensure state is synchronized
      await fetchData();
    } catch (error: any) {
      console.error("Failed to schedule bot:", error);
      const errorMsg = error.response?.data?.error || "Failed to schedule bot";
      toast.error(errorMsg);
    } finally {
      setIsSchedulingId(null);
    }
  };

  const handleCancelBot = async (eventId: string) => {
    setIsSchedulingId(eventId);
    try {
      await cancelBotForEvent(eventId);
      toast.success("Bot cancelled");
      // Refresh data from backend to ensure state is synchronized
      await fetchData();
    } catch (error: any) {
      console.error("Failed to cancel bot:", error);
      const errorMsg = error.response?.data?.error || "Failed to cancel bot";
      toast.error(errorMsg);
    } finally {
      setIsSchedulingId(null);
    }
  };

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    const today = new Date();
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);

    const isToday = date.toDateString() === today.toDateString();
    const isTomorrow = date.toDateString() === tomorrow.toDateString();

    if (isToday) {
      return `Today at ${date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}`;
    } else if (isTomorrow) {
      return `Tomorrow at ${date.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}`;
    } else {
      return date.toLocaleString('en-US', { 
        month: 'short', 
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit'
      });
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Upcoming Meetings</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Virtual meetings that can be recorded by AI bots
          </p>
        </div>
        {calendars.length > 0 && (
        <Button
          variant="outline"
          size="sm"
          onClick={fetchData}
          disabled={isLoading}
        >
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
        </Button>
        )}
      </div>

      <div className="bg-card border rounded-lg p-6">
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">
            Loading events...
          </div>
        ) : calendars.length === 0 ? (
          <div className="text-center py-12 space-y-4">
            <Video className="w-16 h-16 mx-auto text-muted-foreground/30" />
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">No Calendar Connected</h3>
              <p className="text-muted-foreground max-w-md mx-auto">
                Connect your Google or Microsoft calendar to see upcoming virtual meetings and schedule AI bots for automatic recording.
              </p>
            </div>
            <Button
              onClick={() => navigate('/profile')}
              className="mt-4"
            >
              <CalendarIcon className="w-4 h-4 mr-2" />
              Connect Calendar
            </Button>
          </div>
        ) : events.length === 0 ? (
          <div className="text-center py-12 space-y-4">
            <CalendarIcon className="w-16 h-16 mx-auto text-muted-foreground/30" />
            <div className="space-y-2">
              <h3 className="text-lg font-semibold">No Upcoming Meetings</h3>
              <p className="text-muted-foreground max-w-md mx-auto">
                You don't have any upcoming virtual meetings. Zoom, Google Meet, Microsoft Teams, and other virtual meetings will appear here.
            </p>
            </div>
            <Button
              variant="outline"
              onClick={fetchData}
              className="mt-4"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            {events.map((event) => (
              <div
                key={event.id}
                className="flex items-start gap-4 p-4 bg-muted/30 rounded-lg border hover:bg-muted/50 transition-colors"
              >
                <div className="text-2xl pt-1 shrink-0">
                  {getMeetingPlatformIcon(event.meetingPlatform)}
                </div>

                <div className="flex-1 min-w-0 space-y-2">
                  <div>
                    <h3 className="font-medium truncate">
                      {event.title || 'Untitled Meeting'}
                    </h3>
                    <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                      <span>{formatDateTime(event.startTime)}</span>
                      <span>•</span>
                      <span>{formatMeetingPlatform(event.meetingPlatform)}</span>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 flex-wrap">
                    {event.botScheduled ? (
                      <>
                        <Badge variant="secondary" className="text-xs">
                          <Bot className="w-3 h-3 mr-1" />
                          Bot Scheduled
                        </Badge>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleCancelBot(event.id)}
                          disabled={isSchedulingId === event.id}
                          className="h-7 text-xs"
                        >
                          {isSchedulingId === event.id ? (
                            <RefreshCw className="w-3 h-3 mr-1 animate-spin" />
                          ) : (
                            <BotOff className="w-3 h-3 mr-1" />
                          )}
                          Cancel Bot
                        </Button>
                      </>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleScheduleBot(event.id)}
                        disabled={isSchedulingId === event.id}
                        className="h-7 text-xs"
                      >
                        {isSchedulingId === event.id ? (
                          <RefreshCw className="w-3 h-3 mr-1 animate-spin" />
                        ) : (
                          <Bot className="w-3 h-3 mr-1" />
                        )}
                        Schedule Bot
                      </Button>
                    )}

                    <Button
                      variant="ghost"
                      size="sm"
                      asChild
                      className="h-7 text-xs"
                    >
                      <a
                        href={event.meetingUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <ExternalLink className="w-3 h-3 mr-1" />
                        Join
                      </a>
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

