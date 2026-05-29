import { MarkerType, type Edge, type Node } from "reactflow";
import type { EdgeStatus, FeodReport, ReportEdge, ReportNode, Violation } from "../types";

const levelOrder = ["app", "pages", "modules", "common", "global"];
const laneLeft = 0;
const laneTop = 20;
const laneGap = 104;
const nodeWidth = 236;
const nodeHeight = 88;
const nodeGapY = 54;
const labelWidth = 150;
const trackWidth = 430;
const minLaneWidth = 1720;

type Track = {
  id: string;
};

type RouteKind = "level-step" | "long-level" | "same-level" | "same-track";

type EdgeRoute = {
  sourceHandle: string;
  targetHandle: string;
  route: RouteKind;
  routeGroup: string;
};

type EdgeCandidate = {
  edge: ReportEdge;
  route: EdgeRoute;
};

type LaneBounds = {
  top: number;
  bottom: number;
};

export function layoutReportGraph(report: FeodReport) {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const nodeById = new Map(report.nodes.map((node) => [node.id, node]));
  const tracks = buildTracks(report);
  const trackAssignments = assignTracks(report, tracks, nodeById);
  const laneWidth = Math.max(minLaneWidth, labelWidth + tracks.length * trackWidth + 28);
  const laneBounds = new Map<string, LaneBounds>();
  let currentTop = laneTop;

  for (const level of levelOrder) {
    const levelNodes = report.nodes
      .filter((node) => node.level === level && (node.kind !== "level" || node.fileCount > 0))
      .sort((a, b) => sortByTrack(a, b, trackAssignments, nodeById));
    const slots = buildLaneSlots(levelNodes, tracks, trackAssignments, nodeById);
    const laneHeight = Math.max(150, 68 + maxSlotDepth(slots) * (nodeHeight + nodeGapY));
    laneBounds.set(level, { top: currentTop, bottom: currentTop + laneHeight });

    nodes.push({
      id: `lane:${level}`,
      type: "lane",
      selectable: false,
      draggable: false,
      zIndex: 0,
      position: { x: laneLeft, y: currentTop },
      data: { label: level, count: levelNodes.length, trackCount: tracks.length },
      style: { width: laneWidth, height: laneHeight },
    });

    levelNodes.forEach((sourceNode) => {
      const trackId = trackAssignments.get(sourceNode.id) ?? "shared";
      const trackIndex = Math.max(
        0,
        tracks.findIndex((track) => track.id === trackId),
      );
      const slotIndex = slots.get(trackId)?.findIndex((node) => node.id === sourceNode.id) ?? 0;
      const centeredX =
        sourceNode.kind === "level" && sourceNode.level === "app"
          ? labelWidth + ((tracks.length * trackWidth - nodeWidth) / 2)
          : labelWidth + trackIndex * trackWidth + (trackWidth - nodeWidth) / 2;

      nodes.push({
        id: sourceNode.id,
        type: "entity",
        zIndex: 10,
        position: {
          x: laneLeft + centeredX,
          y: currentTop + 44 + slotIndex * (nodeHeight + nodeGapY),
        },
        data: {
          node: sourceNode,
          incoming: report.edges.filter((edge) => edge.target === sourceNode.id).length,
          outgoing: report.edges.filter((edge) => edge.source === sourceNode.id).length,
          violations: report.violations.filter(
            (violation) => violation.from === sourceNode.id || violation.to === sourceNode.id,
          ),
        },
      });
    });

    currentTop += laneHeight + laneGap;
  }

  const visibleNodeIds = new Set(nodes.filter((node) => !String(node.id).startsWith("lane:")).map((node) => node.id));
  const candidates: EdgeCandidate[] = [];
  for (const edge of report.edges) {
    if (!visibleNodeIds.has(edge.source) || !visibleNodeIds.has(edge.target)) {
      continue;
    }
    candidates.push({ edge, route: chooseRoute(edge, nodeById, trackAssignments, tracks) });
  }

  const routedEdges = assignRouteSlots(candidates, nodeById, trackAssignments, tracks);
  const busX = laneLeft + laneWidth - 34;
  for (const routed of routedEdges) {
    const color = edgeColor(routed.edge.status, edgeTone(routed.edge, nodeById));
    edges.push({
      id: routed.edge.id,
      source: routed.edge.source,
      target: routed.edge.target,
      sourceHandle: routed.route.sourceHandle,
      targetHandle: routed.route.targetHandle,
      type: "dependency",
      animated: routed.edge.status === "error",
      data: {
        edge: routed.edge,
        violations: edgeViolations(report.violations, routed.edge),
        route: routed.route.route,
        routeIndex: routed.routeIndex,
        routeCount: routed.routeCount,
        railY: sameLevelRailY(routed.edge, routed.route, nodeById, laneBounds),
        busX: busX - routed.routeIndex * 38,
      },
      style: edgeStyle(routed.edge.status, color),
      zIndex: 8,
      markerEnd: { type: MarkerType.ArrowClosed, color },
    });
  }

  return { nodes, edges };
}

function edgeViolations(violations: Violation[], edge: ReportEdge) {
  return violations.filter((violation) => violation.from === edge.source && violation.to === edge.target);
}

function sameLevelRailY(
  edge: ReportEdge,
  route: EdgeRoute,
  nodeById: Map<string, ReportNode>,
  laneBounds: Map<string, LaneBounds>,
) {
  if (route.route !== "same-level") {
    return undefined;
  }
  const sourceLevel = nodeById.get(edge.source)?.level;
  if (!sourceLevel) {
    return undefined;
  }
  const bounds = laneBounds.get(sourceLevel);
  return bounds ? bounds.bottom - 46 : undefined;
}

function buildTracks(report: FeodReport): Track[] {
  const pages = report.nodes.filter((node) => node.kind === "page");
  const appPageOrder = new Map<string, number>();
  report.edges
    .filter((edge) => edge.source === "level:app" && edge.target.startsWith("page:"))
    .forEach((edge) => {
      const firstLine = Math.min(...edge.imports.map((item) => item.line));
      appPageOrder.set(edge.target, Number.isFinite(firstLine) ? firstLine : appPageOrder.size + 1);
    });

  const orderedPages = [...pages].sort((a, b) => {
    const appOrderDelta = (appPageOrder.get(a.id) ?? 10_000) - (appPageOrder.get(b.id) ?? 10_000);
    return appOrderDelta || a.name.localeCompare(b.name);
  });
  const tracks = orderedPages.map((page) => ({ id: page.name }));

  if (!tracks.some((track) => track.id === "shared")) {
    const sharedIndex = Math.ceil(tracks.length / 2);
    tracks.splice(sharedIndex, 0, { id: "shared" });
  }

  if (tracks.length === 1) {
    const moduleNames = report.nodes
      .filter((node) => node.kind === "module")
      .map((node) => node.name)
      .sort();
    for (const moduleName of moduleNames.slice(0, 3)) {
      tracks.unshift({ id: moduleName });
    }
  }

  return tracks;
}

function assignTracks(report: FeodReport, tracks: Track[], nodeById: Map<string, ReportNode>) {
  const assignments = new Map<string, string>();
  const trackIds = new Set(tracks.map((track) => track.id));
  const pageTrackById = new Map(
    report.nodes.filter((node) => node.kind === "page").map((node) => [node.id, trackIds.has(node.name) ? node.name : "shared"]),
  );

  for (const node of report.nodes) {
    if (node.kind === "page") {
      assignments.set(node.id, pageTrackById.get(node.id) ?? "shared");
    }
    if (node.kind === "level") {
      assignments.set(node.id, node.level === "global" ? "shared" : "shared");
    }
  }

  for (const node of report.nodes.filter((item) => item.kind === "module")) {
    if (trackIds.has(node.name)) {
      assignments.set(node.id, node.name);
      continue;
    }
    const pageSources = incomingPageTracks(report, node.id, pageTrackById, nodeById);
    assignments.set(node.id, pageSources.size === 1 ? [...pageSources][0] : "shared");
  }

  for (const node of report.nodes.filter((item) => item.kind === "submodule")) {
    const parentTrack = node.parentId ? assignments.get(node.parentId) : undefined;
    assignments.set(node.id, parentTrack ?? "shared");
  }

  for (const node of report.nodes.filter((item) => item.kind === "commonEntity")) {
    const importerTracks = incomingEntityTracks(report, node.id, assignments);
    assignments.set(node.id, importerTracks.size === 1 ? [...importerTracks][0] : "shared");
  }

  return assignments;
}

function incomingPageTracks(
  report: FeodReport,
  targetId: string,
  pageTrackById: Map<string, string>,
  nodeById: Map<string, ReportNode>,
) {
  const tracks = new Set<string>();
  for (const edge of report.edges) {
    if (edge.target !== targetId) {
      continue;
    }
    const source = nodeById.get(edge.source);
    if (source?.kind === "page") {
      tracks.add(pageTrackById.get(source.id) ?? "shared");
    }
  }
  return tracks;
}

function incomingEntityTracks(report: FeodReport, targetId: string, assignments: Map<string, string>) {
  const tracks = new Set<string>();
  for (const edge of report.edges) {
    if (edge.target === targetId) {
      tracks.add(assignments.get(edge.source) ?? "shared");
    }
  }
  tracks.delete("shared");
  return tracks;
}

function buildLaneSlots(
  nodes: ReportNode[],
  tracks: Track[],
  assignments: Map<string, string>,
  nodeById: Map<string, ReportNode>,
) {
  const slots = new Map(tracks.map((track) => [track.id, [] as ReportNode[]]));
  for (const node of nodes) {
    const trackId = assignments.get(node.id) ?? "shared";
    slots.get(trackId)?.push(node);
  }
  for (const [trackId, trackNodes] of slots) {
    slots.set(
      trackId,
      trackNodes.sort((a, b) => sortWithinTrack(a, b, nodeById)),
    );
  }
  return slots;
}

function maxSlotDepth(slots: Map<string, ReportNode[]>) {
  return Math.max(1, ...[...slots.values()].map((items) => items.length));
}

function sortByTrack(
  a: ReportNode,
  b: ReportNode,
  assignments: Map<string, string>,
  nodeById: Map<string, ReportNode>,
) {
  const trackDelta = (assignments.get(a.id) ?? "shared").localeCompare(assignments.get(b.id) ?? "shared");
  return trackDelta || sortWithinTrack(a, b, nodeById);
}

function sortWithinTrack(a: ReportNode, b: ReportNode, nodeById: Map<string, ReportNode>) {
  const parentDelta = parentSortKey(a, nodeById).localeCompare(parentSortKey(b, nodeById));
  const kindDelta = kindWeight(a) - kindWeight(b);
  return parentDelta || kindDelta || a.path.localeCompare(b.path);
}

function parentSortKey(node: ReportNode, nodeById: Map<string, ReportNode>) {
  if (node.kind === "submodule" && node.parentId) {
    return nodeById.get(node.parentId)?.path ?? node.path;
  }
  return node.path;
}

function kindWeight(node: ReportNode) {
  switch (node.kind) {
    case "level":
      return 0;
    case "page":
      return 1;
    case "module":
      return 2;
    case "submodule":
      return 3;
    case "commonEntity":
      return 4;
    default:
      return 5;
  }
}

function assignRouteSlots(
  candidates: EdgeCandidate[],
  nodeById: Map<string, ReportNode>,
  assignments: Map<string, string>,
  tracks: Track[],
) {
  const groups = new Map<string, EdgeCandidate[]>();
  for (const candidate of candidates) {
    const group = groups.get(candidate.route.routeGroup) ?? [];
    group.push(candidate);
    groups.set(candidate.route.routeGroup, group);
  }

  return candidates.map((candidate) => {
    const group = groups.get(candidate.route.routeGroup) ?? [candidate];
    const sortedGroup = [...group].sort((a, b) => compareEdgesForRouting(a.edge, b.edge, nodeById, assignments, tracks));
    return {
      ...candidate,
      routeIndex: sortedGroup.findIndex((item) => item.edge.id === candidate.edge.id),
      routeCount: sortedGroup.length,
    };
  });
}

function compareEdgesForRouting(
  a: ReportEdge,
  b: ReportEdge,
  nodeById: Map<string, ReportNode>,
  assignments: Map<string, string>,
  tracks: Track[],
) {
  const trackIndex = new Map(tracks.map((track, index) => [track.id, index]));
  const sourceTrackDelta =
    (trackIndex.get(assignments.get(a.source) ?? "shared") ?? 0) -
    (trackIndex.get(assignments.get(b.source) ?? "shared") ?? 0);
  const targetTrackDelta =
    (trackIndex.get(assignments.get(a.target) ?? "shared") ?? 0) -
    (trackIndex.get(assignments.get(b.target) ?? "shared") ?? 0);
  const sourceDelta = (nodeById.get(a.source)?.path ?? a.source).localeCompare(nodeById.get(b.source)?.path ?? b.source);
  return sourceTrackDelta || targetTrackDelta || sourceDelta || a.target.localeCompare(b.target);
}

function chooseRoute(
  edge: ReportEdge,
  nodeById: Map<string, ReportNode>,
  assignments: Map<string, string>,
  tracks: Track[],
): EdgeRoute {
  const source = nodeById.get(edge.source);
  const target = nodeById.get(edge.target);
  const sourceLevel = levelOrder.indexOf(source?.level ?? "");
  const targetLevel = levelOrder.indexOf(target?.level ?? "");
  const levelDistance = Math.abs(sourceLevel - targetLevel);
  const sameLevel = sourceLevel === targetLevel;
  const sameTrack = (assignments.get(edge.source) ?? "shared") === (assignments.get(edge.target) ?? "shared");
  const trackIndex = new Map(tracks.map((track, index) => [track.id, index]));
  const sourceTrack = trackIndex.get(assignments.get(edge.source) ?? "shared") ?? 0;
  const targetTrack = trackIndex.get(assignments.get(edge.target) ?? "shared") ?? sourceTrack;

  if (!sameLevel) {
    if (levelDistance > 1 || (source?.level === "app" && target?.level !== "pages")) {
      return {
        sourceHandle: "source-right",
        targetHandle: "target-right",
        route: "long-level",
        routeGroup: `long:${source?.level ?? "unknown"}:${target?.level ?? "unknown"}:${sourceTrack}:${targetTrack}`,
      };
    }
    return {
      sourceHandle: "source-bottom",
      targetHandle: "target-top",
      route: "level-step",
      routeGroup: `step:${source?.level ?? "unknown"}:${target?.level ?? "unknown"}:${sourceTrack}:${targetTrack}`,
    };
  }

  if (sameTrack) {
    const isLastTrack = sourceTrack >= tracks.length - 1;
    return isLastTrack
      ? {
          sourceHandle: "source-left",
          targetHandle: "target-left",
          route: "same-track",
          routeGroup: `track:${source?.level ?? "unknown"}:${sourceTrack}:left`,
        }
      : {
          sourceHandle: "source-right",
          targetHandle: "target-right",
          route: "same-track",
          routeGroup: `track:${source?.level ?? "unknown"}:${sourceTrack}:right`,
        };
  }

  if (sourceTrack <= targetTrack) {
    return {
      sourceHandle: "source-right",
      targetHandle: "target-left",
      route: "same-level",
      routeGroup: `level:${source?.level ?? "unknown"}:${Math.min(sourceTrack, targetTrack)}:${Math.max(sourceTrack, targetTrack)}`,
    };
  }
  return {
    sourceHandle: "source-left",
    targetHandle: "target-right",
    route: "same-level",
    routeGroup: `level:${source?.level ?? "unknown"}:${Math.min(sourceTrack, targetTrack)}:${Math.max(sourceTrack, targetTrack)}`,
  };
}

function edgeStyle(status: EdgeStatus, color: string) {
  return {
    stroke: color,
    strokeWidth: status === "allowed" ? 1.8 : 2.5,
    opacity: status === "allowed" ? 0.82 : 0.95,
  };
}

function edgeColor(status: EdgeStatus, tone = 0) {
  switch (status) {
    case "error":
      return errorEdgeColors[tone % errorEdgeColors.length];
    case "warning":
      return "#d97706";
    default:
      return allowedEdgeColors[tone % allowedEdgeColors.length];
  }
}

const allowedEdgeColors = ["#15803d", "#16a34a", "#22c55e", "#059669", "#65a30d"];
const errorEdgeColors = ["#dc2626", "#b91c1c", "#ef4444", "#991b1b", "#f87171"];

function edgeTone(edge: ReportEdge, nodeById: Map<string, ReportNode>) {
  const source = nodeById.get(edge.source);
  const target = nodeById.get(edge.target);
  if (source?.level === "app") {
    return 3;
  }
  if (target?.kind === "commonEntity") {
    return 2;
  }
  if (target?.kind === "submodule") {
    return 1;
  }
  if (source?.kind === "page") {
    return 0;
  }
  return stableHash(`${edge.source}->${edge.target}`) % allowedEdgeColors.length;
}

function stableHash(value: string) {
  let hash = 0;
  for (const character of value) {
    hash = (hash * 31 + character.charCodeAt(0)) >>> 0;
  }
  return hash;
}
