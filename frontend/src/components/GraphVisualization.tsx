import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import ForceGraph2D, {type ForceGraphMethods, type NodeObject, type LinkObject } from 'react-force-graph-2d';
import type { GraphData, GraphLink } from '@/types/graph';
import { LinkTypes } from '@/types/graph';
import { getGraphData, deleteNoteLink, updateNoteLink } from '@/utils/graphApi';
import { useNavigate } from 'react-router-dom';
import { useOrganizationContext } from '@/contexts/OrganizationContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { Loader2, Search, X, ZoomIn, ZoomOut, Maximize2 } from 'lucide-react';

interface GraphVisualizationProps {
  onNodeClick?: (nodeId: string, metadata?: Record<string, string>) => void;
  centerNodeId?: string;
  height?: number;
  width?: number;
}

interface ExtendedNodeObject extends NodeObject {
  id: string;
  name: string;
  linkCount?: number;
  notebookName?: string;
  chapterName?: string;
  metadata?: Record<string, string>;
}

interface ExtendedLinkObject extends LinkObject {
  id: string;
  linkType: string;
  source: string | ExtendedNodeObject;
  target: string | ExtendedNodeObject;
}

const GraphVisualization: React.FC<GraphVisualizationProps> = ({
  onNodeClick,
  centerNodeId,
}) => {
  const [graphData, setGraphData] = useState<GraphData | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  const [selectedLink, setSelectedLink] = useState<GraphLink | null>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const containerRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const graphRef = useRef<ForceGraphMethods | undefined>(undefined);
  const { activeOrg } = useOrganizationContext();

  // Measure container size
  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const { width, height } = containerRef.current.getBoundingClientRect();
        // Only update if dimensions are valid (not zero)
        if (width > 0 && height > 0) {
          console.log('[GraphVisualization] Container dimensions:', { width, height });
          setDimensions({ width, height });
        }
      }
    };

    // Initial measurement with a small delay to ensure DOM is ready
    const timer = setTimeout(updateDimensions, 100);
    
    window.addEventListener('resize', updateDimensions);
    
    // Use ResizeObserver for better responsiveness
    const resizeObserver = new ResizeObserver(() => {
      // Debounce resize updates
      setTimeout(updateDimensions, 50);
    });
    
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current);
    }

    return () => {
      clearTimeout(timer);
      window.removeEventListener('resize', updateDimensions);
      resizeObserver.disconnect();
    };
  }, []);

  // Fetch graph data
  const fetchGraphData = useCallback(async () => {
    try {
      setLoading(true);
      console.log('[GraphVisualization] Fetching graph data for org:', activeOrg?.id);
      const data = await getGraphData(searchQuery, activeOrg?.id);
      console.log('[GraphVisualization] Received graph data:', {
        nodeCount: data.nodes.length,
        linkCount: data.links.length,
        nodes: data.nodes,
        links: data.links,
      });
      setGraphData(data);
    } catch (error) {
      console.error('[GraphVisualization] Failed to fetch graph data:', error);
      toast.error('Failed to load graph data');
    } finally {
      setLoading(false);
    }
  }, [searchQuery, activeOrg?.id]);

  useEffect(() => {
    fetchGraphData();
  }, [fetchGraphData]);

  // Center on specific node if provided
  useEffect(() => {
    if (centerNodeId && graphData && graphRef.current) {
      const node = graphData.nodes.find((n) => n.id === centerNodeId);
      if (node) {
        graphRef.current.centerAt(0, 0, 1000);
        graphRef.current.zoom(2, 1000);
      }
    }
  }, [centerNodeId, graphData]);

  // Prepare data for force graph
  const forceGraphData = useMemo(() => {
    if (!graphData) return { nodes: [], links: [] };

    const nodes = graphData.nodes.map((node) => ({
      id: node.id,
      name: node.name,
      linkCount: node.linkCount,
      notebookName: node.notebookName,
      chapterName: node.chapterName,
      metadata: node.metadata,
    }));

    const links = graphData.links.map((link) => ({
      id: link.id,
      source: link.source,
      target: link.target,
      linkType: link.linkType,
    }));

    return { nodes, links };
  }, [graphData]);

  // Handle node click
  const handleNodeClick = useCallback(
    (node: NodeObject) => {
      const extNode = node as ExtendedNodeObject;
      if (onNodeClick) {
        onNodeClick(extNode.id, extNode.metadata);
      } else {
        navigate(`/note/${extNode.id}`);
      }
    },
    [onNodeClick, navigate]
  );

  // Handle link right-click
  const handleLinkRightClick = useCallback((link: LinkObject) => {
    const extLink = link as ExtendedLinkObject;
    setSelectedLink({
      id: extLink.id,
      source: typeof extLink.source === 'string' ? extLink.source : extLink.source.id!,
      target: typeof extLink.target === 'string' ? extLink.target : extLink.target.id!,
      linkType: extLink.linkType,
    });
  }, []);

  // Delete link
  const handleDeleteLink = useCallback(async () => {
    if (!selectedLink) return;

    try {
      await deleteNoteLink(selectedLink.id);
      toast.success('Link deleted successfully');
      fetchGraphData();
      setSelectedLink(null);
    } catch (error) {
      console.error('Failed to delete link:', error);
      toast.error('Failed to delete link');
    }
  }, [selectedLink, fetchGraphData]);

  // Update link type
  const handleUpdateLinkType = useCallback(
    async (newLinkType: string) => {
      if (!selectedLink) return;

      try {
        await updateNoteLink(selectedLink.id, newLinkType);
        toast.success('Link updated successfully');
        fetchGraphData();
        setSelectedLink(null);
      } catch (error) {
        console.error('Failed to update link:', error);
        toast.error('Failed to update link');
      }
    },
    [selectedLink, fetchGraphData]
  );

  // Node paint function
  const paintNode = useCallback(
    (node: NodeObject, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const extNode = node as ExtendedNodeObject;
      const label = extNode.name || '';
      const fontSize = 12 / globalScale;
      const nodeSize = Math.sqrt((extNode.linkCount || 1) * 4) + 4;

      // Determine node color based on notebook
      let nodeColor = '#3b82f6'; // Default blue
      if (extNode.notebookName) {
        const hash = extNode.notebookName.split('').reduce((acc, char) => {
          return char.charCodeAt(0) + ((acc << 5) - acc);
        }, 0);
        const hue = Math.abs(hash % 360);
        nodeColor = `hsl(${hue}, 70%, 60%)`;
      }

      // Highlight if hovered
      if (hoveredNode === extNode.id) {
        ctx.beginPath();
        ctx.arc(node.x || 0, node.y || 0, nodeSize + 2, 0, 2 * Math.PI);
        ctx.fillStyle = 'rgba(255, 255, 255, 0.3)';
        ctx.fill();
      }

      // Draw node
      ctx.beginPath();
      ctx.arc(node.x || 0, node.y || 0, nodeSize, 0, 2 * Math.PI);
      ctx.fillStyle = nodeColor;
      ctx.fill();
      ctx.strokeStyle = '#fff';
      ctx.lineWidth = 1 / globalScale;
      ctx.stroke();

      // Draw label
      ctx.font = `${fontSize}px Sans-Serif`;
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillStyle = '#fff';
      ctx.fillText(label, node.x || 0, (node.y || 0) + nodeSize + fontSize);
    },
    [hoveredNode]
  );

  // Link paint function
  const paintLink = useCallback(
    (link: LinkObject, ctx: CanvasRenderingContext2D, globalScale: number) => {
      const extLink = link as ExtendedLinkObject;
      const sourceNode =
        typeof extLink.source === 'string'
          ? null
          : (extLink.source as ExtendedNodeObject);
      const targetNode =
        typeof extLink.target === 'string'
          ? null
          : (extLink.target as ExtendedNodeObject);

      if (!sourceNode || !targetNode) return;

      const start = { x: sourceNode.x || 0, y: sourceNode.y || 0 };
      const end = { x: targetNode.x || 0, y: targetNode.y || 0 };

      // Link color based on type
      let linkColor = '#999';
      switch (extLink.linkType) {
        case LinkTypes.BUILDS_ON:
          linkColor = '#10b981';
          break;
        case LinkTypes.CONTRADICTS:
          linkColor = '#ef4444';
          break;
        case LinkTypes.PREREQUISITE:
          linkColor = '#f59e0b';
          break;
        case LinkTypes.RELATED:
          linkColor = '#8b5cf6';
          break;
      }

      // Draw link
      ctx.beginPath();
      ctx.moveTo(start.x, start.y);
      ctx.lineTo(end.x, end.y);
      ctx.strokeStyle = linkColor;
      ctx.lineWidth = 1 / globalScale;
      ctx.stroke();
    },
    []
  );

  // Zoom controls
  const handleZoomIn = () => {
    if (graphRef.current) {
      const currentZoom = graphRef.current.zoom();
      graphRef.current.zoom(currentZoom * 1.2, 300);
    }
  };

  const handleZoomOut = () => {
    if (graphRef.current) {
      const currentZoom = graphRef.current.zoom();
      graphRef.current.zoom(currentZoom / 1.2, 300);
    }
  };

  const handleFitView = () => {
    if (graphRef.current) {
      graphRef.current.zoomToFit(400, 50);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-3">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading graph data...</p>
      </div>
    );
  }

  if (!graphData || graphData.nodes.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center p-8">
        <p className="text-lg font-medium text-muted-foreground mb-2">No linked notes yet</p>
        <p className="text-sm text-muted-foreground">
          Start creating links between your notes to see them visualized here
        </p>
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative w-full h-full">
      {/* Search and controls */}
      <div className="absolute top-4 left-4 z-10 flex gap-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search notes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 pr-9 w-64 bg-background/95 backdrop-blur"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-3 top-1/2 transform -translate-y-1/2"
            >
              <X className="w-4 h-4 text-muted-foreground hover:text-foreground" />
            </button>
          )}
        </div>
      </div>

      {/* Zoom controls */}
      <div className="absolute top-4 right-4 z-10 flex flex-col gap-2">
        <Button size="icon" variant="secondary" onClick={handleZoomIn}>
          <ZoomIn className="w-4 h-4" />
        </Button>
        <Button size="icon" variant="secondary" onClick={handleZoomOut}>
          <ZoomOut className="w-4 h-4" />
        </Button>
        <Button size="icon" variant="secondary" onClick={handleFitView}>
          <Maximize2 className="w-4 h-4" />
        </Button>
      </div>

      {/* Graph info */}
      <div className="absolute bottom-4 left-4 z-10 bg-background/95 backdrop-blur px-3 py-2 rounded-lg text-sm">
        <span className="text-muted-foreground">
          {graphData.nodes.length} notes · {graphData.links.length} links
        </span>
      </div>

      {/* Force graph */}
      <ContextMenu>
        <ContextMenuTrigger>
          <ForceGraph2D
            ref={graphRef}
            graphData={forceGraphData}
            nodeLabel={(node) => (node as ExtendedNodeObject).name}
            nodeCanvasObject={paintNode}
            linkCanvasObject={paintLink}
            onNodeClick={handleNodeClick}
            onNodeHover={(node) => setHoveredNode(node ? (node as ExtendedNodeObject).id : null)}
            onLinkRightClick={handleLinkRightClick}
            cooldownTicks={100}
            d3AlphaDecay={0.02}
            d3VelocityDecay={0.3}
            height={dimensions.height}
            width={dimensions.width}
          />
        </ContextMenuTrigger>
        {selectedLink && (
          <ContextMenuContent>
            <ContextMenuItem onClick={handleDeleteLink}>Delete Link</ContextMenuItem>
            <ContextMenuItem onClick={() => handleUpdateLinkType(LinkTypes.REFERENCES)}>
              Set as References
            </ContextMenuItem>
            <ContextMenuItem onClick={() => handleUpdateLinkType(LinkTypes.BUILDS_ON)}>
              Set as Builds On
            </ContextMenuItem>
            <ContextMenuItem onClick={() => handleUpdateLinkType(LinkTypes.CONTRADICTS)}>
              Set as Contradicts
            </ContextMenuItem>
            <ContextMenuItem onClick={() => handleUpdateLinkType(LinkTypes.RELATED)}>
              Set as Related
            </ContextMenuItem>
            <ContextMenuItem onClick={() => handleUpdateLinkType(LinkTypes.PREREQUISITE)}>
              Set as Prerequisite
            </ContextMenuItem>
          </ContextMenuContent>
        )}
      </ContextMenu>
    </div>
  );
};

export default GraphVisualization;

