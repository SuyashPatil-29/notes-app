import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import api from "@/utils/api";
import { useUser as useClerkUser } from '@clerk/clerk-react';

interface OnboardingWizardProps {
  onComplete: () => void;
}

const onboardingOptions = [
  { value: "personal", label: "Personal", description: "For personal notes, journaling, and ideas" },
  { value: "work", label: "Work", description: "For professional projects and collaboration" },
  { value: "study", label: "Study", description: "For academic notes and research" },
  { value: "creative", label: "Creative", description: "For writing, design, and artistic projects" },
];

export function OnboardingWizard({ onComplete }: OnboardingWizardProps) {
  const [step, setStep] = useState<1 | 2>(1);
  const [selectedType, setSelectedType] = useState<string>("");
  const [openAIKey, setOpenAIKey] = useState<string>("");
  const [anthropicKey, setAnthropicKey] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { user: clerkUser } = useClerkUser();

  const completeOnboarding = async (skipApiKeys: boolean = false) => {
    if (!selectedType) {
      toast.error("Please select an onboarding type");
      return;
    }

    setIsSubmitting(true);
    try {
      // Complete onboarding - this updates Clerk's publicMetadata
      await api.post("/onboarding", { type: selectedType });
      
      // Save API keys if provided and not skipping
      if (!skipApiKeys) {
        const apiKeyErrors: string[] = [];
        
        if (openAIKey.trim()) {
          try {
            await api.post("/settings/ai-credentials", {
              provider: "openai",
              apiKey: openAIKey.trim(),
            });
          } catch (error: any) {
            console.error("Failed to save OpenAI key:", error);
            apiKeyErrors.push("OpenAI");
          }
        }
        
        if (anthropicKey.trim()) {
          try {
            await api.post("/settings/ai-credentials", {
              provider: "anthropic",
              apiKey: anthropicKey.trim(),
            });
          } catch (error: any) {
            console.error("Failed to save Anthropic key:", error);
            apiKeyErrors.push("Anthropic");
          }
        }

        // Show success or warning
        if (apiKeyErrors.length > 0) {
          toast.warning(`Onboarding complete! However, failed to save ${apiKeyErrors.join(" and ")} API key(s). You can add them later in settings.`);
        } else {
          toast.success("Welcome to Notes App!");
        }
      } else {
        toast.success("Welcome to Notes App! You can set up API keys later in settings.");
      }
      
      // Reload Clerk user data to get updated publicMetadata
      await clerkUser?.reload();
      
      setIsSubmitting(false);
      // Call onComplete after a brief delay to ensure state updates
      setTimeout(() => {
        onComplete();
      }, 100);
    } catch (error: any) {
      console.error("Onboarding error:", error);
      const errorMessage = error.response?.data?.error || error.message || "Failed to complete onboarding. Please try again.";
      toast.error(errorMessage);
      setIsSubmitting(false);
    }
  };

  const handleSubmit = () => {
    completeOnboarding(false);
  };

  const handleContinue = () => {
    if (!selectedType) {
      toast.error("Please select an onboarding type");
      return;
    }
    if (step === 1) {
      setStep(2);
    }
  };

  const handleBack = () => {
    if (step === 2) {
      setStep(1);
    }
  };

  const handleSkipApiKeys = () => {
    completeOnboarding(true);
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-md max-h-[90vh] overflow-y-auto bg-card rounded-lg border p-6 shadow-lg">
        {/* Step Indicator */}
        <div className="flex items-center justify-center mb-6">
          <div className="flex items-center gap-2">
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
              step >= 1 ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
            }`}>
              1
            </div>
            <div className={`w-12 h-0.5 ${
              step >= 2 ? "bg-primary" : "bg-muted"
            }`} />
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
              step >= 2 ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
            }`}>
              2
            </div>
          </div>
        </div>

        {/* Step 1: Select Type */}
        {step === 1 && (
          <>
            <div className="text-center mb-6">
              <h1 className="text-2xl font-semibold mb-2">Welcome to Notes App!</h1>
              <p className="text-muted-foreground">
                Let's get you set up. What will you primarily use Notes App for?
              </p>
            </div>
            <div className="space-y-6">
              <div className="space-y-3">
                {onboardingOptions.map((option) => {
                  const isSelected = selectedType === option.value;
                  return (
                    <button
                      key={option.value}
                      type="button"
                      onClick={() => setSelectedType(option.value)}
                      className={`
                        w-full text-left p-4 rounded-md border-2 transition-all duration-200
                        ${isSelected
                          ? "border-primary bg-primary text-primary-foreground shadow-md"
                          : "border-border bg-card text-foreground hover:border-primary/50 hover:bg-accent"
                        }
                      `}
                    >
                      <div className="font-medium text-base">{option.label}</div>
                      <div className={`text-sm mt-1 ${
                        isSelected ? "text-primary-foreground/90" : "text-muted-foreground"
                      }`}>
                        {option.description}
                      </div>
                    </button>
                  );
                })}
              </div>

              <Button
                onClick={handleContinue}
                disabled={!selectedType}
                className="w-full"
              >
                Continue
              </Button>
            </div>
          </>
        )}

        {/* Step 2: API Keys */}
        {step === 2 && (
          <>
            <div className="text-center mb-6">
              <h1 className="text-2xl font-semibold mb-2">Set Up AI API Keys</h1>
              <p className="text-muted-foreground">
                Configure your API keys to start using the AI chat feature. You can skip this and set them up later.
              </p>
            </div>
            <div className="space-y-6">
              <div className="space-y-4">
                <div>
                  <Label htmlFor="openai-key" className="text-sm font-medium mb-2 block text-foreground">
                    OpenAI API Key <span className="text-muted-foreground font-normal">(optional)</span>
                  </Label>
                  <Input
                    id="openai-key"
                    type="password"
                    placeholder="sk-..."
                    value={openAIKey}
                    onChange={(e) => setOpenAIKey(e.target.value)}
                    className="w-full"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <Label htmlFor="anthropic-key" className="text-sm font-medium mb-2 block text-foreground">
                    Anthropic API Key <span className="text-muted-foreground font-normal">(optional)</span>
                  </Label>
                  <Input
                    id="anthropic-key"
                    type="password"
                    placeholder="sk-ant-..."
                    value={anthropicKey}
                    onChange={(e) => setAnthropicKey(e.target.value)}
                    className="w-full"
                    disabled={isSubmitting}
                  />
                </div>
                
                <div className="p-3 rounded-md bg-muted/50 border border-border">
                  <p className="text-xs text-muted-foreground mb-2">
                    <strong className="text-foreground">🔒 Secure Storage:</strong> Your API keys are encrypted using AES-256 encryption and stored securely on our servers. They are never exposed in plain text and only decrypted when needed for API calls.
                  </p>
                  <p className="text-xs text-muted-foreground">
                    You can configure API keys in the AI Settings later to start using the chat feature.
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                {(openAIKey.trim() || anthropicKey.trim()) ? (
                  <Button
                    onClick={handleSubmit}
                    disabled={isSubmitting}
                    className="w-full"
                  >
                    {isSubmitting ? "Setting up..." : "Complete Setup"}
                  </Button>
                ) : (
                  <Button
                    onClick={handleSkipApiKeys}
                    disabled={isSubmitting}
                    className="w-full"
                  >
                    {isSubmitting ? "Setting up..." : "Skip & Complete"}
                  </Button>
                )}
                <Button
                  onClick={handleBack}
                  disabled={isSubmitting}
                  variant="outline"
                  className="w-full"
                >
                  Back
                </Button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
