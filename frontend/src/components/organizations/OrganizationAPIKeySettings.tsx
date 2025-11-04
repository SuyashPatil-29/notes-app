import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Key, Trash2, Save, Loader2, Eye, EyeOff, AlertCircle } from "lucide-react";
import { toast } from "sonner";
import api from "@/utils/api";
import { useQueryClient } from "@tanstack/react-query";
import type { OrganizationAPIKeyStatus, SetOrgAPICredentialRequest, DeleteOrgAPICredentialRequest } from "@/types/backend";

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

interface OrganizationAPIKeySettingsProps {
  organizationId: string;
  isAdmin: boolean;
}

export function OrganizationAPIKeySettings({ organizationId, isAdmin }: OrganizationAPIKeySettingsProps) {
  const queryClient = useQueryClient();
  const [credentials, setCredentials] = useState<OrganizationAPIKeyStatus[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [apiKeys, setApiKeys] = useState<Record<string, string>>({
    openai: "",
    anthropic: "",
    google: ""
  });
  const [showKeys, setShowKeys] = useState<Record<string, boolean>>({
    openai: false,
    anthropic: false,
    google: false
  });
  const [isSavingKey, setIsSavingKey] = useState<string | null>(null);
  const [isDeletingKey, setIsDeletingKey] = useState<string | null>(null);

  useEffect(() => {
    fetchCredentials();
  }, [organizationId]);

  const fetchCredentials = async () => {
    if (!organizationId) return;
    
    setIsLoading(true);
    try {
      const response = await api.get(`/organizations/${organizationId}/api-credentials`);
      // Backend wraps response in a "data" field
      setCredentials(response.data.data?.credentials || []);
    } catch (error: any) {
      console.error("Failed to fetch organization API credentials:", error);
      toast.error(error.response?.data?.error || "Failed to load API credentials");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveApiKey = async (provider: string) => {
    const apiKey = apiKeys[provider];
    if (!apiKey.trim()) {
      toast.error("Please enter an API key");
      return;
    }

    setIsSavingKey(provider);
    
    try {
      const request: SetOrgAPICredentialRequest = {
        provider,
        apiKey: apiKey.trim(),
      };
      
      await api.post(`/organizations/${organizationId}/api-credentials`, request);
      toast.success(`${provider} API key saved successfully`);
      
      // Clear the input
      setApiKeys(prev => ({ ...prev, [provider]: "" }));
      setShowKeys(prev => ({ ...prev, [provider]: false }));
      
      // Refresh credentials
      await fetchCredentials();
      
      // Invalidate API key status queries so other components (like chat) update
      queryClient.invalidateQueries({ queryKey: ['api-key-status'] });
    } catch (error: any) {
      console.error("Save API key error:", error);
      const errorMessage = error.response?.data?.error || "Failed to save API key";
      toast.error(errorMessage);
    } finally {
      setIsSavingKey(null);
    }
  };

  const handleDeleteApiKey = async (provider: string) => {
    setIsDeletingKey(provider);
    
    try {
      const request: DeleteOrgAPICredentialRequest = { provider };
      await api.delete(`/organizations/${organizationId}/api-credentials`, { data: request });
      toast.success(`${provider} API key deleted successfully`);
      
      // Refresh credentials
      await fetchCredentials();
      
      // Invalidate API key status queries so other components (like chat) update
      queryClient.invalidateQueries({ queryKey: ['api-key-status'] });
    } catch (error: any) {
      console.error("Delete API key error:", error);
      const errorMessage = error.response?.data?.error || "Failed to delete API key";
      toast.error(errorMessage);
    } finally {
      setIsDeletingKey(null);
    }
  };

  const toggleShowKey = (provider: string) => {
    setShowKeys(prev => ({ ...prev, [provider]: !prev[provider] }));
  };

  const getCredentialForProvider = (provider: string): OrganizationAPIKeyStatus | undefined => {
    return credentials.find(cred => cred.provider === provider);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Organization API Keys</CardTitle>
            <CardDescription>
              Configure AI provider API keys for your organization
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            Organization API Keys
          </CardTitle>
          <CardDescription>
            Configure AI provider API keys for your organization. These keys will be used by all organization members for AI requests.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {!isAdmin && (
            <div className="flex items-center gap-2 p-3 bg-muted rounded-lg">
              <AlertCircle className="h-4 w-4 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">
                Only organization admins can manage API keys
              </p>
            </div>
          )}

          {AI_PROVIDERS.map((provider) => {
            const credential = getCredentialForProvider(provider.id);
            const hasKey = credential?.hasKey || false;
            const isSaving = isSavingKey === provider.id;
            const isDeleting = isDeletingKey === provider.id;
            const showKey = showKeys[provider.id];

            return (
              <div key={provider.id} className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium flex items-center gap-2">
                      {provider.name}
                      {hasKey && <Badge variant="secondary">Configured</Badge>}
                    </h4>
                    <p className="text-sm text-muted-foreground">
                      {provider.description}
                    </p>
                  </div>
                </div>

                {hasKey && credential && (
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <Label className="text-sm font-medium">Current Key:</Label>
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {credential.maskedKey || "****"}
                      </code>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Added by {credential.createdBy} on {new Date(credential.createdAt).toLocaleDateString()}
                    </div>
                  </div>
                )}

                {isAdmin && (
                  <div className="space-y-3">
                    <div className="flex gap-2">
                      <div className="flex-1">
                        <Input
                          type={showKey ? "text" : "password"}
                          placeholder={hasKey ? "Enter new API key to replace" : provider.placeholder}
                          value={apiKeys[provider.id] || ""}
                          onChange={(e) =>
                            setApiKeys(prev => ({ ...prev, [provider.id]: e.target.value }))
                          }
                          disabled={isSaving || isDeleting}
                        />
                      </div>
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={() => toggleShowKey(provider.id)}
                        disabled={isSaving || isDeleting}
                      >
                        {showKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </Button>
                    </div>

                    <div className="flex gap-2">
                      <Button
                        onClick={() => handleSaveApiKey(provider.id)}
                        disabled={isSaving || isDeleting || !apiKeys[provider.id]?.trim()}
                        size="sm"
                      >
                        {isSaving ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Saving...
                          </>
                        ) : (
                          <>
                            <Save className="mr-2 h-4 w-4" />
                            {hasKey ? "Update" : "Save"} Key
                          </>
                        )}
                      </Button>

                      {hasKey && (
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => handleDeleteApiKey(provider.id)}
                          disabled={isSaving || isDeleting}
                        >
                          {isDeleting ? (
                            <>
                              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                              Deleting...
                            </>
                          ) : (
                            <>
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </>
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                )}

                {provider.id !== AI_PROVIDERS[AI_PROVIDERS.length - 1].id && (
                  <Separator />
                )}
              </div>
            );
          })}

          <div className="mt-6 p-4 bg-muted/50 rounded-lg">
            <h5 className="font-medium mb-2">How Organization API Keys Work</h5>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Organization API keys take priority over individual member keys</li>
              <li>• All organization members will use these keys for AI requests</li>
              <li>• Members can still configure individual keys as fallbacks</li>
              <li>• Only organization admins can manage these keys</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}