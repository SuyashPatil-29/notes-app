import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Calendar as CalendarIcon, Trash2, RefreshCw, CheckCircle2, Download } from "lucide-react";
import { toast } from "sonner";
import type { Calendar } from "@/types/backend";
import {
  getUserCalendars,
  disconnectCalendar,
  initiateCalendarAuth,
  syncCalendar,
  syncMissingCalendars,
  formatPlatformName,
  getPlatformIcon,
} from "@/utils/calendar";

interface CalendarSettingsProps {
  onSyncComplete?: () => void;
}

export function CalendarSettings({ onSyncComplete }: CalendarSettingsProps = {}) {
  const [calendars, setCalendars] = useState<Calendar[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isConnecting, setIsConnecting] = useState<string | null>(null);
  const [isDeletingId, setIsDeletingId] = useState<string | null>(null);
  const [isSyncingId, setIsSyncingId] = useState<string | null>(null);
  const [isSyncingMissing, setIsSyncingMissing] = useState(false);

  useEffect(() => {
    fetchCalendars();
  }, []);

  const fetchCalendars = async () => {
    try {
      const data = await getUserCalendars();
      setCalendars(data);
    } catch (error) {
      console.error("Failed to fetch calendars:", error);
      toast.error("Failed to load calendars");
    } finally {
      setIsLoading(false);
    }
  };

  const handleConnectCalendar = async (provider: 'google' | 'microsoft') => {
    setIsConnecting(provider);
    try {
      await initiateCalendarAuth(provider);
      // User will be redirected to OAuth flow
    } catch (error: any) {
      console.error("Failed to initiate calendar auth:", error);
      toast.error("Failed to connect calendar");
      setIsConnecting(null);
    }
  };

  const handleDisconnect = async (calendarId: string, platform: string) => {
    setIsDeletingId(calendarId);
    try {
      await disconnectCalendar(calendarId);
      setCalendars(calendars.filter(cal => cal.id !== calendarId));
      toast.success(`${formatPlatformName(platform)} disconnected`);
    } catch (error) {
      console.error("Failed to disconnect calendar:", error);
      toast.error("Failed to disconnect calendar");
    } finally {
      setIsDeletingId(null);
    }
  };

  const handleSync = async (calendarId: string) => {
    setIsSyncingId(calendarId);
    try {
      const result = await syncCalendar(calendarId);
      toast.success(`Synced ${result.syncedCount} events`);
      // Refresh calendars to update lastSyncedAt
      await fetchCalendars();
      // Trigger events list refresh
      if (onSyncComplete) {
        onSyncComplete();
      }
    } catch (error) {
      console.error("Failed to sync calendar:", error);
      toast.error("Failed to sync calendar");
    } finally {
      setIsSyncingId(null);
    }
  };

  const handleSyncMissingCalendars = async () => {
    setIsSyncingMissing(true);
    try {
      const result = await syncMissingCalendars();
      if (result.added > 0) {
        toast.success(`Added ${result.added} missing calendar${result.added > 1 ? 's' : ''}`);
        // Refresh calendars to show newly added ones
        await fetchCalendars();
        // Trigger events list refresh
        if (onSyncComplete) {
          onSyncComplete();
        }
      } else {
        toast.info("All calendars are already synced");
      }
    } catch (error) {
      console.error("Failed to sync missing calendars:", error);
      toast.error("Failed to sync missing calendars");
    } finally {
      setIsSyncingMissing(false);
    }
  };

  const hasGoogleCalendar = calendars.some(cal => cal.platform === 'google_calendar');

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold mb-2">Calendar Integration</h2>
        <p className="text-sm text-muted-foreground">
          Connect your calendar to automatically record and transcribe meetings with AI bots.
        </p>
      </div>

      <div className="bg-card border rounded-lg p-6 space-y-6">
        {/* Connect Buttons */}
        <div className="space-y-3">
          <div className="flex gap-2 flex-wrap">
            <Button
              onClick={() => handleConnectCalendar('google')}
              disabled={hasGoogleCalendar || isConnecting === 'google'}
              className="flex-1 min-w-[200px]"
            >
              {isConnecting === 'google' ? (
                <>
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                  Connecting...
                </>
              ) : hasGoogleCalendar ? (
                <>
                  <CheckCircle2 className="w-4 h-4 mr-2" />
                  Google Calendar Connected
                </>
              ) : (
                <>
                  📅 Connect Google Calendar
                </>
              )}
            </Button>

            <Button
              disabled={true}
              variant="outline"
              className="flex-1 min-w-[200px] opacity-60 cursor-not-allowed"
            >
              📆 Microsoft Outlook
              <span className="ml-2 text-xs bg-muted px-2 py-0.5 rounded">Coming Soon</span>
            </Button>
          </div>
        </div>

        {/* Connected Calendars List */}
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">
            Loading calendars...
          </div>
        ) : calendars.length === 0 ? (
          <div className="text-center py-8 space-y-2">
            <CalendarIcon className="w-12 h-12 mx-auto text-muted-foreground/50" />
            <p className="text-muted-foreground">No calendars connected yet</p>
            <p className="text-xs text-muted-foreground">
              Connect your calendar to enable automatic meeting recording
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            <Separator />
            
            {/* Sync Missing Calendars Button */}
            <div className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border">
              <div className="flex-1">
                <p className="text-sm font-medium">Missing Calendars?</p>
                <p className="text-xs text-muted-foreground">
                  Sync all calendars from Recall if some are not showing up
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleSyncMissingCalendars}
                disabled={isSyncingMissing}
              >
                {isSyncingMissing ? (
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <Download className="w-4 h-4 mr-2" />
                )}
                Sync Missing
              </Button>
            </div>
            
            <div className="space-y-3">
              {calendars.map((calendar) => (
                <div
                  key={calendar.id}
                  className="flex items-center justify-between p-4 bg-muted/50 rounded-lg border"
                >
                  <div className="flex items-center gap-3 flex-1">
                    <div className="text-2xl">
                      {getPlatformIcon(calendar.platform)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-medium">
                          {formatPlatformName(calendar.platform)}
                        </h3>
                        <Badge 
                          variant={calendar.status === 'active' ? 'secondary' : 'destructive'}
                          className="text-xs"
                        >
                          {calendar.status}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground truncate">
                        {calendar.platformEmail}
                      </p>
                      {calendar.lastSyncedAt && (
                        <p className="text-xs text-muted-foreground mt-1">
                          Last synced: {new Date(calendar.lastSyncedAt).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleSync(calendar.id)}
                      disabled={isSyncingId === calendar.id}
                    >
                      {isSyncingId === calendar.id ? (
                        <RefreshCw className="w-4 h-4 animate-spin" />
                      ) : (
                        <RefreshCw className="w-4 h-4" />
                      )}
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDisconnect(calendar.id, calendar.platform)}
                      disabled={isDeletingId === calendar.id}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

