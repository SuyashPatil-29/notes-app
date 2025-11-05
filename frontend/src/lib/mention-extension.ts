import Mention from '@tiptap/extension-mention';
import { ReactRenderer } from '@tiptap/react';
import tippy from 'tippy.js'
import type { Instance as TippyInstance } from 'tippy.js';
import type { SuggestionOptions } from '@tiptap/suggestion';
import MentionList, { type MentionListRef } from '@/components/MentionList';
import api from '@/utils/api';
import { queryClient } from './query-client';

interface MentionSuggestion {
  id: string;
  label: string;
}

// Cache for notes to show immediately on subsequent searches
let notesCache: MentionSuggestion[] = [];
let cacheOrgId: string | null | undefined = null;
let isLoading = false;

export const createNoteMention = (organizationId?: string | null) => {
  return Mention.configure({
    HTMLAttributes: {
      class: 'mention bg-primary/10 text-primary rounded px-1 cursor-pointer hover:bg-primary/20',
    },
    suggestion: {
      char: '@',
      allowSpaces: true,
    items: async ({ query }: { query: string }): Promise<MentionSuggestion[]> => {
      // Fetch notes from API (only if cache is empty or org changed)
      if (notesCache.length === 0 || cacheOrgId !== organizationId) {
        isLoading = true;
        try {
          // First, try to get from React Query cache for instant loading
          const cachedNotebooks = queryClient.getQueryData<any[]>(['userNotebooks', organizationId]);
          
          let notebooks;
          if (cachedNotebooks) {
            console.log('[Mention] Using cached notebooks from React Query');
            notebooks = cachedNotebooks;
          } else {
            console.log('[Mention] Fetching notebooks for org:', organizationId);
            
            const params: Record<string, string> = {};
            if (organizationId) {
              params.organizationId = organizationId;
            }

            const response = await api.get('/notebooks', { params });
            notebooks = response.data;
          }
          
          console.log('[Mention] Processing notebooks:', notebooks.length, 'notebooks');
          const notes: MentionSuggestion[] = [];

          // Extract all notes from notebooks
          for (const notebook of notebooks) {
            if (notebook.chapters) {
              for (const chapter of notebook.chapters) {
                if (chapter.notes && Array.isArray(chapter.notes)) {
                  console.log(`[Mention] Chapter "${chapter.name}" has ${chapter.notes.length} notes`);
                  for (const note of chapter.notes) {
                    notes.push({
                      id: note.id,
                      label: `${note.name} (${notebook.name} / ${chapter.name})`,
                    });
                  }
                }
              }
            }
          }

          console.log('[Mention] Total notes extracted:', notes.length);
          
          // Update cache
          notesCache = notes;
          cacheOrgId = organizationId;
          isLoading = false;
        } catch (error) {
          console.error('[Mention] Error fetching notes for mention:', error);
          isLoading = false;
          return [];
        }
      }

      // Filter cached notes based on query
      const lowerQuery = query.toLowerCase();
      const filtered = notesCache
        .filter((note) => note.label.toLowerCase().includes(lowerQuery))
        .slice(0, 10);
      
      console.log('[Mention] Filtered to:', filtered.length, 'notes for query:', query);
      return filtered;
    },
    render: () => {
      let component: ReactRenderer;
      let popup: TippyInstance[];

      return {
        onStart: (props: any) => {
          // Start with loading state if cache is empty
          const initialLoading = notesCache.length === 0 || cacheOrgId !== organizationId;
          
          component = new ReactRenderer(MentionList, {
            props: {
              ...props,
              loading: initialLoading,
            },
            editor: props.editor,
          });

          if (!props.clientRect) {
            return;
          }

          popup = tippy('body', {
            getReferenceClientRect: props.clientRect,
            appendTo: () => document.body,
            content: component.element,
            showOnCreate: true,
            interactive: true,
            trigger: 'manual',
            placement: 'bottom-start',
          });
        },

        onUpdate(props: any) {
          component.updateProps({
            ...props,
            loading: isLoading,
          });

          if (!props.clientRect) {
            return;
          }

          popup[0].setProps({
            getReferenceClientRect: props.clientRect,
          });
        },

        onKeyDown(props: any) {
          if (props.event.key === 'Escape') {
            popup[0].hide();
            return true;
          }

          return (component.ref as MentionListRef | null)?.onKeyDown(props);
        },

        onExit() {
          popup[0].destroy();
          component.destroy();
        },
      };
    },
  } as Partial<SuggestionOptions>,
  });
};

