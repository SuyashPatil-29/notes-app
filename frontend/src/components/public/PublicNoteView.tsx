import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getPublicNote } from '@/utils/publish';
import { MarkdownRenderer } from './MarkdownRenderer';
import { Loader2, ArrowLeft, Calendar, FileText } from 'lucide-react';

export const PublicNoteView: React.FC = () => {
  const { notebookId, chapterId, noteId } = useParams<{
    notebookId: string;
    chapterId: string;
    noteId: string;
  }>();

  const { data: note, isLoading, error } = useQuery({
    queryKey: ['publicNote', notebookId, chapterId, noteId],
    queryFn: () => getPublicNote(notebookId!, chapterId!, noteId!),
    enabled: !!notebookId && !!chapterId && !!noteId,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex items-center gap-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading note...</span>
        </div>
      </div>
    );
  }

  if (error || !note) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Note Not Found</h1>
          <p className="text-muted-foreground mb-6">
            The note you're looking for doesn't exist or is not publicly available.
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              to={`/public/${notebookId}/${chapterId}`}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Chapter
            </Link>
            <Link
              to={`/public/${notebookId}`}
              className="inline-flex items-center gap-2 px-4 py-2 border border-input bg-background hover:bg-accent hover:text-accent-foreground rounded-md transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to Notebook
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center gap-4 mb-4">
            <Link
              to={`/public/${notebookId}/${chapterId}`}
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to {note.chapter.name}
            </Link>
          </div>

          <div className="flex items-start gap-4">
            <div className="p-3 bg-primary/10 rounded-lg">
              <FileText className="h-8 w-8 text-primary" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                <Link
                  to={`/public/${notebookId}`}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                >
                  {note.chapter.notebook.name}
                </Link>
                <span className="text-muted-foreground">/</span>
                <Link
                  to={`/public/${notebookId}/${chapterId}`}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                >
                  {note.chapter.name}
                </Link>
                <span className="text-muted-foreground">/</span>
              </div>
              <h1 className="text-3xl font-bold mb-2">{note.name}</h1>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Created {new Date(note.createdAt).toLocaleDateString()}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto">
          {note.content ? (
            <MarkdownRenderer content={note.content} />
          ) : (
            <div className="text-center py-12">
              <FileText className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h2 className="text-xl font-semibold mb-2">No Content</h2>
              <p className="text-muted-foreground">
                This note doesn't have any content yet.
              </p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};
