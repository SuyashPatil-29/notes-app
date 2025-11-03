import { useState, useEffect } from "react";
import { useUser } from "@/hooks/auth";
import { Header } from "@/components/Header";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Key, Trash2, User as UserIcon, Mail, Save, Calendar, Video, Sparkles, Check, X, Building2 } from "lucide-react";
import { toast } from "sonner";
import api from "@/utils/api";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarSettings } from "@/components/CalendarSettings";
import { CalendarEventsList } from "@/components/CalendarEventsList";
import { useOrganizationContext } from "@/contexts/OrganizationContext";
import { OrganizationSettings } from "@/components/organizations/OrganizationSettings";
import { OrganizationsMemberView } from "@/components/organizations/OrganizationsMemberView";
import { JoinOrganizationDialog } from "@/components/organizations/JoinOrganizationDialog";
import { useOrganization } from "@clerk/clerk-react";

interface ApiKeyStatus {
  openai: boolean;
  anthropic: boolean;
  google: boolean;
}

const AI_PROVIDERS = [
  {
    id: "openai",
    name: "OpenAI",
    placeholder: "sk-...",
    description: "GPT-4, GPT-3.5 models"
  },
  {
    id: "anthropic",
    name: "Anthropic",
    placeholder: "sk-ant-...",
    description: "Claude 3 Opus, Sonnet, Haiku"
  },
  {
    id: "google",
    name: "Google AI",
    placeholder: "AIza...",
    description: "Gemini Pro models"
  }
];

export function Profile() {
  const { user, loading: userLoading, refetch: refetchUser } = useUser();
  const queryClient = useQueryClient();
  const { organizations, isLoadingOrgs, activeOrg } = useOrganizationContext();
  const { membership } = useOrganization();
  const [activeTab, setActiveTab] = useState<'profile' | 'ai-keys' | 'calendar' | 'meetings' | 'organizations'>('profile');
  const [isJoinDialogOpen, setIsJoinDialogOpen] = useState(false);
  const [apiKeys, setApiKeys] = useState<Record<string, string>>({
    openai: "",
    anthropic: "",
    google: ""
  });
  const [isSavingKey, setIsSavingKey] = useState<string | null>(null);
  const [isDeletingKey, setIsDeletingKey] = useState<string | null>(null);
  const [apiKeyStatus, setApiKeyStatus] = useState<ApiKeyStatus>({
    openai: false,
    anthropic: false,
    google: false,
  });
  const [calendarSyncTrigger, setCalendarSyncTrigger] = useState(0);
  const [hasCalendar, setHasCalendar] = useState(false);

  const isAdmin = membership?.role === 'org:admin';

  // Fetch API key status and calendar status on mount
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
    
    const fetchCalendarStatus = async () => {
      try {
        const response = await api.get("/api/calendars");
        setHasCalendar(response.data.calendars?.length > 0);
      } catch (error) {
        console.error("Failed to fetch calendar status:", error);
      }
    };
    
    fetchApiKeyStatus();
    fetchCalendarStatus();
  }, []);

  const handleSaveApiKey = async (provider: string) => {
    const apiKey = apiKeys[provider];
    if (!apiKey.trim()) {
      toast.error("Please enter an API key");
      return;
    }

    setIsSavingKey(provider);
    // Optimistic update
    setApiKeyStatus((prev) => ({ ...prev, [provider]: true }));
    
    try {
      await api.post("/settings/ai-credentials", {
        provider,
        apiKey: apiKey.trim(),
      });
      toast.success(`${provider} API key saved successfully`);
      
      // Clear the input
      setApiKeys(prev => ({ ...prev, [provider]: "" }));
      
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
      setIsSavingKey(null);
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
      
      // Update calendar status
      setHasCalendar(true);
      // Trigger a calendar sync refresh without full page reload
      setCalendarSyncTrigger(prev => prev + 1);
      // Switch to calendar tab
      setActiveTab('calendar');
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
          <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
            {/* Loading skeleton */}
            <div className="space-y-2">
              <Skeleton className="h-9 w-64" />
              <Skeleton className="h-5 w-96" />
            </div>
            <div className="h-12" />
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-40 rounded-lg" />
              ))}
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

  const configuredKeysCount = Object.values(apiKeyStatus).filter(Boolean).length;

  return (
    <div className="flex flex-col h-screen">
      <Header user={user} breadcrumbs={[{ label: "Profile" }]} onOnboardingReset={handleOnboardingReset} />
      
      <main className="flex-1 overflow-auto">
        <div className="max-w-6xl mx-auto px-6 py-12 space-y-8">
          {/* Welcome Section */}
          <div className="space-y-2">
            <h2 className="text-3xl font-bold text-foreground">
              Profile & Settings
            </h2>
            <p className="text-muted-foreground">
              Manage your account, API keys, and integrations
            </p>
          </div>

          {/* Tabs */}
          <div className="flex items-center gap-1 border-b border-border">
            <button
              onClick={() => setActiveTab('profile')}
              className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                activeTab === 'profile'
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <UserIcon className="h-4 w-4" />
              Profile
              {activeTab === 'profile' && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
              )}
            </button>
            
            {/* Show these tabs only if: in personal account OR in org as admin */}
            {(!activeOrg || isAdmin) && (
              <>
                <button
                  onClick={() => setActiveTab('ai-keys')}
                  className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                    activeTab === 'ai-keys'
                      ? 'text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Sparkles className="h-4 w-4" />
                  AI Keys
                  {configuredKeysCount > 0 && (
                    <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
                      {configuredKeysCount}
                    </Badge>
                  )}
                  {activeTab === 'ai-keys' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                  )}
                </button>
                <button
                  onClick={() => setActiveTab('calendar')}
                  className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                    activeTab === 'calendar'
                      ? 'text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Calendar className="h-4 w-4" />
                  Calendar
                  {activeTab === 'calendar' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                  )}
                </button>
                <button
                  onClick={() => setActiveTab('meetings')}
                  className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                    activeTab === 'meetings'
                      ? 'text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Video className="h-4 w-4" />
                  Meetings
                  {activeTab === 'meetings' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                  )}
                </button>
              </>
            )}
            
            {/* Organizations tab - always visible for all users */}
                <button
                  onClick={() => setActiveTab('organizations')}
                  className={`flex items-center gap-2 px-4 py-3 font-medium transition-colors relative ${
                    activeTab === 'organizations'
                      ? 'text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Building2 className="h-4 w-4" />
                  Organizations
                  {!isLoadingOrgs && organizations.length > 0 && (
                    <Badge variant="secondary" className="ml-1 h-5 px-1.5 text-xs">
                      {organizations.length}
                    </Badge>
                  )}
                  {activeTab === 'organizations' && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t-lg" />
                  )}
                </button>
          </div>

          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="space-y-6">
              {/* User Card */}
              <div className="bg-card border border-border rounded-lg p-6">
                <div className="flex items-start gap-6">
                  {user.imageUrl ? (
                    <img
                      src={user.imageUrl}
                      alt={user.name}
                      className="w-20 h-20 rounded-full ring-2 ring-border"
                    />
                  ) : (
                    <div className="w-20 h-20 rounded-full bg-primary flex items-center justify-center ring-2 ring-border">
                      <UserIcon className="w-10 h-10 text-primary-foreground" />
                    </div>
                  )}
                  <div className="flex-1 space-y-4">
                    <div>
                      <h3 className="text-2xl font-semibold mb-2">{user.name}</h3>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Mail className="w-4 h-4" />
                        <span>{user.email}</span>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        {user.onboardingCompleted ? (
                          <>
                            <Check className="w-4 h-4 text-green-500" />
                            <span className="text-sm text-muted-foreground">Onboarding Complete</span>
                          </>
                        ) : (
                          <>
                            <X className="w-4 h-4 text-muted-foreground" />
                            <span className="text-sm text-muted-foreground">Onboarding Incomplete</span>
                          </>
                        )}
                      </div>
                      <Separator orientation="vertical" className="h-4" />
                      <div className="flex items-center gap-2">
                        {configuredKeysCount > 0 ? (
                          <>
                            <Key className="w-4 h-4 text-primary" />
                            <span className="text-sm text-muted-foreground">{configuredKeysCount} API Key{configuredKeysCount !== 1 ? 's' : ''} Configured</span>
                          </>
                        ) : (
                          <>
                            <Key className="w-4 h-4 text-muted-foreground" />
                            <span className="text-sm text-muted-foreground">No API Keys</span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Organization Member Notice */}
              {activeOrg && !isAdmin && (
                <div className="bg-muted/30 border border-border rounded-lg p-6">
                  <div className="flex items-start gap-3">
                    <Building2 className="h-5 w-5 text-muted-foreground mt-0.5" />
                    <div>
                      <h4 className="font-medium mb-1">Organization Member</h4>
                      <p className="text-sm text-muted-foreground">
                        You're currently viewing <strong className="text-foreground">{activeOrg.name}</strong> as a member. 
                        Additional settings are only available to organization administrators.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Quick Actions - Only show if in personal account or org admin */}
              {(!activeOrg || isAdmin) && (
                <div>
                  <h3 className="text-lg font-semibold mb-4">Quick Actions</h3>
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    <Button
                      variant="outline"
                      className="h-auto py-6 justify-start"
                      onClick={() => setActiveTab('ai-keys')}
                    >
                      <div className="flex items-start gap-3 text-left">
                        <Sparkles className="w-5 h-5 mt-0.5 text-primary" />
                        <div>
                          <div className="font-medium">Configure AI Keys</div>
                          <div className="text-xs text-muted-foreground mt-1">Set up OpenAI, Anthropic, or Google AI</div>
                        </div>
                      </div>
                    </Button>
                    
                    <Button
                      variant="outline"
                      className="h-auto py-6 justify-start"
                      onClick={() => setActiveTab('calendar')}
                    >
                      <div className="flex items-start gap-3 text-left">
                        <Calendar className={`w-5 h-5 mt-0.5 ${hasCalendar ? 'text-green-500' : 'text-primary'}`} />
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            {hasCalendar ? 'Calendar Connected' : 'Connect Calendar'}
                            {hasCalendar && <Check className="w-4 h-4 text-green-500" />}
                          </div>
                          <div className="text-xs text-muted-foreground mt-1">
                            {hasCalendar ? 'Manage calendar integration' : 'Enable automatic meeting recording'}
                          </div>
                        </div>
                      </div>
                    </Button>
                    
                    <Button
                      variant="outline"
                      className="h-auto py-6 justify-start"
                      onClick={() => setActiveTab('meetings')}
                    >
                      <div className="flex items-start gap-3 text-left">
                        <Video className="w-5 h-5 mt-0.5 text-primary" />
                        <div>
                          <div className="font-medium">View Meetings</div>
                          <div className="text-xs text-muted-foreground mt-1">Upcoming virtual meetings</div>
                        </div>
                      </div>
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* AI Keys Tab - Only accessible in personal account or as org admin */}
          {activeTab === 'ai-keys' && (!activeOrg || isAdmin) && (
            <div className="space-y-6">
              <div className="p-4 rounded-lg bg-muted/30 border border-border">
                <p className="text-sm text-muted-foreground">
                  <strong className="text-foreground">🔒 Secure Storage:</strong> Your API keys are encrypted using AES-256 encryption and stored securely. They are never exposed in plain text and only decrypted when needed for API calls.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-1">
                {AI_PROVIDERS.map((provider) => {
                  const isConfigured = apiKeyStatus[provider.id as keyof ApiKeyStatus];
                  const isSaving = isSavingKey === provider.id;
                  const isDeleting = isDeletingKey === provider.id;
                  
                  return (
                    <div
                      key={provider.id}
                      className="bg-card border border-border rounded-lg p-6 hover:border-primary/30 transition-colors"
                    >
                      <div className="flex items-start gap-4">
                        <div className="flex-1 space-y-4">
                          <div className="flex items-start justify-between">
                            <div>
                              <div className="flex items-center gap-2 mb-1">
                                <h3 className="text-lg font-semibold">{provider.name}</h3>
                                {isConfigured && (
                                  <Badge variant="secondary" className="text-xs">
                                    <Key className="w-3 h-3 mr-1" />
                                    Configured
                                  </Badge>
                                )}
                              </div>
                              <p className="text-sm text-muted-foreground">{provider.description}</p>
                            </div>
                          </div>

                          {!isConfigured ? (
                            <div className="flex gap-2">
                              <Input
                                type="password"
                                placeholder={provider.placeholder}
                                value={apiKeys[provider.id] || ""}
                                onChange={(e) => setApiKeys(prev => ({ ...prev, [provider.id]: e.target.value }))}
                                className="flex-1"
                                disabled={isSaving}
                              />
                              <Button
                                onClick={() => handleSaveApiKey(provider.id)}
                                disabled={isSaving || !apiKeys[provider.id]?.trim()}
                                size="sm"
                              >
                                {isSaving ? (
                                  <Save className="w-4 h-4 animate-spin" />
                                ) : (
                                  <>
                                    <Save className="w-4 h-4 mr-2" />
                                    Save
                                  </>
                                )}
                              </Button>
                            </div>
                          ) : (
                            <div className="flex items-center justify-between p-3 bg-muted/50 rounded-md">
                              <div className="flex items-center gap-2">
                                <Check className="w-4 h-4 text-green-500" />
                                <span className="text-sm text-muted-foreground">
                                  API key is configured and ready to use
                                </span>
                              </div>
                              <Button
                                variant="destructive"
                                size="sm"
                                onClick={() => handleDeleteApiKey(provider.id)}
                                disabled={isDeleting}
                              >
                                {isDeleting ? (
                                  <Trash2 className="w-4 h-4 animate-spin" />
                                ) : (
                                  <>
                                    <Trash2 className="w-4 h-4 mr-2" />
                                    Delete
                                  </>
                                )}
                              </Button>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Calendar Tab - Only accessible in personal account or as org admin */}
          {activeTab === 'calendar' && (!activeOrg || isAdmin) && (
            <div className="space-y-6">
              <CalendarSettings onSyncComplete={() => setCalendarSyncTrigger(prev => prev + 1)} />
            </div>
          )}

          {/* Meetings Tab - Only accessible in personal account or as org admin */}
          {activeTab === 'meetings' && (!activeOrg || isAdmin) && (
            <div className="space-y-6">
              <CalendarEventsList key={calendarSyncTrigger} />
            </div>
          )}

          {/* Organizations Tab - Always accessible for all users */}
          {activeTab === 'organizations' && (
            <div className="space-y-6">
              {activeOrg && isAdmin ? (
                // Show admin controls if user is admin of active org
                <OrganizationSettings
                  organization={activeOrg}
                  userRole="admin"
                />
              ) : (
                // Show member view for: members in org OR anyone in personal account
                <OrganizationsMemberView />
              )}
            </div>
          )}
        </div>
      </main>

      {/* Join Organization Dialog */}
      <JoinOrganizationDialog
        open={isJoinDialogOpen}
        onOpenChange={setIsJoinDialogOpen}
      />
    </div>
  );
}
