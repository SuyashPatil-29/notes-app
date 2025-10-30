import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getPublicChapter } from '@/utils/publish';
import { Loader2, FileText, ArrowLeft, Calendar, BookOpen } from 'lucide-react';
import { getPreviewText } from '@/utils/markdown';

export const PublicChapterView: React.FC = () => {
  const { notebookId, chapterId } = useParams<{ notebookId: string; chapterId: string }>();

  const { data: chapter, isLoading, error } = useQuery({
    queryKey: ['publicChapter', notebookId, chapterId],
    queryFn: () => getPublicChapter(notebookId!, chapterId!),
    enabled: !!notebookId && !!chapterId,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading chapter...</span>
        </div>
      </div>
    );
  }

  if (error || !chapter) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Chapter Not Found</h1>
          <p className="text-muted-foreground mb-6">
            The chapter you're looking for doesn't exist or is not publicly available.
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              to={`/public/${notebookId}`}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Notebook
            </Link>
            <Link
              to="/"
              className="inline-flex items-center gap-2 px-4 py-2 border border-input bg-background hover:bg-accent hover:text-accent-foreground rounded-md transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Home
            </Link>
          </div>
        </div>
      </div>
    );
  }

  const publicNotes = chapter.notes?.filter(note => note.isPublic) || [];

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center gap-4 mb-4">
            <Link
              to={`/public/${notebookId}`}
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to {chapter.notebook.name}
            </Link>
          </div>

          <div className="flex items-start gap-4">
            <div className="p-3 bg-primary/10 rounded-lg">
              <BookOpen className="h-8 w-8 text-primary" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                <h2 className="text-lg text-muted-foreground">{chapter.notebook.name}</h2>
                <span className="text-muted-foreground">/</span>
              </div>
              <h1 className="text-3xl font-bold mb-2">{chapter.name}</h1>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Created {new Date(chapter.createdAt).toLocaleDateString()}
                </span>
                <span>{publicNotes.length} note{publicNotes.length !== 1 ? 's' : ''}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="container mx-auto px-4 py-8">
        {publicNotes.length === 0 ? (
          <div className="text-center py-12">
            <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">No Public Notes</h2>
            <p className="text-muted-foreground">
              This chapter doesn't have any publicly available notes yet.
            </p>
          </div>
        ) : (
          <div>
            <h2 className="text-xl font-semibold mb-6">Notes</h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {publicNotes.map((note) => (
                <Link
                  key={note.id}
                  to={`/public/${notebookId}/${chapterId}/${note.id}`}
                  className="block p-6 border rounded-lg hover:shadow-md transition-shadow bg-card"
                >
                  <div className="flex items-start justify-between mb-2">
                    <h3 className="font-semibold text-lg leading-tight">{note.name}</h3>
                    <FileText className="h-5 w-5 text-muted-foreground shrink-0 ml-2" />
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span>{new Date(note.createdAt).toLocaleDateString()}</span>
                  </div>
                  <div className="mt-3">
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {getPreviewText(note.content, 100)}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}
      </main>
    </div>
  );
};
