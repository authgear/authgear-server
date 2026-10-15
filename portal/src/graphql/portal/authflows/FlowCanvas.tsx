import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import {
  Background,
  NodeChange,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  useNodesState,
  useReactFlow,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { FlowStepItem } from "./tree";
import { treeToGraph } from "./graph";
import { layoutGraph, Size } from "./layout";
import { FlowNode, flowNodeTypes } from "./FlowNodes";
import { LaneEdge, flowEdgeTypes } from "./FlowEdges";
import { splitNodeChanges } from "./nodeChanges";
import styles from "./FlowCanvas.module.css";

const MIN_ZOOM = 0.2;
// How long the minimap stays after panning or zooming stops.
const MINIMAP_LINGER_MS = 800;
const PRO_OPTIONS = { hideAttribution: true };

export interface FlowCanvasProps {
  tree: FlowStepItem[];
  selectedStepID: string | null;
  onSelectStep: (id: string | null) => void;
}

function FlowCanvasInner({
  tree,
  selectedStepID,
  onSelectStep,
}: FlowCanvasProps): React.ReactElement {
  const graph = useMemo(() => treeToGraph(tree), [tree]);
  const [nodes, , onNodesChange] = useNodesState<FlowNode>(
    graph.nodes.map((n) => ({
      id: n.id,
      type: n.kind,
      position: { x: 0, y: 0 },
      data: { item: n.item, flowRef: n.flowRef ?? null },
    }))
  );
  const handleNodesChange = useCallback(
    (changes: NodeChange<FlowNode>[]) => {
      const { measurements, selectedID } = splitNodeChanges(changes);
      if (measurements.length > 0) {
        onNodesChange(measurements);
      }
      if (selectedID !== undefined) {
        const isStep =
          graph.nodes.find((n) => n.id === selectedID)?.item != null;
        onSelectStep(isStep ? selectedID : null);
      }
    },
    [graph, onNodesChange, onSelectStep]
  );
  const { fitView } = useReactFlow<FlowNode>();

  // xyflow measures nodes after the first render; until every node has a size,
  // the canvas stays hidden. Layout is derived, never stored.
  const layout = useMemo(() => {
    const sizes = new Map<string, Size>();
    for (const n of nodes) {
      if (n.measured?.width == null || n.measured.height == null) {
        return null;
      }
      sizes.set(n.id, { width: n.measured.width, height: n.measured.height });
    }
    return layoutGraph(tree, sizes, graph.terminalIDs);
  }, [nodes, tree, graph]);
  const laidOut = layout != null;

  const displayNodes = useMemo(
    () =>
      nodes.map((n) => ({
        ...n,
        position: layout?.positions.get(n.id) ?? n.position,
        selected: n.id === selectedStepID,
      })),
    [nodes, layout, selectedStepID]
  );

  // The minimap only helps while moving, so it shows then and fades after.
  const [moving, setMoving] = useState(false);
  const minimapTimer = useRef<number | null>(null);
  const onMoveStart = useCallback(() => {
    if (minimapTimer.current != null) {
      window.clearTimeout(minimapTimer.current);
      minimapTimer.current = null;
    }
    setMoving(true);
  }, []);
  const onMoveEnd = useCallback(() => {
    minimapTimer.current = window.setTimeout(() => {
      minimapTimer.current = null;
      setMoving(false);
    }, MINIMAP_LINGER_MS);
  }, []);
  useEffect(
    () => () => {
      if (minimapTimer.current != null) {
        window.clearTimeout(minimapTimer.current);
      }
    },
    []
  );

  const fitted = useRef(false);
  useEffect(() => {
    if (laidOut && !fitted.current) {
      fitted.current = true;
      fitView().catch((e: unknown) => {
        console.error(e);
      });
    }
  }, [laidOut, fitView]);

  const edges = useMemo(
    () =>
      graph.edges.map(
        (e): LaneEdge => ({
          id: e.id,
          source: e.source,
          sourceHandle: e.sourceHandle,
          target: e.target,
          targetHandle: e.targetHandle,
          type: "lane",
          data: {
            laneY: e.lane != null ? layout?.lanes.get(e.lane) ?? null : null,
          },
        })
      ),
    [graph, layout]
  );

  return (
    <div
      className={styles.canvas}
      style={{ visibility: laidOut ? "visible" : "hidden" }}
    >
      <ReactFlow<FlowNode, LaneEdge>
        nodes={displayNodes}
        edges={edges}
        nodeTypes={flowNodeTypes}
        edgeTypes={flowEdgeTypes}
        onNodesChange={handleNodesChange}
        deleteKeyCode={null}
        nodesDraggable={false}
        nodesConnectable={false}
        edgesFocusable={false}
        onPaneClick={() => onSelectStep(null)}
        minZoom={MIN_ZOOM}
        onMoveStart={onMoveStart}
        onMoveEnd={onMoveEnd}
        proOptions={PRO_OPTIONS}
      >
        <Background />
        <Controls showInteractive={false} />
        <MiniMap
          pannable={true}
          zoomable={true}
          className={cn(styles.minimap, moving ? styles.minimapVisible : null)}
        />
      </ReactFlow>
    </div>
  );
}

export function FlowCanvas(props: FlowCanvasProps): React.ReactElement {
  return (
    <ReactFlowProvider>
      <FlowCanvasInner {...props} />
    </ReactFlowProvider>
  );
}
