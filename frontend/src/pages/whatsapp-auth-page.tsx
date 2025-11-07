import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '@clerk/clerk-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle2, XCircle, Loader2, MessageSquare } from 'lucide-react';
import api from '@/utils/api';

export function WhatsAppAuthPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isSignedIn, isLoaded } = useAuth();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [errorMessage, setErrorMessage] = useState('');
  const token = searchParams.get('token');

  useEffect(() => {
    // Wait for Clerk to be loaded
    if (!isLoaded) {
      return;
    }

    // If not signed in, redirect to sign-in page with return URL
    if (!isSignedIn) {
      const returnUrl = `/whatsapp-auth?token=${token}`;
      navigate(`/sign-in?redirect_url=${encodeURIComponent(returnUrl)}`);
      return;
    }

    // If no token, show error
    if (!token) {
      setStatus('error');
      setErrorMessage('Invalid authentication link. Please request a new one from WhatsApp.');
      return;
    }

    // Link WhatsApp account
    linkWhatsAppAccount();
  }, [isLoaded, isSignedIn, token, navigate]);

  const linkWhatsAppAccount = async () => {
    try {
      setStatus('loading');
      
      const response = await api.post(`/api/whatsapp/link?token=${token}`);
      
      if (response.status === 200) {
        setStatus('success');
      } else {
        setStatus('error');
        setErrorMessage('Failed to link WhatsApp account. Please try again.');
      }
    } catch (error: any) {
      console.error('Error linking WhatsApp account:', error);
      setStatus('error');
      
      if (error.response?.status === 400) {
        setErrorMessage('Invalid or expired authentication link. Please request a new one from WhatsApp.');
      } else if (error.response?.status === 401) {
        setErrorMessage('Please sign in to link your WhatsApp account.');
      } else if (error.response?.data?.error) {
        setErrorMessage(error.response.data.error);
      } else {
        setErrorMessage('An unexpected error occurred. Please try again.');
      }
    }
  };

  const handleTryAgain = () => {
    navigate('/dashboard');
  };

  if (!isLoaded || (isSignedIn && status === 'loading')) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
            <CardTitle>Linking WhatsApp Account</CardTitle>
            <CardDescription>Please wait while we connect your WhatsApp account...</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 flex items-center justify-center">
              <CheckCircle2 className="h-12 w-12 text-green-500" />
            </div>
            <CardTitle>WhatsApp Account Linked!</CardTitle>
            <CardDescription>Your WhatsApp account has been successfully connected.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert>
              <MessageSquare className="h-4 w-4" />
              <AlertDescription>
                You can now use WhatsApp commands to:
                <ul className="mt-2 ml-4 list-disc space-y-1">
                  <li>Create and manage notes</li>
                  <li>Search and retrieve notes</li>
                  <li>Organize notebooks and chapters</li>
                  <li>And much more!</li>
                </ul>
              </AlertDescription>
            </Alert>

            <div className="bg-muted p-4 rounded-md">
              <p className="text-sm font-medium mb-2">Try these commands in WhatsApp:</p>
              <ul className="text-sm space-y-1 text-muted-foreground">
                <li><code className="bg-background px-1.5 py-0.5 rounded">/help</code> - View all available commands</li>
                <li><code className="bg-background px-1.5 py-0.5 rounded">/add [title] [content]</code> - Create a note</li>
                <li><code className="bg-background px-1.5 py-0.5 rounded">/list</code> - List your notebooks</li>
              </ul>
            </div>

            <Button
              onClick={handleTryAgain}
              className="w-full"
            >
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 flex items-center justify-center">
              <XCircle className="h-12 w-12 text-destructive" />
            </div>
            <CardTitle>Link Failed</CardTitle>
            <CardDescription>We couldn't link your WhatsApp account</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>

            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                To get a new authentication link, send any message to the WhatsApp bot.
              </p>
            </div>

            <Button
              onClick={handleTryAgain}
              className="w-full"
              variant="outline"
            >
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return null;
}

