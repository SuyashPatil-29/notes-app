import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getPublicUserProfile } from '@/utils/publish';
import { Loader2, BookOpen, ArrowLeft, Calendar, User as UserIcon } from 'lucide-react';
import type { Notebook } from '@/types/backend';

export const PublicUserProfile: React.FC = () => {
  const { email } = useParams<{ email: string }>();

  const { data: user, isLoading, error } = useQuery({
    queryKey: ['publicUserProfile', email],
    queryFn: () => getPublicUserProfile(email!),
    enabled: !!email,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading profile...</span>
        </div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">User Not Found</h1>
          <p className="text-muted-foreground mb-6">
            The user profile you're looking for doesn't exist or has no public content.
          </p>
          <Link
            to="/"
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
            Go Home
          </Link>
        </div>
      </div>
    );
  }

  const publicNotebooks = user.notebooks?.filter((notebook: Notebook) => notebook.isPublic) || [];
  const totalPublicNotes = publicNotebooks.reduce((total: number, notebook: Notebook) => {
    return total + (notebook.chapters?.reduce((chapterTotal: number, chapter) => {
      return chapterTotal + (chapter.notes?.filter(note => note.isPublic).length || 0);
    }, 0) || 0);
  }, 0);

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center gap-4 mb-6">
            <Link
              to="/"
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Home
            </Link>
          </div>

          <div className="flex items-start gap-4">
            <div className="p-3 bg-primary/10 rounded-lg">
              <UserIcon className="h-8 w-8 text-primary" />
            </div>
            <div className="flex-1">
              <h1 className="text-3xl font-bold mb-2">{user.name}</h1>
              <p className="text-muted-foreground mb-4">{user.email}</p>
              <div className="flex items-center gap-6 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <BookOpen className="h-4 w-4" />
                  {publicNotebooks.length} notebook{publicNotebooks.length !== 1 ? 's' : ''}
                </span>
                <span>{totalPublicNotes} note{totalPublicNotes !== 1 ? 's' : ''}</span>
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Joined {new Date(user.createdAt).toLocaleDateString()}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="container mx-auto px-4 py-8">
        {publicNotebooks.length === 0 ? (
          <div className="text-center py-12">
            <BookOpen className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">No Public Notebooks</h2>
            <p className="text-muted-foreground">
              This user hasn't published any notebooks yet.
            </p>
          </div>
        ) : (
          <div>
            <h2 className="text-xl font-semibold mb-6">Public Notebooks</h2>
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {publicNotebooks.map((notebook: Notebook) => {
                const publicChapters = notebook.chapters?.filter(chapter => chapter.isPublic) || [];
                const publicNotes = publicChapters.reduce((total: number, chapter) => {
                  return total + (chapter.notes?.filter(note => note.isPublic).length || 0);
                }, 0);

                return (
                  <Link
                    key={notebook.id}
                    to={`/public/${notebook.id}`}
                    className="block p-6 border rounded-lg hover:shadow-md transition-shadow bg-card"
                  >
                    <div className="flex items-start justify-between mb-3">
                      <h3 className="font-semibold text-lg leading-tight">{notebook.name}</h3>
                      <BookOpen className="h-5 w-5 text-muted-foreground shrink-0 ml-2" />
                    </div>
                    <div className="space-y-1 text-sm text-muted-foreground mb-3">
                      <div>{publicChapters.length} chapter{publicChapters.length !== 1 ? 's' : ''}</div>
                      <div>{publicNotes} note{publicNotes !== 1 ? 's' : ''}</div>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      Created {new Date(notebook.createdAt).toLocaleDateString()}
                    </div>
                  </Link>
                );
              })}
            </div>
          </div>
        )}
      </main>
    </div>
  );
};
