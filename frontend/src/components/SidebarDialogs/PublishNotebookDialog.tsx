import React, { useState, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getUserNotebooks } from '@/utils/notebook';
import { publishNotebook, updatePublishedNotes, unpublishNotebook } from '@/utils/publish';
import { getPublishSettings, isNotebookPublished } from '@/utils/publish';
import { toast } from 'sonner';
import { Loader2, Globe, Lock, FileText, ChevronRight, ChevronDown, BookOpen } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Checkbox } from '@/components/ui/checkbox';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import type { Notebook, Chapter } from '@/types/backend';
import { useOrganizationContext } from '@/contexts/OrganizationContext';

interface PublishNotebookDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  notebookId: string | null;
}

export const PublishNotebookDialog: React.FC<PublishNotebookDialogProps> = ({
  open,
  onOpenChange,
  notebookId,
}) => {
  const queryClient = useQueryClient();
  const { activeOrg } = useOrganizationContext();
  const [selectedNoteIds, setSelectedNoteIds] = useState<string[]>([]);
  const [expandedChapters, setExpandedChapters] = useState<Set<string>>(new Set());
  const [isPublishing, setIsPublishing] = useState(false);

  // Get the notebook data
  const { data: notebooks, isLoading } = useQuery({
    queryKey: ['userNotebooks', activeOrg?.id],
    queryFn: () => getUserNotebooks(activeOrg?.id),
    enabled: open,
  });

  const notebook = notebooks?.find((n: Notebook) => n.id === notebookId);

  // Initialize selected notes when dialog opens
  useEffect(() => {
    if (notebook && open) {
      const settings = getPublishSettings(notebook);
      setSelectedNoteIds(settings.selectedNoteIds);

      // Expand chapters that have published notes
      const chaptersToExpand = new Set<string>();
      notebook.chapters?.forEach((chapter: Chapter) => {
        if (chapter.notes?.some(note => note.isPublic)) {
          chaptersToExpand.add(chapter.id);
        }
      });
      setExpandedChapters(chaptersToExpand);
    }
  }, [notebook, open]);

  const handleNoteToggle = (noteId: string) => {
    setSelectedNoteIds(prev =>
      prev.includes(noteId)
        ? prev.filter(id => id !== noteId)
        : [...prev, noteId]
    );
  };

  const handleChapterToggle = (chapterId: string) => {
    const chapter = notebook?.chapters?.find((c: Chapter) => c.id === chapterId);
    if (!chapter) return;

    const chapterNoteIds = chapter.notes?.map(note => note.id) || [];
    const allSelected = chapterNoteIds.every((id: string) => selectedNoteIds.includes(id));

    if (allSelected) {
      // Deselect all notes in chapter
      setSelectedNoteIds(prev => prev.filter(id => !chapterNoteIds.includes(id)));
    } else {
      // Select all notes in chapter
      setSelectedNoteIds(prev => [...new Set([...prev, ...chapterNoteIds])]);
    }
  };

  const handlePublish = async () => {
    if (!notebook) return;

    setIsPublishing(true);
    try {
      if (selectedNoteIds.length === 0) {
        // Unpublish if no notes selected
        await unpublishNotebook(notebook.id);
        toast.success('Notebook unpublished successfully');
      } else if (isNotebookPublished(notebook)) {
        // Update published notes
        await updatePublishedNotes(notebook.id, selectedNoteIds);
        toast.success('Published notes updated successfully');
      } else {
        // Publish notebook
        await publishNotebook(notebook.id, selectedNoteIds);
        toast.success('Notebook published successfully');
      }

      // Refresh notebooks data
      queryClient.invalidateQueries({ queryKey: ['userNotebooks'] });
      onOpenChange(false);
    } catch (error: any) {
      console.error('Publish error:', error);
      toast.error(error.response?.data?.message || 'Failed to update publishing status');
    } finally {
      setIsPublishing(false);
    }
  };

  const toggleChapterExpansion = (chapterId: string) => {
    setExpandedChapters(prev => {
      const newSet = new Set(prev);
      if (newSet.has(chapterId)) {
        newSet.delete(chapterId);
      } else {
        newSet.add(chapterId);
      }
      return newSet;
    });
  };

  const getSelectedCount = () => {
    return selectedNoteIds.length;
  };

  const getTotalCount = () => {
    return notebook?.chapters?.reduce((total: number, chapter: Chapter) => {
      return total + (chapter.notes?.length || 0);
    }, 0) || 0;
  };

  if (isLoading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
            <span className="ml-2">Loading notebook...</span>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (!notebook) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Notebook Not Found</DialogTitle>
            <DialogDescription>
              The notebook you're trying to publish could not be found.
            </DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    );
  }

  const isPublished = isNotebookPublished(notebook);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {isPublished ? (
              <Globe className="h-5 w-5 text-primary" />
            ) : (
              <Lock className="h-5 w-5 text-muted-foreground" />
            )}
            {isPublished ? 'Update Published Content' : 'Publish Notebook'}
          </DialogTitle>
          <DialogDescription>
            {isPublished
              ? 'Select which notes to keep published. Chapters and the notebook will remain public as long as at least one note is published.'
              : 'Select which notes you want to make public. Only chapters containing published notes will be visible.'
            }
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Notebook header */}
          <div className="flex items-center justify-between p-3 bg-muted rounded-lg">
            <div className="flex items-center gap-3">
              <BookOpen className="h-5 w-5" />
              <div>
                <h3 className="font-medium">{notebook.name}</h3>
                <p className="text-sm text-muted-foreground">
                  {notebook.chapters?.length || 0} chapters, {getTotalCount()} notes
                </p>
              </div>
            </div>
            <Badge variant={isPublished ? "default" : "secondary"}>
              {isPublished ? "Published" : "Private"}
            </Badge>
          </div>

          {/* Selection summary */}
          <div className="flex items-center justify-between text-sm">
            <span>
              {getSelectedCount()} of {getTotalCount()} notes selected
            </span>
            {getSelectedCount() > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setSelectedNoteIds([])}
                className="text-xs"
              >
                Clear all
              </Button>
            )}
          </div>

          {/* Notes tree */}
          <ScrollArea className="max-h-96 border rounded-lg p-4">
            <div className="space-y-2">
              {notebook.chapters?.map((chapter: Chapter) => {
                const chapterNotes = chapter.notes || [];
                const selectedNotesInChapter = chapterNotes.filter(note =>
                  selectedNoteIds.includes(note.id)
                );
                const allSelected = chapterNotes.length > 0 &&
                  chapterNotes.every(note => selectedNoteIds.includes(note.id));
                const isExpanded = expandedChapters.has(chapter.id);

                return (
                  <div key={chapter.id} className="space-y-1">
                    {/* Chapter header */}
                    <div
                      className="flex items-center gap-2 p-2 hover:bg-muted rounded cursor-pointer"
                      onClick={() => toggleChapterExpansion(chapter.id)}
                    >
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 w-4 p-0"
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleChapterExpansion(chapter.id);
                        }}
                      >
                        {isExpanded ? (
                          <ChevronDown className="h-3 w-3" />
                        ) : (
                          <ChevronRight className="h-3 w-3" />
                        )}
                      </Button>

                      <Checkbox
                        checked={allSelected}
                        onCheckedChange={() => handleChapterToggle(chapter.id)}
                        onClick={(e: React.MouseEvent) => e.stopPropagation()}
                      />

                      <div className="flex items-center gap-2 flex-1">
                        <BookOpen className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{chapter.name}</span>
                        <span className="text-xs text-muted-foreground">
                          ({selectedNotesInChapter.length}/{chapterNotes.length})
                        </span>
                      </div>
                    </div>

                    {/* Chapter notes */}
                    {isExpanded && chapterNotes.map((note: any) => (
                      <div
                        key={note.id}
                        className="flex items-center gap-2 pl-8 py-1"
                      >
                        <Checkbox
                          checked={selectedNoteIds.includes(note.id)}
                          onCheckedChange={() => handleNoteToggle(note.id)}
                        />
                        <FileText className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">{note.name}</span>
                      </div>
                    ))}
                  </div>
                );
              })}
            </div>
          </ScrollArea>

          {/* Actions */}
          <div className="flex justify-between items-center pt-4 border-t">
            <div className="text-sm text-muted-foreground">
              {selectedNoteIds.length === 0 ? (
                'No notes selected - notebook will be unpublished'
              ) : (
                `${selectedNoteIds.length} note${selectedNoteIds.length !== 1 ? 's' : ''} will be published`
              )}
            </div>

            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isPublishing}
              >
                Cancel
              </Button>
              <Button
                onClick={handlePublish}
                disabled={isPublishing}
              >
                {isPublishing && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
                {selectedNoteIds.length === 0 ? 'Unpublish' : (isPublished ? 'Update' : 'Publish')}
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};
