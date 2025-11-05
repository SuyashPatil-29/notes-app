import { useEffect, useState } from 'react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import GraphVisualization from '@/components/GraphVisualization';
import { useNavigate } from 'react-router-dom';

interface GraphModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function GraphModal({ open, onOpenChange }: GraphModalProps) {
  const navigate = useNavigate();

  const handleNodeClick = (nodeId: string, metadata?: Record<string, string>) => {
    // Construct proper path: /notebookId/chapterId/noteId
    if (metadata?.notebookId && metadata?.chapterId) {
      navigate(`/${metadata.notebookId}/${metadata.chapterId}/${nodeId}`);
    } else {
      // Fallback to old path if metadata is missing
      navigate(`/note/${nodeId}`);
    }
    onOpenChange(false);
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="md:min-w-[50vw] max-w-none p-0 flex flex-col">
        <SheetHeader className="px-6 pt-6 pb-2 shrink-0">
          <SheetTitle>Note Graph</SheetTitle>
          <SheetDescription>
            Explore the connections between your notes. Click a node to navigate to that note.
          </SheetDescription>
        </SheetHeader>
        <div className="flex-1 w-full overflow-hidden">
          <GraphVisualization
            onNodeClick={handleNodeClick}
          />
        </div>
      </SheetContent>
    </Sheet>
  );
}

// Hook to manage global graph modal state with keyboard shortcut
export function useGraphModal() {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Check for Cmd/Ctrl + G
      if ((event.metaKey || event.ctrlKey) && event.key === 'g') {
        event.preventDefault();
        setIsOpen((prev) => !prev);
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  return {
    isOpen,
    openGraph: () => setIsOpen(true),
    closeGraph: () => setIsOpen(false),
    toggleGraph: () => setIsOpen((prev) => !prev),
  };
}

