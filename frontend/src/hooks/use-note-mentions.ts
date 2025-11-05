import { useEffect, useRef } from 'react';
import { Editor } from '@tiptap/react';
import { createNoteLink } from '@/utils/graphApi';
import { useOrganizationContext } from '@/contexts/OrganizationContext';

/**
 * Hook to automatically create note links when mentions are added in the editor
 */
export function useNoteMentions(editor: Editor | null, currentNoteId: string | undefined) {
  const processedMentions = useRef<Set<string>>(new Set());
  const { activeOrg } = useOrganizationContext();

  useEffect(() => {
    if (!editor || !currentNoteId) {
      console.log('[useNoteMentions] Hook not active:', { editor: !!editor, currentNoteId });
      return;
    }

    console.log('[useNoteMentions] Hook initialized for note:', currentNoteId, 'org:', activeOrg?.id);

    const handleUpdate = () => {
      // Get all mention nodes from the current document
      const { doc } = editor.state;
      const mentions: string[] = [];

      doc.descendants((node) => {
        if (node.type.name === 'mention' && node.attrs.id) {
          mentions.push(node.attrs.id);
        }
      });

      console.log('[useNoteMentions] Found mentions in document:', mentions);

      // Create links for new mentions
      mentions.forEach(async (mentionId) => {
        const linkKey = `${currentNoteId}-${mentionId}`;
        
        if (mentionId !== currentNoteId && !processedMentions.current.has(linkKey)) {
          processedMentions.current.add(linkKey);
          
          console.log('[useNoteMentions] Creating link:', {
            from: currentNoteId,
            to: mentionId,
            org: activeOrg?.id,
          });

          try {
            await createNoteLink({
              sourceNoteId: currentNoteId,
              targetNoteId: mentionId,
              linkType: 'references',
            }, activeOrg?.id);
            console.log('[useNoteMentions] ✅ Link created successfully');
          } catch (error) {
            console.error('[useNoteMentions] ❌ Failed to create link:', error);
            // Silently fail - link might already exist
          }
        }
      });
    };

    // Listen to editor updates
    editor.on('update', handleUpdate);

    // Run once on mount to catch existing mentions
    handleUpdate();

    return () => {
      editor.off('update', handleUpdate);
      processedMentions.current.clear();
    };
  }, [editor, currentNoteId, activeOrg?.id]);
}

