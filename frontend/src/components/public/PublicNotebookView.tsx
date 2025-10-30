import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getPublicNotebook } from '@/utils/publish';
import { Loader2, BookOpen, ArrowLeft, Calendar } from 'lucide-react';

export const PublicNotebookView: React.FC = () => {
  const { notebookId } = useParams<{ notebookId: string }>();

  const { data: notebook, isLoading, error } = useQuery({
    queryKey: ['publicNotebook', notebookId],
    queryFn: () => getPublicNotebook(notebookId!),
    enabled: !!notebookId,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading notebook...</span>
        </div>
      </div>
    );
  }

  if (error || !notebook) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Notebook Not Found</h1>
          <p className="text-muted-foreground mb-6">
            The notebook you're looking for doesn't exist or is not publicly available.
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

  const publicChapters = notebook.chapters?.filter(chapter => chapter.isPublic) || [];

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center gap-4 mb-4">
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
              <BookOpen className="h-8 w-8 text-primary" />
            </div>
            <div className="flex-1">
              <h1 className="text-3xl font-bold mb-2">{notebook.name}</h1>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Created {new Date(notebook.createdAt).toLocaleDateString()}
                </span>
                <span>{publicChapters.length} chapter{publicChapters.length !== 1 ? 's' : ''}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="container mx-auto px-4 py-8">
        {publicChapters.length === 0 ? (
          <div className="text-center py-12">
            <BookOpen className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">No Public Chapters</h2>
            <p className="text-muted-foreground">
              This notebook doesn't have any publicly available chapters yet.
            </p>
          </div>
        ) : (
          <div>
            <h2 className="text-xl font-semibold mb-6">Chapters</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {publicChapters.map((chapter) => {
                const publicNotes = chapter.notes?.filter(note => note.isPublic) || [];

                return (
                  <Link
                    key={chapter.id}
                    to={`/public/${notebookId}/${chapter.id}`}
                    className="block p-6 border rounded-lg hover:shadow-md transition-shadow bg-card"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h3 className="font-semibold text-lg leading-tight">{chapter.name}</h3>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>{publicNotes.length} note{publicNotes.length !== 1 ? 's' : ''}</span>
                      <span>{new Date(chapter.createdAt).toLocaleDateString()}</span>
                    </div>
                    {publicNotes.length > 0 && (
                      <div className="mt-3">
                        <p className="text-sm text-muted-foreground">
                          Latest: {publicNotes[0].name}
                        </p>
                      </div>
                    )}
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
