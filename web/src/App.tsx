import { useEffect, useMemo, useRef, useState } from "react";
import ReactFlow, {
  BaseEdge,
  EdgeLabelRenderer,
  Handle,
  Position,
  getSmoothStepPath,
  type Edge,
  type EdgeProps,
  type Node,
  type NodeProps,
} from "reactflow";
import ReactMarkdown from "react-markdown";
import { ArrowDownToLine, ArrowUpToLine, CircleCheck, ShieldAlert, X } from "lucide-react";
import remarkGfm from "remark-gfm";
import { Button } from "./components/ui/button";
import { layoutReportGraph } from "./lib/graph-layout";
import { formatDate, loadReport } from "./lib/report";
import type { FeodReport, ReportEdge, ReportNode, Violation } from "./types";

type EntityNodeData = {
  node: ReportNode;
  incoming: number;
  outgoing: number;
  violations: Violation[];
};

type LaneNodeData = {
  label: string;
  count: number;
  trackCount: number;
};

type HoveredDependency = {
  id: string;
  source: string;
  target: string;
};

type NodeStats = {
  incoming: number;
  outgoing: number;
  issues: number;
};

const nodeTypes = {
  entity: EntityNode,
  lane: LaneNode,
};

const edgeTypes = {
  dependency: DependencyEdge,
};

const scrollEdgeThreshold = 180;

export function App() {
  const [report, setReport] = useState<FeodReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [graphNodes, setGraphNodes] = useState<Node[]>([]);
  const [graphEdges, setGraphEdges] = useState<Edge[]>([]);
  const [hoveredDependency, setHoveredDependency] = useState<HoveredDependency | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [isAtPageBottom, setIsAtPageBottom] = useState(false);
  const hoverClearTimerRef = useRef<number | null>(null);
  const pointerDownRef = useRef(false);
  const violationsRef = useRef<HTMLElement | null>(null);

  const setHoveredDependencyStable = (next: HoveredDependency | null) => {
    if (hoverClearTimerRef.current !== null) {
      window.clearTimeout(hoverClearTimerRef.current);
      hoverClearTimerRef.current = null;
    }
    setHoveredDependency((current) => {
      if (current?.id === next?.id && current?.source === next?.source && current?.target === next?.target) {
        return current;
      }
      return next;
    });
  };

  const clearHoveredDependencySoon = () => {
    if (hoverClearTimerRef.current !== null) {
      window.clearTimeout(hoverClearTimerRef.current);
    }
    hoverClearTimerRef.current = window.setTimeout(() => {
      setHoveredDependency(null);
      hoverClearTimerRef.current = null;
    }, 90);
  };

  useEffect(() => {
    loadReport()
      .then((loaded) => {
        const graph = layoutReportGraph(loaded);
        setReport(loaded);
        setError(null);
        setGraphNodes(graph.nodes);
        setGraphEdges(graph.edges);
        setHoveredDependencyStable(null);
        setSelectedNodeId(null);
      })
      .catch((reason: unknown) => setError(reason instanceof Error ? reason.message : String(reason)))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    return () => {
      if (hoverClearTimerRef.current !== null) {
        window.clearTimeout(hoverClearTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    const releasePointerGuard = () => {
      window.setTimeout(() => {
        pointerDownRef.current = false;
      }, 80);
    };
    const handlePointerDown = (event: PointerEvent) => {
      if (event.button === 0) {
        pointerDownRef.current = true;
      }
    };

    window.addEventListener("pointerdown", handlePointerDown, true);
    window.addEventListener("pointerup", releasePointerGuard, true);
    window.addEventListener("pointercancel", releasePointerGuard, true);

    return () => {
      window.removeEventListener("pointerdown", handlePointerDown, true);
      window.removeEventListener("pointerup", releasePointerGuard, true);
      window.removeEventListener("pointercancel", releasePointerGuard, true);
    };
  }, []);

  useEffect(() => {
    let animationFrame = 0;
    const updateScrollState = () => {
      window.cancelAnimationFrame(animationFrame);
      animationFrame = window.requestAnimationFrame(() => {
        const scrollRoot = document.documentElement;
        const maxScroll = Math.max(0, scrollRoot.scrollHeight - window.innerHeight);
        setIsAtPageBottom(maxScroll > scrollEdgeThreshold && window.scrollY >= maxScroll - scrollEdgeThreshold);
      });
    };
    const resizeObserver = new ResizeObserver(updateScrollState);

    const handleScroll = () => {
      const scrollRoot = document.documentElement;
      const maxScroll = Math.max(0, scrollRoot.scrollHeight - window.innerHeight);
      setIsAtPageBottom(maxScroll > scrollEdgeThreshold && window.scrollY >= maxScroll - scrollEdgeThreshold);
    };

    updateScrollState();
    resizeObserver.observe(document.body);
    window.addEventListener("scroll", handleScroll, { passive: true });
    window.addEventListener("resize", updateScrollState);

    return () => {
      window.cancelAnimationFrame(animationFrame);
      resizeObserver.disconnect();
      window.removeEventListener("scroll", handleScroll);
      window.removeEventListener("resize", updateScrollState);
    };
  }, []);

  useEffect(() => {
    if (!hoveredDependency) {
      return;
    }

    const clearWhenPointerLeavesEdge = (event: PointerEvent) => {
      if (pointerDownRef.current) {
        return;
      }
      const target = event.target;
      const edgeElement = target instanceof Element ? target.closest(".dependency-edge") : null;
      if (edgeElement?.getAttribute("data-edge-id") !== hoveredDependency.id) {
        clearHoveredDependencySoon();
      }
    };

    const clearHover = () => setHoveredDependencyStable(null);
    window.addEventListener("pointermove", clearWhenPointerLeavesEdge, true);
    window.addEventListener("blur", clearHover);

    return () => {
      window.removeEventListener("pointermove", clearWhenPointerLeavesEdge, true);
      window.removeEventListener("blur", clearHover);
    };
  }, [hoveredDependency]);

  const displayNodes = useMemo(() => {
    return graphNodes.map((node) => {
      const classes = [node.className];
      if (node.id === selectedNodeId) {
        classes.push("is-selected");
      }
      if (!hoveredDependency || node.type === "lane") {
        return { ...node, className: classes.filter(Boolean).join(" ") };
      }
      const isRelated = node.id === hoveredDependency.source || node.id === hoveredDependency.target;
      const stateClass = isRelated ? "is-highlighted" : "is-dimmed";
      classes.push(stateClass);
      return { ...node, className: classes.filter(Boolean).join(" ") };
    });
  }, [graphNodes, hoveredDependency, selectedNodeId]);

  const displayEdges = useMemo(() => {
    return graphEdges.map((edge) => {
      const isActive = hoveredDependency?.id === edge.id;
      const isDimmed = Boolean(hoveredDependency && !isActive);
      return {
        ...edge,
        data: {
          ...(edge.data ?? {}),
          dimmed: isDimmed,
          highlighted: isActive,
          onHover: (value: ReportEdge | null) => {
            if (pointerDownRef.current) {
              return;
            }
            if (value) {
              setHoveredDependencyStable({ id: value.id, source: value.source, target: value.target });
            } else {
              clearHoveredDependencySoon();
            }
          },
        },
      };
    });
  }, [graphEdges, hoveredDependency]);

  const selectedNodeDetails = useMemo(() => {
    if (!report || !selectedNodeId) {
      return null;
    }
    const node = report.nodes.find((item) => item.id === selectedNodeId);
    if (!node) {
      return null;
    }
    const nodeById = new Map(report.nodes.map((item) => [item.id, item]));
    const incoming = report.edges.filter((edge) => edge.target === selectedNodeId);
    const outgoing = report.edges.filter((edge) => edge.source === selectedNodeId);
    const violations = report.violations.filter((violation) => violation.from === selectedNodeId || violation.to === selectedNodeId);
    const statsByNodeId = new Map<string, NodeStats>();

    for (const item of report.nodes) {
      statsByNodeId.set(item.id, { incoming: 0, outgoing: 0, issues: 0 });
    }
    for (const edge of report.edges) {
      const sourceStats = statsByNodeId.get(edge.source);
      const targetStats = statsByNodeId.get(edge.target);
      if (sourceStats) {
        sourceStats.outgoing += 1;
      }
      if (targetStats) {
        targetStats.incoming += 1;
      }
    }
    for (const violation of report.violations) {
      for (const id of [violation.from, violation.to]) {
        if (!id) {
          continue;
        }
        const stats = statsByNodeId.get(id);
        if (stats) {
          stats.issues += 1;
        }
      }
    }

    const submodules = node.kind === "module" ? report.nodes.filter((item) => item.kind === "submodule" && item.parentId === node.id) : [];
    const parentModule = node.kind === "submodule" && node.parentId ? nodeById.get(node.parentId) : undefined;
    return { node, nodeById, incoming, outgoing, violations, submodules, parentModule, statsByNodeId };
  }, [report, selectedNodeId]);

  const handleScrollJump = () => {
    if (isAtPageBottom) {
      window.scrollTo({ top: 0, behavior: "smooth" });
      return;
    }
    violationsRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  };

  if (loading) {
    return <StateShell title="FEOD Analyzer" description="Загружаю feod-report.json..." />;
  }

  if (error || !report) {
    return (
      <StateShell
        title="Не удалось открыть отчёт"
        description={error ?? "JSON report is missing."}
        action="Сгенерируйте отчёт командой feod-analyzer analyze --formats html,json"
      />
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand">
          <img className="brand-logo" src="./feod-logo.svg" alt="FEOD" />
          <div>
            <h1>FEOD Analyzer</h1>
            <p>{report.meta.rootDir}</p>
          </div>
        </div>
        <div className="topbar-actions">
          <span className="run-meta">
            {report.summary.errors} errors, {report.summary.warnings} warnings, {report.summary.edges} edges
          </span>
          <span className="run-meta">{formatDate(report.meta.generated)}</span>
          <Button variant="outline" onClick={() => window.open("./feod-report.json", "_blank")}>
            <ArrowDownToLine data-icon="inline-start" />
            JSON
          </Button>
        </div>
      </header>

      <section className="graph-panel">
        <div className="graph-toolbar">
          <div>
            <h2>Dependency graph</h2>
          </div>
        </div>
        <div className="graph-canvas">
          <ReactFlow
            nodes={displayNodes}
            edges={displayEdges}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            fitView
            minZoom={0.35}
            maxZoom={1.4}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            selectionOnDrag={false}
            selectNodesOnDrag={false}
            nodesFocusable={false}
            edgesFocusable={false}
            selectionKeyCode={null}
            multiSelectionKeyCode={null}
            deleteKeyCode={null}
            onNodeClick={(_, node) => {
              if (node.type === "entity") {
                setSelectedNodeId(String(node.id));
              }
            }}
            onPaneClick={() => setSelectedNodeId(null)}
            onMoveStart={() => {
              pointerDownRef.current = true;
            }}
            onPointerDown={(event) => {
              if (event.button === 0) {
                pointerDownRef.current = true;
              }
            }}
            onPointerUp={() => {
              pointerDownRef.current = false;
            }}
            onPointerCancel={() => {
              pointerDownRef.current = false;
            }}
          />
        </div>
        {selectedNodeDetails && (
          <NodeInspector
            details={selectedNodeDetails}
            onSelectNode={setSelectedNodeId}
            onClose={() => setSelectedNodeId(null)}
          />
        )}
      </section>

      <section className="violations-panel" ref={violationsRef}>
        <div className="violations-header">
          <h2>Ошибки и предупреждения</h2>
          <p>
            {report.summary.errors} errors, {report.summary.warnings} warnings
          </p>
        </div>
        <ViolationsList violations={report.violations} />
      </section>
      <Button
        className="scroll-jump-button"
        size="icon"
        variant="outline"
        aria-label={isAtPageBottom ? "Прокрутить наверх" : "Прокрутить к ошибкам и предупреждениям"}
        title={isAtPageBottom ? "Наверх" : "К ошибкам и предупреждениям"}
        onClick={handleScrollJump}
      >
        {isAtPageBottom ? <ArrowUpToLine /> : <ArrowDownToLine />}
      </Button>
    </main>
  );
}

function EntityNode({ data }: NodeProps<EntityNodeData>) {
  const node = data.node;
  const hasErrors = data.violations.some((violation) => violation.severity === "error");
  const hasWarnings = data.violations.some((violation) => violation.severity === "warning");
  const tone = hasErrors ? "error" : hasWarnings ? "warning" : "allowed";

  return (
    <div className={`entity-node ${tone} ${node.kind}`}>
      <Handle id="target-top" type="target" position={Position.Top} />
      <Handle id="target-left" type="target" position={Position.Left} />
      <Handle id="target-right" type="target" position={Position.Right} />
      <Handle id="source-bottom" type="source" position={Position.Bottom} />
      <Handle id="source-left" type="source" position={Position.Left} />
      <Handle id="source-right" type="source" position={Position.Right} />
      <div className="entity-node-header">
        <span className="entity-kind">{kindLabel(node.kind)}</span>
        <span>{node.fileCount} files</span>
      </div>
      <strong>{node.name}</strong>
      <span className="entity-path">{node.path}</span>
      <div className="entity-node-footer">
        <span>in {data.incoming}</span>
        <span>out {data.outgoing}</span>
        {data.violations.length > 0 && <span>{data.violations.length} issues</span>}
      </div>
    </div>
  );
}

function LaneNode({ data }: NodeProps<LaneNodeData>) {
  return (
    <div className="lane-node">
      <div className="lane-title">
        <span>{data.label}</span>
        <small>{data.count} nodes</small>
      </div>
      <div className="lane-tracks" style={{ gridTemplateColumns: `repeat(${data.trackCount}, 430px)` }} />
    </div>
  );
}

function NodeInspector({
  details,
  onSelectNode,
  onClose,
}: {
  details: {
    node: ReportNode;
    nodeById: Map<string, ReportNode>;
    incoming: ReportEdge[];
    outgoing: ReportEdge[];
    violations: Violation[];
    submodules: ReportNode[];
    parentModule?: ReportNode;
    statsByNodeId: Map<string, NodeStats>;
  };
  onSelectNode: (nodeId: string) => void;
  onClose: () => void;
}) {
  const { node, nodeById, incoming, outgoing, violations, submodules, parentModule, statsByNodeId } = details;
  const sortedViolations = [...violations].sort((a, b) => severityWeight(a.severity) - severityWeight(b.severity));

  return (
    <aside className="node-inspector" aria-label="Информация о блоке">
      <div className="node-inspector-header">
        <div>
          <span>{kindLabel(node.kind)}</span>
          <h3>{node.name}</h3>
          <p>{node.path}</p>
        </div>
        <Button variant="ghost" size="icon" aria-label="Закрыть информацию о блоке" onClick={onClose}>
          <X />
        </Button>
      </div>
      <div className="node-inspector-summary">
        <span>{outgoing.length} dependencies</span>
        <span>{incoming.length} dependents</span>
        <span>{violations.length} issues</span>
      </div>
      {(node.kind === "module" || node.kind === "submodule") && <NodeReadmeSection node={node} />}
      {node.kind === "module" && (
        <NodeSubmodulesSection
          submodules={submodules}
          statsByNodeId={statsByNodeId}
          onSelectNode={onSelectNode}
        />
      )}
      {node.kind === "submodule" && parentModule && (
        <NodeParentModuleSection
          parentModule={parentModule}
          stats={statsByNodeId.get(parentModule.id)}
          onSelectNode={onSelectNode}
        />
      )}
      <NodeDependencySection
        title="Зависимости"
        empty="Нет исходящих зависимостей."
        edges={outgoing}
        direction="outgoing"
        nodeById={nodeById}
      />
      <NodeDependencySection
        title="Зависимые"
        empty="Нет входящих зависимостей."
        edges={incoming}
        direction="incoming"
        nodeById={nodeById}
      />
      <section className="node-inspector-section">
        <h4>Ошибки и предупреждения</h4>
        {sortedViolations.length === 0 ? (
          <p className="node-inspector-empty">Нарушений не найдено.</p>
        ) : (
          <ol className="node-violations-list">
            {sortedViolations.map((violation, index) => (
              <li
                className={`node-violation ${violation.severity}`}
                key={`${violation.rule}:${violation.file ?? ""}:${violation.line ?? ""}:${index}`}
              >
                <div>
                  <span>{violation.severity}</span>
                  <code>{violation.rule}</code>
                </div>
                <p>{violation.message}</p>
                {(violation.file || violation.importPath || violation.from || violation.to) && (
                  <small>
                    {violation.file ? `${violation.file}${violation.line ? `:${violation.line}` : ""}` : ""}
                    {violation.file && violation.importPath ? " · " : ""}
                    {violation.importPath ?? ""}
                    {(violation.file || violation.importPath) && (violation.from || violation.to) ? " · " : ""}
                    {[violation.from, violation.to].filter(Boolean).join(" -> ")}
                  </small>
                )}
              </li>
            ))}
          </ol>
        )}
      </section>
    </aside>
  );
}

function NodeReadmeSection({ node }: { node: ReportNode }) {
  return (
    <section className="node-inspector-section">
      <h4>README.md</h4>
      {node.readme ? (
        <div className="node-readme">
          <div className="node-readme-path">{node.readme.path}</div>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{node.readme.content}</ReactMarkdown>
        </div>
      ) : (
        <p className="node-inspector-empty">README.md не найден.</p>
      )}
    </section>
  );
}

function NodeSubmodulesSection({
  submodules,
  statsByNodeId,
  onSelectNode,
}: {
  submodules: ReportNode[];
  statsByNodeId: Map<string, NodeStats>;
  onSelectNode: (nodeId: string) => void;
}) {
  return (
    <section className="node-inspector-section">
      <h4>Сабмодули</h4>
      {submodules.length === 0 ? (
        <p className="node-inspector-empty">Сабмодули не найдены.</p>
      ) : (
        <div className="node-card-list">
          {submodules.map((submodule) => (
            <NodeRelationCard
              key={submodule.id}
              node={submodule}
              stats={statsByNodeId.get(submodule.id)}
              onSelectNode={onSelectNode}
            />
          ))}
        </div>
      )}
    </section>
  );
}

function NodeParentModuleSection({
  parentModule,
  stats,
  onSelectNode,
}: {
  parentModule: ReportNode;
  stats?: NodeStats;
  onSelectNode: (nodeId: string) => void;
}) {
  return (
    <section className="node-inspector-section">
      <h4>Родительский модуль</h4>
      <div className="node-card-list">
        <NodeRelationCard node={parentModule} stats={stats} onSelectNode={onSelectNode} />
      </div>
    </section>
  );
}

function NodeRelationCard({
  node,
  stats,
  onSelectNode,
}: {
  node: ReportNode;
  stats?: NodeStats;
  onSelectNode: (nodeId: string) => void;
}) {
  return (
    <button className={`node-relation-card ${node.kind}`} type="button" onClick={() => onSelectNode(node.id)}>
      <div className="node-relation-card-main">
        <div>
          <strong>{node.name}</strong>
          <small>{node.path}</small>
        </div>
        {node.readme && <span className="node-readme-badge">README</span>}
      </div>
      <div className="node-relation-card-stats">
        <span>{node.fileCount} files</span>
        <span>in {stats?.incoming ?? 0}</span>
        <span>out {stats?.outgoing ?? 0}</span>
        <span>{stats?.issues ?? 0} issues</span>
      </div>
    </button>
  );
}

function NodeDependencySection({
  title,
  empty,
  edges,
  direction,
  nodeById,
}: {
  title: string;
  empty: string;
  edges: ReportEdge[];
  direction: "incoming" | "outgoing";
  nodeById: Map<string, ReportNode>;
}) {
  return (
    <section className="node-inspector-section">
      <h4>{title}</h4>
      {edges.length === 0 ? (
        <p className="node-inspector-empty">{empty}</p>
      ) : (
        <ol className="node-dependency-list">
          {edges.map((edge) => {
            const relatedId = direction === "outgoing" ? edge.target : edge.source;
            const relatedNode = nodeById.get(relatedId);
            return (
              <li className={`node-dependency ${edge.status}`} key={edge.id}>
                <div className="node-dependency-main">
                  <span>{edge.status}</span>
                  <strong>{relatedNode?.name ?? relatedId}</strong>
                </div>
                <small>{relatedNode?.path ?? relatedId}</small>
                <div className="node-dependency-imports">
                  <span>{edge.imports.length} imports</span>
                  {edge.imports.slice(0, 3).map((item) => (
                    <code key={`${edge.id}:${item.file}:${item.line}:${item.importPath}`}>
                      {item.file}:{item.line} {item.importPath}
                    </code>
                  ))}
                  {edge.imports.length > 3 && <small>+{edge.imports.length - 3} more imports</small>}
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </section>
  );
}

function severityWeight(severity: Violation["severity"]) {
  switch (severity) {
    case "error":
      return 0;
    case "warning":
      return 1;
    default:
      return 2;
  }
}

function DependencyEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  markerEnd,
  style,
  data,
}: EdgeProps<{
  edge: ReportEdge;
  violations?: Violation[];
  route?: "level-step" | "long-level" | "same-level" | "same-track";
  routeIndex?: number;
  routeCount?: number;
  railY?: number;
  busX?: number;
  onHover?: (edge: ReportEdge | null) => void;
  highlighted?: boolean;
  dimmed?: boolean;
}>) {
  const [edgePath, labelX, labelY] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    offset: data?.route === "same-track" ? 78 : 36,
    borderRadius: 12,
  });
  const edge = data?.edge;
  const imports = edge?.imports ?? [];
  const violations = data?.violations ?? [];
  const isActive = Boolean(data?.highlighted);
  const activate = () => {
    if (edge) {
      data?.onHover?.(edge);
    }
  };
  const deactivate = () => data?.onHover?.(null);
  const [visiblePath, computedLabelX, computedLabelY] =
    data?.route
      ? getRoutedPath({
          route: data.route,
          routeIndex: data.routeIndex ?? 0,
          routeCount: data.routeCount ?? 1,
          railY: data.railY,
          busX: data.busX,
          sourceX,
          sourceY,
          targetX,
          targetY,
          sourcePosition,
          targetPosition,
        })
      : [edgePath, labelX, labelY];

  return (
    <g
      data-edge-id={id}
      className={`dependency-edge ${data?.dimmed ? "is-dimmed" : ""} ${isActive ? "is-highlighted" : ""}`}
      onPointerEnter={activate}
      onPointerLeave={deactivate}
      onFocus={activate}
      onBlur={deactivate}
    >
      <BaseEdge id={id} path={visiblePath} markerEnd={markerEnd} style={style} />
      <path className="edge-hover-path" d={visiblePath} />
      {isActive && edge && (
        <EdgeLabelRenderer>
          <div
            className="edge-tooltip"
            style={{
              transform: `translate(-50%, -100%) translate(${computedLabelX}px, ${computedLabelY - 12}px)`,
            }}
          >
            <strong>
              {edge.source} {"->"} {edge.target}
            </strong>
            {imports.slice(0, 5).map((item) => (
              <div className="edge-tooltip-row" key={`${item.file}:${item.line}:${item.importPath}`}>
                <span>
                  {item.file}:{item.line}
                </span>
                <code>{item.importPath}</code>
              </div>
            ))}
            {imports.length > 5 && <small>+{imports.length - 5} more imports</small>}
            {violations.length > 0 && (
              <div className="edge-tooltip-violations">
                <span className="edge-tooltip-section-title">Ошибки и предупреждения</span>
                {violations.slice(0, 4).map((violation, index) => (
                  <div
                    className={`edge-tooltip-violation ${violation.severity}`}
                    key={`${violation.rule}:${violation.file ?? ""}:${violation.line ?? ""}:${index}`}
                  >
                    <div>
                      <span>{violation.severity}</span>
                      <code>{violation.rule}</code>
                    </div>
                    <p>{violation.message}</p>
                    {(violation.file || violation.importPath) && (
                      <small>
                        {violation.file ? `${violation.file}${violation.line ? `:${violation.line}` : ""}` : ""}
                        {violation.file && violation.importPath ? " · " : ""}
                        {violation.importPath ?? ""}
                      </small>
                    )}
                  </div>
                ))}
                {violations.length > 4 && <small>+{violations.length - 4} more issues</small>}
              </div>
            )}
          </div>
        </EdgeLabelRenderer>
      )}
    </g>
  );
}

function getRoutedPath({
  route,
  routeIndex,
  routeCount,
  railY,
  busX,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
}: {
  route: "level-step" | "long-level" | "same-level" | "same-track";
  routeIndex: number;
  routeCount: number;
  railY?: number;
  busX?: number;
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
}) {
  if (route === "level-step") {
    const topY = Math.min(sourceY, targetY);
    const bottomY = Math.max(sourceY, targetY);
    const railStart = topY + 26;
    const railEnd = bottomY - 26;
    const railRange = Math.max(1, railEnd - railStart);
    const railY = railStart + (railRange * (routeIndex + 1)) / (routeCount + 1);
    const path = `M ${sourceX},${sourceY} L ${sourceX},${railY} L ${targetX},${railY} L ${targetX},${targetY}`;
    return [path, (sourceX + targetX) / 2, railY] as const;
  }

  if (route === "long-level") {
    const rank = routeIndex - (routeCount - 1) / 2;
    const railX = (busX ?? Math.max(sourceX, targetX) + 140) + rank * 28;
    const path = `M ${sourceX},${sourceY} L ${railX},${sourceY} L ${railX},${targetY} L ${targetX},${targetY}`;
    return [path, railX, (sourceY + targetY) / 2] as const;
  }

  if (route === "same-track") {
    const side = sourcePosition === Position.Left || targetPosition === Position.Left ? -1 : 1;
    const railX =
      side < 0
        ? Math.min(sourceX, targetX) - 74 - routeIndex * 34
        : Math.max(sourceX, targetX) + 74 + routeIndex * 34;
    const path = `M ${sourceX},${sourceY} L ${railX},${sourceY} L ${railX},${targetY} L ${targetX},${targetY}`;
    return [path, railX, (sourceY + targetY) / 2] as const;
  }

  const side = sourceX <= targetX ? 1 : -1;
  const safeRailY = (railY ?? Math.max(sourceY, targetY) + 58) + routeIndex * 30;
  const sourceBendX = sourceX + side * 58;
  const targetBendX = targetX - side * 58;
  const path = [
    `M ${sourceX},${sourceY}`,
    `L ${sourceBendX},${sourceY}`,
    `L ${sourceBendX},${safeRailY}`,
    `L ${targetBendX},${safeRailY}`,
    `L ${targetBendX},${targetY}`,
    `L ${targetX},${targetY}`,
  ].join(" ");
  return [path, (sourceBendX + targetBendX) / 2, safeRailY] as const;
}

function ViolationsList({ violations }: { violations: Violation[] }) {
  const pageSize = 6;
  const [page, setPage] = useState(1);
  const pageCount = Math.max(1, Math.ceil(violations.length / pageSize));
  const safePage = Math.min(page, pageCount);
  const visibleViolations = violations.slice((safePage - 1) * pageSize, safePage * pageSize);

  if (violations.length === 0) {
    return (
      <div className="empty-list">
        <CircleCheck />
        <span>Нарушений не найдено.</span>
      </div>
    );
  }

  return (
    <>
      <ol className="violations-list">
        {visibleViolations.map((violation, index) => (
          <li className={`violation-item ${violation.severity}`} key={`${violation.rule}-${safePage}-${index}`}>
            <div className="violation-main">
              <span className="severity">{violation.severity}</span>
              <code>{violation.rule}</code>
              <span className="location">
                {violation.file ? `${violation.file}${violation.line ? `:${violation.line}` : ""}` : violation.from}
              </span>
            </div>
            <p>{violation.message}</p>
            {violation.importPath && <code className="import-path">{violation.importPath}</code>}
            {violation.suggestion && <small>{violation.suggestion}</small>}
          </li>
        ))}
      </ol>
      {pageCount > 1 && (
        <nav className="pagination" aria-label="Violations pagination">
          <Button variant="outline" disabled={safePage === 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>
            Prev
          </Button>
          <span>
            Page {safePage} / {pageCount}
          </span>
          <Button
            variant="outline"
            disabled={safePage === pageCount}
            onClick={() => setPage((value) => Math.min(pageCount, value + 1))}
          >
            Next
          </Button>
        </nav>
      )}
    </>
  );
}

function StateShell({ title, description, action }: { title: string; description: string; action?: string }) {
  return (
    <main className="state-shell">
      <div className="state-card">
        <ShieldAlert />
        <h1>{title}</h1>
        <p>{description}</p>
        {action && <code>{action}</code>}
      </div>
    </main>
  );
}

function kindLabel(kind: string) {
  switch (kind) {
    case "commonEntity":
      return "common";
    case "submodule":
      return "submodule";
    default:
      return kind;
  }
}
