import { useState, useEffect } from "react";
import { useUser } from "@/hooks/auth";
import { Header } from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Key, Trash2, User, Mail, Save } from "lucide-react";
import { toast } from "sonner";
import api from "@/utils/api";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarSettings } from "@/components/CalendarSettings";
import { CalendarEventsList } from "@/components/CalendarEventsList";

interface ApiKeyStatus {
  openai: boolean;
  anthropic: boolean;
  google: boolean;
}

export function Profile() {
  const { user, loading: userLoading, refetch: refetchUser } = useUser();
  const queryClient = useQueryClient();
  const [openAIKey, setOpenAIKey] = useState("");
  const [anthropicKey, setAnthropicKey] = useState("");
  const [googleKey, setGoogleKey] = useState("");
  const [isSavingKey, setIsSavingKey] = useState(false);
  const [isDeletingKey, setIsDeletingKey] = useState<string | null>(null);
  const [apiKeyStatus, setApiKeyStatus] = useState<ApiKeyStatus>({
    openai: false,
    anthropic: false,
    google: false,
  });
  const [calendarSyncTrigger, setCalendarSyncTrigger] = useState(0);

  // Fetch API key status on mount
  useEffect(() => {
    const fetchApiKeyStatus = async () => {
      try {
        const response = await api.get("/settings/ai-credentials");
        const providers = response.data.providers || {};
        setApiKeyStatus({
          openai: providers.openai || false,
          anthropic: providers.anthropic || false,
          google: providers.google || false,
        });
      } catch (error) {
        console.error("Failed to fetch API key status:", error);
      }
    };
    fetchApiKeyStatus();
  }, []);

  const handleSaveApiKey = async (provider: string, apiKey: string) => {
    if (!apiKey.trim()) {
      toast.error("Please enter an API key");
      return;
    }

    setIsSavingKey(true);
    // Optimistic update
    setApiKeyStatus((prev) => ({ ...prev, [provider]: true }));
    
    try {
      await api.post("/settings/ai-credentials", {
        provider,
        apiKey: apiKey.trim(),
      });
      toast.success(`${provider} API key saved successfully`);
      
      // Clear the input
      if (provider === "openai") setOpenAIKey("");
      if (provider === "anthropic") setAnthropicKey("");
      if (provider === "google") setGoogleKey("");
      
      // Refresh user data and API key status
      await refetchUser();
      const response = await api.get("/settings/ai-credentials");
      const providers = response.data.providers || {};
      setApiKeyStatus({
        openai: providers.openai || false,
        anthropic: providers.anthropic || false,
        google: providers.google || false,
      });
      // Invalidate API key status query so other components update
      queryClient.invalidateQueries({ queryKey: ['api-key-status'] });
    } catch (error: any) {
      console.error("Save API key error:", error);
      // Revert optimistic update on error
      setApiKeyStatus((prev) => ({ ...prev, [provider]: false }));
      const errorMessage = error.response?.data?.error || "Failed to save API key";
      toast.error(errorMessage);
    } finally {
      setIsSavingKey(false);
    }
  };

  const handleDeleteApiKey = async (provider: string) => {
    setIsDeletingKey(provider);
    // Optimistic update - immediately remove from UI
    const previousStatus = apiKeyStatus[provider as keyof ApiKeyStatus];
    setApiKeyStatus((prev) => ({ ...prev, [provider]: false }));
    
    try {
      await api.delete("/settings/ai-credentials", {
        data: { provider },
      });
      toast.success(`${provider} API key deleted successfully`);
      
      // Refresh user data and confirm API key status
      await refetchUser();
      const response = await api.get("/settings/ai-credentials");
      const providers = response.data.providers || {};
      setApiKeyStatus({
        openai: providers.openai || false,
        anthropic: providers.anthropic || false,
        google: providers.google || false,
      });
      // Invalidate API key status query so other components update
      queryClient.invalidateQueries({ queryKey: ['api-key-status'] });
    } catch (error: any) {
      console.error("Delete API key error:", error);
      // Revert optimistic update on error
      setApiKeyStatus((prev) => ({ ...prev, [provider]: previousStatus }));
      toast.error("Failed to delete API key");
    } finally {
      setIsDeletingKey(null);
    }
  };

  const handleOnboardingReset = async () => {
    await refetchUser();
    window.location.reload();
  };

  // Check for calendar OAuth callback
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    
    if (params.get('calendar_success')) {
      const provider = params.get('calendar_success');
      toast.success(`${provider === 'google' ? 'Google' : 'Microsoft'} calendar connected successfully! Syncing events...`);
      // Clear the query params
      window.history.replaceState({}, '', window.location.pathname);
      
      // Reload page after a short delay to show synced events
      setTimeout(() => {
        window.location.reload();
      }, 3000);
    }
    
    if (params.get('calendar_error')) {
      toast.error('Failed to connect calendar. Please try again.');
      // Clear the query params
      window.history.replaceState({}, '', window.location.pathname);
    }
  }, []);

  if (userLoading) {
    return (
      <div className="flex flex-col h-screen">
        <Header user={null} breadcrumbs={[{ label: "Profile" }]} />
        <main className="flex-1 overflow-auto">
          <div className="max-w-4xl mx-auto px-6 py-12 space-y-8">
            {/* Loading skeleton */}
            <div className="space-y-4">
              <Skeleton className="h-8 w-48" />
              <div className="bg-card border rounded-lg p-6 space-y-4">
                <div className="flex items-center gap-4">
                  <Skeleton className="w-16 h-16 rounded-full" />
                  <div className="space-y-2">
                    <Skeleton className="h-6 w-48" />
                    <Skeleton className="h-4 w-64" />
                  </div>
                </div>
              </div>
            </div>
            <Separator />
            <div className="space-y-4">
              <Skeleton className="h-8 w-32" />
              <div className="space-y-4">
                <Skeleton className="h-24 w-full rounded-lg" />
                <Skeleton className="h-24 w-full rounded-lg" />
                <Skeleton className="h-24 w-full rounded-lg" />
              </div>
            </div>
          </div>
        </main>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex flex-col h-screen">
        <Header user={user} breadcrumbs={[{ label: "Profile" }]} />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-muted-foreground">Please log in to view your profile</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={[{ label: "Profile" }]} onOnboardingReset={handleOnboardingReset} />
      
      <main className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto px-6 py-12 space-y-8">
          {/* User Information Section */}
          <div className="space-y-4">
            <h2 className="text-2xl font-semibold">User Information</h2>
            <div className="bg-card border rounded-lg p-6 space-y-4">
              <div className="flex items-center gap-4">
                {user.imageUrl ? (
                  <img
                    src={user.imageUrl}
                    alt={user.name}
                    className="w-16 h-16 rounded-full ring-2 ring-border"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-full bg-primary flex items-center justify-center ring-2 ring-border">
                    <User className="w-8 h-8 text-primary-foreground" />
                  </div>
                )}
                <div className="space-y-1">
                  <h3 className="text-lg font-semibold">{user.name}</h3>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Mail className="w-4 h-4" />
                    <span>{user.email}</span>
                  </div>
                </div>
              </div>
              
              <Separator />
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="text-sm font-medium text-muted-foreground">Onboarding Type</Label>
                  <div className="text-sm">
                    {user.onboardingCompleted ? (
                      <Badge variant="secondary">Completed</Badge>
                    ) : (
                      <Badge variant="outline">Not Completed</Badge>
                    )}
                  </div>
                </div>
                <div className="space-y-2">
                  <Label className="text-sm font-medium text-muted-foreground">API Keys Status</Label>
                  <div className="text-sm">
                    {user.hasApiKey ? (
                      <Badge variant="secondary" className="gap-1">
                        <Key className="w-3 h-3" />
                        Configured
                      </Badge>
                    ) : (
                      <Badge variant="outline">Not Configured</Badge>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* API Keys Section */}
          <div className="space-y-4">
            <div>
              <h2 className="text-2xl font-semibold mb-2">AI API Keys</h2>
              <p className="text-sm text-muted-foreground">
                Manage your API keys for different AI providers. Keys are encrypted and stored securely.
              </p>
            </div>

            <div className="bg-card border rounded-lg p-6 space-y-6">
              {/* OpenAI */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label htmlFor="openai-key" className="text-sm font-medium">
                    OpenAI API Key
                  </Label>
                  {apiKeyStatus.openai && (
                    <Badge variant="secondary" className="text-xs">
                      <Key className="w-3 h-3 mr-1" />
                      Configured
                    </Badge>
                  )}
                </div>
                <div className="flex gap-2">
                  <Input
                    id="openai-key"
                    type="password"
                    placeholder="sk-..."
                    value={openAIKey}
                    onChange={(e) => setOpenAIKey(e.target.value)}
                    className="flex-1"
                    disabled={isSavingKey}
                  />
                  <Button
                    onClick={() => handleSaveApiKey("openai", openAIKey)}
                    disabled={isSavingKey || !openAIKey.trim()}
                    size="sm"
                  >
                    {isSavingKey ? <Save className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                    Save
                  </Button>
                </div>
                {apiKeyStatus.openai && (
                  <div className="flex justify-between items-center p-3 bg-muted rounded-md">
                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">
                        API key configured for OpenAI
                      </span>
                    </div>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDeleteApiKey("openai")}
                      disabled={isDeletingKey === "openai"}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                )}
              </div>

              <Separator />

              {/* Anthropic */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label htmlFor="anthropic-key" className="text-sm font-medium">
                    Anthropic API Key
                  </Label>
                  {apiKeyStatus.anthropic && (
                    <Badge variant="secondary" className="text-xs">
                      <Key className="w-3 h-3 mr-1" />
                      Configured
                    </Badge>
                  )}
                </div>
                <div className="flex gap-2">
                  <Input
                    id="anthropic-key"
                    type="password"
                    placeholder="sk-ant-..."
                    value={anthropicKey}
                    onChange={(e) => setAnthropicKey(e.target.value)}
                    className="flex-1"
                    disabled={isSavingKey}
                  />
                  <Button
                    onClick={() => handleSaveApiKey("anthropic", anthropicKey)}
                    disabled={isSavingKey || !anthropicKey.trim()}
                    size="sm"
                  >
                    {isSavingKey ? <Save className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                    Save
                  </Button>
                </div>
                {apiKeyStatus.anthropic && (
                  <div className="flex justify-between items-center p-3 bg-muted rounded-md">
                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">
                        API key configured for Anthropic
                      </span>
                    </div>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDeleteApiKey("anthropic")}
                      disabled={isDeletingKey === "anthropic"}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                )}
              </div>

              <Separator />

              {/* Google */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label htmlFor="google-key" className="text-sm font-medium">
                    Google API Key
                  </Label>
                  {apiKeyStatus.google && (
                    <Badge variant="secondary" className="text-xs">
                      <Key className="w-3 h-3 mr-1" />
                      Configured
                    </Badge>
                  )}
                </div>
                <div className="flex gap-2">
                  <Input
                    id="google-key"
                    type="password"
                    placeholder="AIza..."
                    value={googleKey}
                    onChange={(e) => setGoogleKey(e.target.value)}
                    className="flex-1"
                    disabled={isSavingKey}
                  />
                  <Button
                    onClick={() => handleSaveApiKey("google", googleKey)}
                    disabled={isSavingKey || !googleKey.trim()}
                    size="sm"
                  >
                    {isSavingKey ? <Save className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                    Save
                  </Button>
                </div>
                {apiKeyStatus.google && (
                  <div className="flex justify-between items-center p-3 bg-muted rounded-md">
                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">
                        API key configured for Google
                      </span>
                    </div>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleDeleteApiKey("google")}
                      disabled={isDeletingKey === "google"}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                )}
              </div>

              {/* Security Notice */}
              <div className="p-3 rounded-md bg-muted/50 border border-border">
                <p className="text-xs text-muted-foreground">
                  <strong className="text-foreground">🔒 Secure Storage:</strong> Your API keys are encrypted using AES-256 encryption and stored securely on our servers. They are never exposed in plain text and only decrypted when needed for API calls.
                </p>
              </div>
            </div>
          </div>

          {/* Calendar Integration Section */}
          <CalendarSettings onSyncComplete={() => setCalendarSyncTrigger(prev => prev + 1)} />

          {/* Upcoming Meetings Section */}
          <CalendarEventsList key={calendarSyncTrigger} />
        </div>
      </main>
    </div>
  );
}
