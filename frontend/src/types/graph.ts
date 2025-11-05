export interface GraphNode {
  id: string;
  name: string;
  chapterName?: string;
  notebookName?: string;
  createdAt: string;
  updatedAt: string;
  linkCount: number;
  metadata?: Record<string, string>;
}

export interface GraphLink {
  id: string;
  source: string;
  target: string;
  linkType: string;
}

export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

export interface NoteLink {
  id: string;
  sourceNoteId: string;
  targetNoteId: string;
  linkType: string;
  organizationId?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateNoteLinkRequest {
  sourceNoteId: string;
  targetNoteId: string;
  linkType?: string;
}

export const LinkTypes = {
  REFERENCES: 'references',
  BUILDS_ON: 'builds-on',
  CONTRADICTS: 'contradicts',
  RELATED: 'related',
  PREREQUISITE: 'prerequisite',
} as const;

export type LinkType = typeof LinkTypes[keyof typeof LinkTypes];

