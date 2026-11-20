import React, {
  useCallback,
  useContext,
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
  useReactFlow,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { FlowStepItem } from "./tree";
import { findStep } from "./findStep";
import { treeToGraph } from "./graph";
import { layoutGraph, Size } from "./layout";
import {
  BOX_NODE_TYPE,
  FlowNode,
  boxNodeID,
  flowNodeTypes,
  isBoxNodeID,
} from "./FlowNodes";
import { LaneEdge, flowEdgeTypes } from "./FlowEdges";
import { MeasuredSize, applySizes, splitNodeChanges } from "./nodeChanges";
import { useFlowEditor } from "./edit/FlowEditorContext";
import {
  StepDrag,
  StepDragContext,
  insertionKey,
} from "./edit/StepDragContext";
import { InsertionPoint } from "./edit/pointer";
import { Context } from "../../../intl";
import styles from "./FlowCanvas.module.css";

const MIN_ZOOM = 0.2;
// How long the minimap stays after panning or zooming stops.
const MINIMAP_LINGER_MS = 800;
const PRO_OPTIONS = { hideAttribution: true };
// Laid out for a node not measured yet, which stays hidden until it is.
const UNMEASURED_SIZE: Size = { width: 200, height: 120 };
const HIDDEN_STYLE = { visibility: "hidden" } as const;
// Boxes sit behind lines and cards, and let clicks through to the pane.
const BOX_Z_INDEX = -1;
const BOX_STYLE = { pointerEvents: "none" } as const;
// How near a "+" a dragged card's pointer must be to land on it, in pixels.
const DROP_RADIUS = 48;
// Ignore pointer jitter, so a click still selects the card.
const DRAG_THRESHOLD = 4;
const DRAGGED_Z_INDEX = 1000;

// The "+" nearest the pointer, within DROP_RADIUS on screen.
function dropTargetAt(
  root: HTMLElement,
  event: MouseEvent | TouchEvent
): string | null {
  const point = "changedTouches" in event ? event.changedTouches[0] : event;
  const x = point.clientX;
  const y = point.clientY;
  let best: string | null = null;
  let bestDistance = DROP_RADIUS;
  root.querySelectorAll<HTMLElement>("[data-insert-at]").forEach((el) => {
    const r = el.getBoundingClientRect();
    const d = Math.hypot(r.left + r.width / 2 - x, r.top + r.height / 2 - y);
    if (d < bestDistance) {
      bestDistance = d;
      best = el.dataset.insertAt ?? null;
    }
  });
  return best;
}

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
  // Sizes outlive edits: nodes come from the graph on every render, so an
  // edit relayouts in place and only new or resized nodes are measured.
  const [sizes, setSizes] = useState<ReadonlyMap<string, MeasuredSize>>(
    () => new Map()
  );
  const handleNodesChange = useCallback(
    (changes: NodeChange<FlowNode>[]) => {
      const { measurements, selectedID } = splitNodeChanges(
        // Boxes are sized by the layout, not measured for it.
        changes.filter((c) => !("id" in c && isBoxNodeID(c.id)))
      );
      if (measurements.length > 0) {
        setSizes((prev) => applySizes(prev, measurements));
      }
      if (selectedID !== undefined) {
        const isStep =
          graph.nodes.find((n) => n.id === selectedID)?.item != null;
        onSelectStep(isStep ? selectedID : null);
      }
    },
    [graph, onSelectStep]
  );
  const { fitView } = useReactFlow<FlowNode>();
  const editor = useFlowEditor();
  const editing = editor != null;
  // Where a step sits says which options run it, so a card moves by
  // dragging it onto a "+"; with Alt it is copied.
  const { renderToString } = useContext(Context);
  const rootRef = useRef<HTMLDivElement>(null);
  const [stepDrag, setStepDrag] = useState<
    (StepDrag & { position: { x: number; y: number } }) | null
  >(null);
  const insertionPoints = useMemo(() => {
    const m = new Map<string, InsertionPoint>();
    for (const e of graph.edges) {
      for (const at of [e.insertAt, e.mergeInsertAt]) {
        if (at != null) {
          m.set(insertionKey(at), at);
        }
      }
    }
    return m;
  }, [graph]);
  const onNodeDrag = useCallback(
    (event: MouseEvent | TouchEvent, node: FlowNode) => {
      const root = rootRef.current;
      const key = root != null ? dropTargetAt(root, event) : null;
      setStepDrag({
        stepID: node.id,
        target: key != null ? insertionPoints.get(key) ?? null : null,
        position: node.position,
      });
    },
    [insertionPoints]
  );
  const onNodeDragStop = useCallback(
    (event: MouseEvent | TouchEvent, node: FlowNode) => {
      setStepDrag(null);
      const root = rootRef.current;
      const key = root != null ? dropTargetAt(root, event) : null;
      const target = key != null ? insertionPoints.get(key) : undefined;
      if (editor != null && target != null) {
        editor.onMove(node.id, target, event.altKey);
      }
    },
    [editor, insertionPoints]
  );
  const [hoveredEdgeID, setHoveredEdgeID] = useState<string | null>(null);
  const hoveredTarget =
    graph.edges.find((e) => e.id === hoveredEdgeID)?.target ?? null;

  // The canvas stays hidden until every node of the first graph is measured.
  const allMeasured = graph.nodes.every((n) => sizes.has(n.id));
  const [laidOut, setLaidOut] = useState(false);
  if (allMeasured && !laidOut) {
    setLaidOut(true);
  }
  const layout = useMemo(() => {
    const known = new Map<string, Size>();
    for (const n of graph.nodes) {
      known.set(n.id, sizes.get(n.id) ?? UNMEASURED_SIZE);
    }
    return layoutGraph(tree, known, graph.terminalIDs);
  }, [graph, sizes, tree]);

  const nodes = useMemo(
    () => [
      ...layout.boxes.map(
        (box): FlowNode => ({
          id: boxNodeID(box.id),
          type: BOX_NODE_TYPE,
          position: { x: box.x, y: box.y },
          data: { item: findStep(tree, box.id), flowRef: null, box },
          width: box.width,
          height: box.height,
          zIndex: BOX_Z_INDEX,
          selectable: false,
          focusable: false,
          style: BOX_STYLE,
        })
      ),
      ...graph.nodes.map((n): FlowNode => {
        const measured = sizes.get(n.id);
        const dragged = stepDrag?.stepID === n.id ? stepDrag : null;
        return {
          id: n.id,
          type: n.kind,
          position: dragged?.position ??
            layout.positions.get(n.id) ?? { x: 0, y: 0 },
          draggable: editing && n.item != null,
          zIndex: dragged != null ? DRAGGED_Z_INDEX : undefined,
          data: { item: n.item, flowRef: n.flowRef ?? null },
          measured,
          selected: n.id === selectedStepID,
          style: measured == null ? HIDDEN_STYLE : undefined,
        };
      }),
    ],
    [graph, sizes, layout, selectedStepID, tree, stepDrag, editing]
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
            laneY: e.lane != null ? layout.lanes.get(e.lane) ?? null : null,
            insertAt: editing ? e.insertAt : null,
            hovered: e.id === hoveredEdgeID,
            mergeInsertAt: editing ? e.mergeInsertAt : null,
            afterChoice: e.afterChoice === true,
            fromOption: e.kind === "option",
            targetHovered: e.target === hoveredTarget,
          },
        })
      ),
    [graph, layout, editing, hoveredEdgeID, hoveredTarget]
  );

  return (
    <div
      ref={rootRef}
      className={styles.canvas}
      style={{ visibility: laidOut ? "visible" : "hidden" }}
    >
      {stepDrag != null ? (
        <div className={styles.dragHint}>
          {renderToString("AuthFlowsScreen.edit.move.hint")}
        </div>
      ) : null}
      <StepDragContext.Provider value={stepDrag}>
        <ReactFlow<FlowNode, LaneEdge>
          nodes={nodes}
          edges={edges}
          nodeTypes={flowNodeTypes}
          edgeTypes={flowEdgeTypes}
          onNodesChange={handleNodesChange}
          deleteKeyCode={null}
          nodesDraggable={editing}
          nodeDragThreshold={DRAG_THRESHOLD}
          selectNodesOnDrag={false}
          onNodeDrag={onNodeDrag}
          onNodeDragStop={onNodeDragStop}
          nodesConnectable={false}
          edgesFocusable={false}
          onPaneClick={() => onSelectStep(null)}
          onEdgeMouseEnter={(_, edge) => setHoveredEdgeID(edge.id)}
          onEdgeMouseLeave={() => setHoveredEdgeID(null)}
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
            className={cn(
              styles.minimap,
              moving ? styles.minimapVisible : null
            )}
          />
        </ReactFlow>
      </StepDragContext.Provider>
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
