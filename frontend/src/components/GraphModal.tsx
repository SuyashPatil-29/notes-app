import { useEffect, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import GraphVisualization from '@/components/GraphVisualization';
import { useNavigate } from 'react-router-dom';

interface GraphModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  centerNodeId?: string;
}

export function GraphModal({ open, onOpenChange, centerNodeId }: GraphModalProps) {
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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[95vw] w-[95vw] h-[90vh] p-0 flex flex-col">
        <DialogHeader className="px-6 pt-6 pb-2 shrink-0">
          <DialogTitle>Note Graph</DialogTitle>
          <DialogDescription>
            Explore the connections between your notes. Click a node to navigate to that note.
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 w-full overflow-hidden">
          <GraphVisualization
            onNodeClick={handleNodeClick}
            centerNodeId={centerNodeId}
          />
        </div>
      </DialogContent>
    </Dialog>
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

