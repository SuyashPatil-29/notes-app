import { useEffect, useState } from "react";
import { useSignIn, useSignUp, useOrganization } from "@clerk/clerk-react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

export function AcceptInvitationPage() {
  const { isLoaded: signUpLoaded, signUp, setActive: setActiveSignUp } = useSignUp();
  const { isLoaded: signInLoaded, signIn, setActive: setActiveSignIn } = useSignIn();
  const { organization } = useOrganization();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [fullName, setFullName] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [isLoadingEmail, setIsLoadingEmail] = useState(true);

  // Get the token and account status from the query params
  const token = searchParams.get("__clerk_ticket");
  const accountStatus = searchParams.get("__clerk_status");

  // Extract email from the invitation ticket
  useEffect(() => {
    if (!token || !signUp || accountStatus !== "sign_up") {
      setIsLoadingEmail(false);
      return;
    }

    const fetchInvitationEmail = async () => {
      try {
        // Create a temporary sign-up to get the email from the ticket
        // Clerk automatically extracts the email from the ticket
        const tempSignUp = await signUp.create({
          strategy: "ticket",
          ticket: token,
        });

        // Extract email from the sign-up attempt
        if (tempSignUp.emailAddress) {
          setEmail(tempSignUp.emailAddress);
        }
      } catch (err: any) {
        console.error("Error fetching invitation email:", err);
        // If we can't get the email, we'll let the user continue
      } finally {
        setIsLoadingEmail(false);
      }
    };

    fetchInvitationEmail();
  }, [token, signUp, accountStatus]);

  // If there is no invitation token, redirect or show error
  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Invalid Invitation</CardTitle>
            <CardDescription>
              No invitation token found. Please check your invitation link.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate("/sign-in")} className="w-full">
              Go to Sign In
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Handle sign-in for existing users
  useEffect(() => {
    if (!signInLoaded || !setActiveSignIn || !token || organization || accountStatus !== "sign_in") {
      return;
    }

    const createSignIn = async () => {
      setIsProcessing(true);
      try {
        // Create a new SignIn with the supplied invitation token
        const signInAttempt = await signIn.create({
          strategy: "ticket",
          ticket: token,
        });

        // If the sign-in was successful, set the session to active
        if (signInAttempt.status === "complete") {
          await setActiveSignIn({
            session: signInAttempt.createdSessionId,
          });
          
          toast.success("Successfully joined the organization!");
          
          // Navigate to dashboard after a brief delay
          setTimeout(() => {
            navigate("/");
          }, 500);
        } else {
          // If the sign-in attempt is not complete, check why
          console.error("Sign-in not complete:", JSON.stringify(signInAttempt, null, 2));
          toast.error("Could not complete sign-in. Please try again.");
        }
      } catch (err: any) {
        console.error("Sign-in error:", JSON.stringify(err, null, 2));
        toast.error(err.errors?.[0]?.message || "Failed to sign in. Please try again.");
      } finally {
        setIsProcessing(false);
      }
    };

    createSignIn();
  }, [signIn, signInLoaded, setActiveSignIn, token, organization, accountStatus, navigate]);

  // Handle submission of the sign-up form for new users
  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!signUpLoaded || !signUp) {
      toast.error("Please wait, loading...");
      return;
    }

    if (!fullName.trim() || !password.trim()) {
      toast.error("Please fill in all fields");
      return;
    }

    // Split full name into first and last name
    const nameParts = fullName.trim().split(/\s+/);
    const firstName = nameParts[0] || "";
    const lastName = nameParts.slice(1).join(" ") || nameParts[0]; // Use first name as last if only one word

    if (!firstName) {
      toast.error("Please enter your full name");
      return;
    }

    setIsProcessing(true);
    try {
      // Create a new sign-up with the supplied invitation token
      const signUpAttempt = await signUp.create({
        strategy: "ticket",
        ticket: token!,
        firstName: firstName,
        lastName: lastName,
        password: password.trim(),
      });

      // If the sign-up was successful, set the session to active
      if (signUpAttempt.status === "complete") {
        await setActiveSignUp!({ session: signUpAttempt.createdSessionId });
        
        toast.success("Account created! Welcome to the organization!");
        
        // Navigate to dashboard after a brief delay
        setTimeout(() => {
          navigate("/");
        }, 500);
      } else {
        // If the sign-up attempt is not complete, check why
        console.error("Sign-up not complete:", JSON.stringify(signUpAttempt, null, 2));
        toast.error("Could not complete sign-up. Please try again.");
      }
    } catch (err: any) {
      console.error("Sign-up error:", JSON.stringify(err, null, 2));
      toast.error(err.errors?.[0]?.message || "Failed to create account. Please try again.");
    } finally {
      setIsProcessing(false);
    }
  };

  // Show loading state while signing in existing user
  if (accountStatus === "sign_in" && !organization) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-background">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="flex flex-col items-center justify-center space-y-4 py-8">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <div className="text-center space-y-2">
                <h2 className="text-xl font-semibold">Signing you in...</h2>
                <p className="text-sm text-muted-foreground">
                  Please wait while we process your invitation
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Show sign-up form for new users
  if (accountStatus === "sign_up" && !organization) {
    // Show loading spinner while fetching email - centered without card
    if (isLoadingEmail || !email) {
      return (
        <div className="min-h-screen flex items-center justify-center p-4 bg-background">
          <div className="flex flex-col items-center justify-center space-y-4">
            <Loader2 className="h-16 w-16 animate-spin text-primary" />
            <p className="text-lg font-medium">Loading invitation...</p>
            <p className="text-sm text-muted-foreground">Please wait while we prepare your account</p>
          </div>
        </div>
      );
    }

    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Complete Your Profile</CardTitle>
            <CardDescription>
              Create your account to accept the organization invitation
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSignUp} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email Address</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  disabled
                  className="bg-muted cursor-not-allowed"
                />
                <p className="text-xs text-muted-foreground">
                  This email was invited to join the organization
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="fullName">Full Name</Label>
                <Input
                  id="fullName"
                  type="text"
                  placeholder="John Doe"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  disabled={isProcessing}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  disabled={isProcessing}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Must be at least 8 characters long
                </p>
              </div>

              <Button type="submit" disabled={isProcessing} className="w-full">
                {isProcessing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating Account...
                  </>
                ) : (
                  "Create Account & Join"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Show success message if already in organization
  if (accountStatus === "complete" || organization) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4 bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>✓ Invitation Accepted!</CardTitle>
            <CardDescription>
              You have successfully joined the organization.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate("/")} className="w-full">
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Fallback loading state
  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          <div className="flex flex-col items-center justify-center space-y-4 py-8">
            <Loader2 className="h-12 w-12 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Processing invitation...</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

