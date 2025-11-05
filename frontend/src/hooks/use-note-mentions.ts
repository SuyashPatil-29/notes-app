import { useCallback, useRef } from 'react';
import { Editor } from '@tiptap/react';
import { createNoteLink, deleteNoteLink, getAllLinks } from '@/utils/graphApi';
import { useOrganizationContext } from '@/contexts/OrganizationContext';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

/**
 * Hook to automatically create/delete note links when mentions are added/removed in the editor
 * Returns a function to be called from the editor's onUpdate callback
 */
export function useNoteMentions(currentNoteId: string | undefined) {
  const processedMentions = useRef<Set<string>>(new Set());
  const linkIdMap = useRef<Map<string, string>>(new Map()); // Maps linkKey to linkId for deletion
  const previousMentions = useRef<Set<string>>(new Set());
  const { activeOrg } = useOrganizationContext();
  const queryClient = useQueryClient();
  const isProcessing = useRef(false);
  const lastProcessedContent = useRef<string>('');
  const isInitialized = useRef(false);

  const processMentions = useCallback(async (editor: Editor) => {
    if (!editor || !currentNoteId) {
      return;
    }

    // Prevent concurrent processing
    if (isProcessing.current) {
      console.log('[useNoteMentions] Already processing, skipping...');
      return;
    }

    // Get current content hash to avoid reprocessing same content
    const currentContent = JSON.stringify(editor.getJSON());
    if (currentContent === lastProcessedContent.current) {
      return;
    }
    lastProcessedContent.current = currentContent;

    isProcessing.current = true;

    try {
      // Get all mention nodes from the current document
      const { doc } = editor.state;
      const currentMentionIds: string[] = [];

      doc.descendants((node) => {
        if (node.type.name === 'mention' && node.attrs.id) {
          currentMentionIds.push(node.attrs.id);
        }
      });

      const currentMentionsSet = new Set(currentMentionIds);
      console.log('[useNoteMentions] Current mentions in document:', Array.from(currentMentionsSet));

      // Initialize on first run - fetch existing links to populate linkIdMap
      if (!isInitialized.current) {
        try {
          const allLinks = await getAllLinks();
          for (const link of allLinks) {
            if (link.sourceNoteId === currentNoteId) {
              const linkKey = `${currentNoteId}-${link.targetNoteId}`;
              linkIdMap.current.set(linkKey, link.id);
              processedMentions.current.add(linkKey);
            }
          }
          console.log('[useNoteMentions] Initialized with existing links:', Array.from(linkIdMap.current.keys()));
        } catch (error) {
          console.error('[useNoteMentions] Failed to fetch existing links:', error);
        }
        isInitialized.current = true;
        previousMentions.current = currentMentionsSet;
        isProcessing.current = false;
        return;
      }

      // Find removed mentions (in previous but not in current)
      const removedMentions = Array.from(previousMentions.current).filter(
        (mentionId) => !currentMentionsSet.has(mentionId)
      );

      // Find added mentions (in current but not in previous)
      const addedMentions = Array.from(currentMentionsSet).filter(
        (mentionId) => !previousMentions.current.has(mentionId) && mentionId !== currentNoteId
      );

      console.log('[useNoteMentions] Added mentions:', addedMentions);
      console.log('[useNoteMentions] Removed mentions:', removedMentions);

      const promises: Promise<void>[] = [];
      let linksCreated = 0;
      let linksDeleted = 0;

      // Create links for added mentions
      for (const mentionId of addedMentions) {
        const linkKey = `${currentNoteId}-${mentionId}`;
        
        if (!processedMentions.current.has(linkKey)) {
          processedMentions.current.add(linkKey);
          linksCreated++;
          
          console.log('[useNoteMentions] Creating link:', {
            from: currentNoteId,
            to: mentionId,
            org: activeOrg?.id,
          });

          const linkPromise = createNoteLink({
            sourceNoteId: currentNoteId,
            targetNoteId: mentionId,
            linkType: 'references',
          }, activeOrg?.id)
            .then((link) => {
              console.log('[useNoteMentions] ✅ Link created successfully');
              // Store the link ID for future deletion
              linkIdMap.current.set(linkKey, link.id);
            })
            .catch((error) => {
              console.error('[useNoteMentions] ❌ Failed to create link:', error);
              // Remove from processed set if creation failed
              processedMentions.current.delete(linkKey);
            });
          
          promises.push(linkPromise);
        }
      }

      // Delete links for removed mentions
      for (const mentionId of removedMentions) {
        const linkKey = `${currentNoteId}-${mentionId}`;
        const linkId = linkIdMap.current.get(linkKey);
        
        if (linkId && processedMentions.current.has(linkKey)) {
          linksDeleted++;
          
          console.log('[useNoteMentions] Deleting link:', {
            from: currentNoteId,
            to: mentionId,
            linkId,
          });

          const deletePromise = deleteNoteLink(linkId)
            .then(() => {
              console.log('[useNoteMentions] ✅ Link deleted successfully');
              // Remove from tracking
              processedMentions.current.delete(linkKey);
              linkIdMap.current.delete(linkKey);
            })
            .catch((error) => {
              console.error('[useNoteMentions] ❌ Failed to delete link:', error);
            });
          
          promises.push(deletePromise);
        }
      }

      // Update previous mentions set
      previousMentions.current = currentMentionsSet;

      // Wait for all operations to complete
      if (promises.length > 0) {
        await Promise.all(promises);
        
        console.log('[useNoteMentions] Invalidating graph data...');
        
        // Invalidate graph query immediately to update the visualization
        queryClient.invalidateQueries({ 
          queryKey: ['graphData', activeOrg?.id]
        });
        
        console.log('[useNoteMentions] Graph data invalidated');
        
        // Show appropriate toast messages
        if (linksCreated > 0 && linksDeleted > 0) {
          toast.success(`${linksCreated} link${linksCreated > 1 ? 's' : ''} created, ${linksDeleted} deleted`);
        } else if (linksCreated > 0) {
          toast.success(`${linksCreated} note link${linksCreated > 1 ? 's' : ''} created`);
        } else if (linksDeleted > 0) {
          toast.success(`${linksDeleted} note link${linksDeleted > 1 ? 's' : ''} deleted`);
        }
      }
    } finally {
      isProcessing.current = false;
    }
  }, [currentNoteId, activeOrg?.id, queryClient]);

  return { processMentions };
}

