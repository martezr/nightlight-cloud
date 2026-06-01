"use client"

import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Handle,
  Position,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type NodeProps,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"
import React, { useCallback, useEffect, useState } from "react"
import { Building2, Cable, Globe2, Network, Server, X } from "lucide-react"

// ── API shapes ─────────────────────────────────────────────────────────────────

interface OVSInterface {
  name: string
  type: string
  macAddress?: string
}

interface OVSBridge {
  id: string
  name: string
  interfaces: OVSInterface[]
}

interface Site {
  id: string        // name of the core switch / bridge for this site
  name: string      // human-readable site label
  bridges: string[] // names of bridges connected to the core switch
}

interface InstanceNIC {
  id: string
  vpcId: string
  bridgeName: string
  connected: boolean
  macAddress?: string
}

interface Instance {
  id: string
  name: string
  status: string
  instanceType?: string
  devices?: {
    networkInterfaces?: InstanceNIC[]
  }
}

// ── Node data types ────────────────────────────────────────────────────────────

type PhysicalIfNode = Node<{ label: string; ifType: string }, "physicalIf">
type SiteNodeT     = Node<{ label: string }, "site">
type BridgeNodeT   = Node<{ label: string; portCount?: number }, "bridge">
type RouterNodeT   = Node<{ label: string; status: string }, "router">
type InstanceNodeT = Node<{ label: string; status: string }, "instance">
type AppNode = PhysicalIfNode | SiteNodeT | BridgeNodeT | RouterNodeT | InstanceNodeT

// ── Custom node components ─────────────────────────────────────────────────────

// Shared circle node. Label floats ABOVE the bounding box so handles land
// exactly on the circle edges with no handle dots visible on the lines.
// Label line heights: status≈13px, name≈16px, sublabel≈13px, mb-2 gap≈8px.
// Top handle is offset upward by enough to clear the label so the connector
// terminates just above the text rather than drawing through it.
const LABEL_H_WITH_STATUS = 58   // status + name + sublabel + gap
const LABEL_H_NO_STATUS   = 45   // name + sublabel + gap

function CircleNode({
  icon,
  label,
  sublabel,
  status,
  borderColor = "border-blue-300",
  iconColor = "text-blue-500",
  hasTopHandle = false,
  hasBottomHandle = false,
}: {
  icon: React.ReactNode
  label: string
  sublabel: string
  status?: string
  borderColor?: string
  iconColor?: string
  hasTopHandle?: boolean
  hasBottomHandle?: boolean
}) {
  const topOffset = -(status !== undefined ? LABEL_H_WITH_STATUS : LABEL_H_NO_STATUS)
  return (
    <div className="relative h-[72px] w-[72px]">
      {/* Label above — outside bounding box so it doesn't shift handle positions */}
      <div className="pointer-events-none absolute bottom-full left-1/2 mb-2 w-max max-w-[140px] -translate-x-1/2 text-center">
        {status !== undefined && (
          <div className="mb-0.5 flex items-center justify-center gap-1">
            <span className={`inline-block size-1.5 rounded-full ${status === "running" ? "bg-emerald-500" : "bg-gray-300"}`} />
            <span className="text-[9px] capitalize text-gray-400">{status}</span>
          </div>
        )}
        <p className="text-[11px] font-semibold leading-tight text-gray-700">{label}</p>
        <p className="text-[9px] leading-tight text-gray-400">{sublabel}</p>
      </div>
      {hasTopHandle && (
        <Handle
          type="target"
          position={Position.Top}
          style={{ top: topOffset }}
          className="!h-px !w-px !border-0 !bg-transparent"
        />
      )}
      <div className={`flex h-full w-full items-center justify-center rounded-full border-2 bg-blue-50 shadow-sm ${borderColor}`}>
        <span className={iconColor}>{icon}</span>
      </div>
      {hasBottomHandle && (
        <Handle type="source" position={Position.Bottom} className="!h-px !w-px !border-0 !bg-transparent" />
      )}
    </div>
  )
}

function PhysicalIfNodeComponent({ data }: NodeProps<PhysicalIfNode>) {
  return (
    <CircleNode
      icon={<Cable className="size-7" />}
      label={data.label}
      sublabel={data.ifType}
      hasBottomHandle
    />
  )
}

const BRIDGE_LABEL_H          = 37  // name + subtitle + mb-2 gap
const BRIDGE_LABEL_H_WITH_PORT = 50  // adds port-count line

function BridgeNodeComponent({ data }: NodeProps<BridgeNodeT>) {
  const labelH = data.portCount !== undefined ? BRIDGE_LABEL_H_WITH_PORT : BRIDGE_LABEL_H
  return (
    <div className="relative h-12 w-24">
      {/* Label above the card */}
      <div className="pointer-events-none absolute bottom-full left-1/2 mb-2 w-max max-w-[140px] -translate-x-1/2 text-center">
        <p className="text-[11px] font-semibold leading-tight text-gray-700">{data.label}</p>
        <p className="text-[9px] leading-tight text-gray-400">OVS Bridge</p>
        {data.portCount !== undefined && (
          <p className="text-[9px] text-blue-400">{data.portCount} ports</p>
        )}
      </div>
      <Handle type="target" position={Position.Top} style={{ top: -labelH }} className="!h-px !w-px !border-0 !bg-transparent" />
      <div className="flex h-full w-full items-center justify-center rounded-lg border border-slate-200 bg-white shadow-sm">
        <Network className="size-5 text-blue-400" />
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-px !w-px !border-0 !bg-transparent" />
    </div>
  )
}

function RouterNodeComponent({ data }: NodeProps<RouterNodeT>) {
  return (
    <CircleNode
      icon={<Globe2 className="size-7" />}
      label={data.label}
      sublabel="Router"
      status={data.status}
      borderColor="border-blue-400"
      iconColor="text-blue-600"
      hasTopHandle
      hasBottomHandle
    />
  )
}

function SiteNodeComponent({ data }: NodeProps<SiteNodeT>) {
  return (
    <div className="h-full w-full rounded-xl border border-dashed border-slate-300 bg-slate-50/50 p-3">
      <div className="flex items-center gap-1.5">
        <Building2 className="size-3 text-slate-400" />
        <span className="text-[9px] font-medium uppercase tracking-widest text-slate-400">
          {data.label}
        </span>
      </div>
    </div>
  )
}

function InstanceNodeComponent({ data }: NodeProps<InstanceNodeT>) {
  return (
    <CircleNode
      icon={<Server className="size-7" />}
      label={data.label}
      sublabel="Instance"
      status={data.status}
      hasTopHandle
    />
  )
}

const nodeTypes = {
  physicalIf: PhysicalIfNodeComponent,
  site: SiteNodeComponent,
  bridge: BridgeNodeComponent,
  router: RouterNodeComponent,
  instance: InstanceNodeComponent,
}

// ── Details panel ──────────────────────────────────────────────────────────────

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="py-2.5 border-b border-gray-50 last:border-0">
      <p className="text-[10px] uppercase tracking-wider text-gray-400 mb-0.5">{label}</p>
      <div className="text-xs text-gray-700">{value}</div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const running = status === "running"
  return (
    <span className={`inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded-full font-medium ${running ? "bg-emerald-50 text-emerald-600" : "bg-gray-100 text-gray-500"}`}>
      <span className={`size-1.5 rounded-full ${running ? "bg-emerald-500" : "bg-gray-400"}`} />
      {status}
    </span>
  )
}

function DetailsPanel({
  node,
  bridges,
  sites,
  instances,
  onClose,
}: {
  node: AppNode
  bridges: OVSBridge[]
  sites: Site[]
  instances: Instance[]
  onClose: () => void
}) {
  const type = node.type

  const badgeMeta: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
    physicalIf: { label: "Physical Interface", color: "bg-gray-100 text-gray-600",    icon: <Cable className="size-3" /> },
    bridge:     { label: "OVS Bridge",         color: "bg-slate-100 text-slate-600",  icon: <Network className="size-3" /> },
    router:     { label: "Router",             color: "bg-blue-100 text-blue-600",    icon: <Globe2 className="size-3" /> },
    instance:   { label: "Instance",           color: "bg-indigo-100 text-indigo-600",icon: <Server className="size-3" /> },
    site:       { label: "Site",               color: "bg-violet-100 text-violet-600",icon: <Building2 className="size-3" /> },
  }
  const badge = badgeMeta[type ?? ""] ?? { label: type ?? "", color: "bg-gray-100 text-gray-600", icon: null }

  let title = ""
  let rows: React.ReactNode = null

  if (type === "physicalIf") {
    const data = (node as PhysicalIfNode).data
    title = data.label
    const connectedBridges = bridges.filter((b) => b.interfaces.some((i) => i.name === data.label))
    rows = (
      <>
        <DetailRow label="Interface Name" value={data.label} />
        <DetailRow label="Interface Type" value={data.ifType} />
        <DetailRow
          label={`Connected Bridges (${connectedBridges.length})`}
          value={
            connectedBridges.length === 0
              ? <span className="text-gray-400">None</span>
              : <ul className="space-y-0.5">{connectedBridges.map((b) => <li key={b.name}>{b.name}</li>)}</ul>
          }
        />
      </>
    )
  } else if (type === "bridge") {
    const bridgeName = node.id.substring("bridge-".length)
    const bridge = bridges.find((b) => b.name === bridgeName)
    title = bridgeName
    rows = (
      <>
        <DetailRow label="Bridge Name" value={bridgeName} />
        <DetailRow label="Port Count" value={String(bridge?.interfaces.length ?? 0)} />
        <DetailRow
          label={`Interfaces (${bridge?.interfaces.length ?? 0})`}
          value={
            (bridge?.interfaces ?? []).length === 0
              ? <span className="text-gray-400">None</span>
              : <ul className="divide-y divide-gray-50">
                  {bridge!.interfaces.map((iface) => (
                    <li key={iface.name} className="flex items-center justify-between py-1">
                      <span className="font-mono text-[10px]">{iface.name}</span>
                      <span className="text-[10px] text-gray-400">{iface.type}</span>
                    </li>
                  ))}
                </ul>
          }
        />
      </>
    )
  } else if (type === "router") {
    const data = (node as RouterNodeT).data
    const instId = node.id.substring("inst-".length)
    const inst = instances.find((i) => i.id === instId)
    const nics = inst?.devices?.networkInterfaces ?? []
    title = data.label
    rows = (
      <>
        <DetailRow label="Name" value={data.label} />
        <DetailRow label="Status" value={<StatusBadge status={data.status} />} />
        <DetailRow label="Type" value="Router" />
        {inst?.id && <DetailRow label="ID" value={<span className="font-mono text-[10px]">{inst.id}</span>} />}
        <DetailRow
          label={`Network Interfaces (${nics.length})`}
          value={
            nics.length === 0
              ? <span className="text-gray-400">None</span>
              : <ul className="divide-y divide-gray-50">
                  {nics.map((nic, i) => (
                    <li key={nic.id} className="py-1.5">
                      <div className="flex items-center justify-between">
                        <span className="text-gray-700">{nic.bridgeName || "—"}</span>
                        <span className={`text-[9px] px-1.5 py-0.5 rounded-full font-medium ${i === 0 ? "bg-blue-100 text-blue-600" : "bg-green-100 text-green-600"}`}>
                          {i === 0 ? "upstream" : "downstream"}
                        </span>
                      </div>
                      {nic.macAddress && (
                        <p className="mt-0.5 font-mono text-[10px] text-gray-400">{nic.macAddress}</p>
                      )}
                    </li>
                  ))}
                </ul>
          }
        />
      </>
    )
  } else if (type === "instance") {
    const data = (node as InstanceNodeT).data
    const instId = node.id.substring("inst-".length)
    const inst = instances.find((i) => i.id === instId)
    const nics = inst?.devices?.networkInterfaces ?? []
    title = data.label
    rows = (
      <>
        <DetailRow label="Name" value={data.label} />
        <DetailRow label="Status" value={<StatusBadge status={data.status} />} />
        {inst?.id && <DetailRow label="ID" value={<span className="font-mono text-[10px]">{inst.id}</span>} />}
        <DetailRow
          label={`Network Interfaces (${nics.length})`}
          value={
            nics.length === 0
              ? <span className="text-gray-400">None</span>
              : <ul className="divide-y divide-gray-50">
                  {nics.map((nic) => (
                    <li key={nic.id} className="py-1.5">
                      <p className="text-gray-700">{nic.bridgeName || "—"}</p>
                      {nic.macAddress && (
                        <p className="mt-0.5 font-mono text-[10px] text-gray-400">{nic.macAddress}</p>
                      )}
                    </li>
                  ))}
                </ul>
          }
        />
      </>
    )
  } else if (type === "site") {
    const data = (node as SiteNodeT).data
    const siteId = node.id.substring("site-".length)
    const site = sites.find((s) => s.id === siteId)
    title = data.label
    rows = (
      <>
        <DetailRow label="Name" value={data.label} />
        <DetailRow label="ID" value={<span className="font-mono text-[10px]">{siteId}</span>} />
        <DetailRow label="Core Bridge" value={siteId} />
        <DetailRow
          label={`Downstream Bridges (${site?.bridges.length ?? 0})`}
          value={
            (site?.bridges ?? []).length === 0
              ? <span className="text-gray-400">None</span>
              : <ul className="space-y-0.5">
                  {site!.bridges.map((b) => <li key={b} className="text-gray-700">{b}</li>)}
                </ul>
          }
        />
      </>
    )
  }

  return (
    <div className="absolute right-0 top-0 z-10 flex h-full w-72 flex-col border-l border-gray-200 bg-white shadow-xl">
      <div className="flex items-start justify-between border-b border-gray-100 p-4">
        <div className="min-w-0 flex-1">
          <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium ${badge.color}`}>
            {badge.icon}
            {badge.label}
          </span>
          <h2 className="mt-1.5 truncate text-sm font-semibold text-gray-800">{title}</h2>
        </div>
        <button
          onClick={onClose}
          className="ml-3 shrink-0 rounded p-0.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
        >
          <X className="size-4" />
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {rows}
      </div>
    </div>
  )
}

// ── Topology builder ───────────────────────────────────────────────────────────

const COL_GAP = 200
const ROW_GAP = 180  // extra to clear labels floating above next row

// Bridge card dimensions (icon-only rectangular card).
const BRIDGE_W = 96
const BRIDGE_H = 48

const SITE_PAD       = 20   // left/right padding inside site container
const SITE_TITLE     = 40   // height of the site label bar
const SITE_LABEL_PAD = 44   // vertical clearance above each bridge card for its floating label
const SITE_IN_GAP    = 16   // gap between bridge cards inside a site
const SITE_OUT_GAP   = 60   // gap between adjacent sites / standalone nodes
const SITE_INTER_ROW = 60   // vertical gap between core bridge row and downstream row

// Two bridge rows inside the site: core on top, downstream below.
const SITE_H = SITE_TITLE + SITE_LABEL_PAD + BRIDGE_H + SITE_INTER_ROW + BRIDGE_H + SITE_PAD

const EXTERNAL_BRIDGE = "nl-external"
const EXTERNAL_IFACE  = "eth0"

function siteW(downstreamCount: number): number {
  const n = Math.max(1, downstreamCount)
  return SITE_PAD * 2 + n * BRIDGE_W + Math.max(0, n - 1) * SITE_IN_GAP
}

function centerRow(count: number, index: number, totalWidth: number): number {
  if (count === 0) return 0
  const rowWidth = (count - 1) * COL_GAP + BRIDGE_W
  return Math.max(0, (totalWidth - rowWidth) / 2) + index * COL_GAP
}

function buildTopology(
  apiBridges: OVSBridge[],
  sites: Site[],
  instances: Instance[],
): { nodes: AppNode[]; edges: Edge[] } {
  const nodes: AppNode[] = []
  const edges: Edge[] = []

  // ── 1. Bridge map: nl-external always seeded ────────────────────────────────
  const bridgeMap = new Map<string, OVSBridge>()

  bridgeMap.set(EXTERNAL_BRIDGE, {
    id: EXTERNAL_BRIDGE,
    name: EXTERNAL_BRIDGE,
    interfaces: [{ name: EXTERNAL_IFACE, type: "system" }],
  })

  apiBridges.forEach((b) => {
    if (b.name === EXTERNAL_BRIDGE) {
      const hasEth0 = b.interfaces.some((i) => i.name === EXTERNAL_IFACE)
      bridgeMap.set(b.name, {
        ...b,
        interfaces: hasEth0
          ? b.interfaces
          : [{ name: EXTERNAL_IFACE, type: "system" }, ...b.interfaces],
      })
    } else {
      bridgeMap.set(b.name, b)
    }
  })

  // Ensure every bridge referenced by a site exists in the map.
  sites.forEach((site) => {
    if (!bridgeMap.has(site.id))
      bridgeMap.set(site.id, { id: site.id, name: site.id, interfaces: [] })
    site.bridges.forEach((bName) => {
      if (!bridgeMap.has(bName))
        bridgeMap.set(bName, { id: bName, name: bName, interfaces: [] })
    })
  })

  // Infer additional bridges from instance NIC data.
  instances.forEach((inst) =>
    inst.devices?.networkInterfaces?.forEach((nic) => {
      if (nic.bridgeName && !bridgeMap.has(nic.bridgeName))
        bridgeMap.set(nic.bridgeName, { id: nic.bridgeName, name: nic.bridgeName, interfaces: [] })
    }),
  )

  // ── 2. Site membership ────────────────────────────────────────────────────
  const bridgeToSiteId = new Map<string, string>()
  sites.forEach((site) => {
    bridgeToSiteId.set(site.id, site.id)
    site.bridges.forEach((b) => bridgeToSiteId.set(b, site.id))
  })

  const standaloneBridgeNames = Array.from(bridgeMap.keys()).filter(
    (name) => !bridgeToSiteId.has(name),
  )

  // Standalone bridges connected to non-first router NICs are "downstream" —
  // they should render below the router row, not above it.
  const downstreamBridgeNamesSet = new Set<string>()
  instances
    .filter((i) => i.instanceType === "router")
    .forEach((router) => {
      ;(router.devices?.networkInterfaces ?? []).slice(1).forEach((nic) => {
        if (nic.bridgeName) downstreamBridgeNamesSet.add(nic.bridgeName)
      })
    })
  const standaloneUpstreamNames   = standaloneBridgeNames.filter((n) => !downstreamBridgeNamesSet.has(n))
  const standaloneDownstreamNames = standaloneBridgeNames.filter((n) =>  downstreamBridgeNamesSet.has(n))

  // ── 3. Physical interfaces: eth0 always first ──────────────────────────────
  const seenIfs = new Set<string>([EXTERNAL_IFACE])
  const physicalIfs: OVSInterface[] = [{ name: EXTERNAL_IFACE, type: "system" }]

  apiBridges.forEach((b) =>
    b.interfaces
      .filter((i) => (i.type === "system" || i.type === "physical") && !seenIfs.has(i.name))
      .forEach((i) => { physicalIfs.push(i); seenIfs.add(i.name) }),
  )

  // ── 4. Row 1 layout: sites + standalone bridges ────────────────────────────
  type Row1Item =
    | { kind: "site"; site: Site; x: number; width: number }
    | { kind: "standalone"; name: string; x: number }

  let row1X = 0
  const row1Items: Row1Item[] = []

  sites.forEach((site) => {
    const w = siteW(site.bridges.length)
    row1Items.push({ kind: "site", site, x: row1X, width: w })
    row1X += w + SITE_OUT_GAP
  })

  standaloneUpstreamNames.forEach((name) => {
    row1Items.push({ kind: "standalone", name, x: row1X })
    row1X += BRIDGE_W + SITE_OUT_GAP
  })

  const row1TotalW = row1X > 0 ? row1X - SITE_OUT_GAP : BRIDGE_W

  // ── 5. Canvas width ────────────────────────────────────────────────────────
  const routers  = instances.filter((i) => i.instanceType === "router")
  const vms      = instances.filter((i) => i.instanceType !== "router")

  const totalWidth = Math.max(
    row1TotalW,
    (physicalIfs.length - 1) * COL_GAP + BRIDGE_W,
    routers.length > 0 ? (routers.length - 1) * COL_GAP + BRIDGE_W : 0,
    standaloneDownstreamNames.length > 0 ? (standaloneDownstreamNames.length - 1) * COL_GAP + BRIDGE_W : 0,
    vms.length     > 0 ? (vms.length     - 1) * COL_GAP + BRIDGE_W : 0,
    BRIDGE_W,
  )

  const row1OffsetX = Math.max(0, (totalWidth - row1TotalW) / 2)

  let currentY = 0

  // ── 6. Row 0: physical interfaces ──────────────────────────────────────────
  physicalIfs.forEach((iface, idx) => {
    nodes.push({
      id: `iface-${iface.name}`,
      type: "physicalIf",
      position: { x: centerRow(physicalIfs.length, idx, totalWidth), y: currentY },
      data: { label: iface.name, ifType: iface.type },
    })
  })
  currentY += ROW_GAP

  // ── 7. Row 1: site containers + standalone bridges ──────────────────────────
  const bridgeRowY = currentY

  row1Items.forEach((item) => {
    const absX = row1OffsetX + item.x

    if (item.kind === "site") {
      const { site, width } = item

      nodes.push({
        id: `site-${site.id}`,
        type: "site",
        position: { x: absX, y: bridgeRowY },
        style: { width, height: SITE_H },
        data: { label: site.name },
      })

      // Core bridge: centred at top row inside site.
      const coreX = (width - BRIDGE_W) / 2
      const coreBridge = bridgeMap.get(site.id)
      nodes.push({
        id: `bridge-${site.id}`,
        type: "bridge",
        parentId: `site-${site.id}`,
        extent: "parent" as const,
        position: { x: coreX, y: SITE_TITLE + SITE_LABEL_PAD },
        data: { label: site.id, portCount: coreBridge?.interfaces.length || undefined },
      })

      // Downstream bridges in a row below the core.
      const downstreamY = SITE_TITLE + SITE_LABEL_PAD + BRIDGE_H + SITE_INTER_ROW
      site.bridges.forEach((bridgeName, idx) => {
        const bridge = bridgeMap.get(bridgeName)
        nodes.push({
          id: `bridge-${bridgeName}`,
          type: "bridge",
          parentId: `site-${site.id}`,
          extent: "parent" as const,
          position: { x: SITE_PAD + idx * (BRIDGE_W + SITE_IN_GAP), y: downstreamY },
          data: { label: bridgeName, portCount: bridge?.interfaces.length || undefined },
        })
      })

      // Core → downstream edges.
      site.bridges.forEach((bridgeName) => {
        edges.push({
          id: `e-bridge-${site.id}-bridge-${bridgeName}`,
          source: `bridge-${site.id}`,
          target: `bridge-${bridgeName}`,
          style: { stroke: "#cbd5e1", strokeWidth: 1.5 },
        })
      })
    } else {
      const bridge = bridgeMap.get(item.name)
      nodes.push({
        id: `bridge-${item.name}`,
        type: "bridge",
        position: { x: absX, y: bridgeRowY + Math.max(0, (SITE_H - BRIDGE_H) / 2) },
        data: { label: item.name, portCount: bridge?.interfaces.length || undefined },
      })
    }
  })

  // Physical interface → bridge edges.
  bridgeMap.forEach((bridge, bridgeName) => {
    bridge.interfaces
      .filter((i) => seenIfs.has(i.name))
      .forEach((iface) => {
        edges.push({
          id: `e-iface-${iface.name}-bridge-${bridgeName}`,
          source: `iface-${iface.name}`,
          target: `bridge-${bridgeName}`,
          style: { stroke: "#cbd5e1", strokeWidth: 1.5 },
        })
      })
  })

  currentY += SITE_H + ROW_GAP

  // ── 8. Row 2: routers ──────────────────────────────────────────────────────
  if (routers.length > 0) {
    routers.forEach((inst, idx) => {
      const running = inst.status === "running"
      nodes.push({
        id: `inst-${inst.id}`,
        type: "router",
        position: { x: centerRow(routers.length, idx, totalWidth), y: currentY },
        data: { label: inst.name, status: inst.status },
      })

      const matchingNics = inst.devices?.networkInterfaces?.filter(
        (n) => n.bridgeName && bridgeMap.has(n.bridgeName),
      ) ?? []

      matchingNics.forEach((nic, nicIdx) => {
        const isUpstream = nicIdx === 0
        if (isUpstream) {
          // Upstream: bridge → router
          edges.push({
            id: `e-up-${nic.bridgeName}-${inst.id}`,
            source: `bridge-${nic.bridgeName}`,
            target: `inst-${inst.id}`,
            animated: running,
            label: "upstream",
            labelStyle: { fontSize: 9, fill: "#94a3b8" },
            labelBgStyle: { fill: "#f8fafc" },
            style: { stroke: "#93c5fd", strokeWidth: 1.5 },
          })
        } else {
          // Downstream: router → bridge
          edges.push({
            id: `e-dn-${inst.id}-${nic.bridgeName}`,
            source: `inst-${inst.id}`,
            target: `bridge-${nic.bridgeName}`,
            animated: running,
            label: "downstream",
            labelStyle: { fontSize: 9, fill: "#94a3b8" },
            labelBgStyle: { fill: "#f8fafc" },
            style: { stroke: "#86efac", strokeWidth: 1.5, strokeDasharray: "5 4" },
          })
        }
      })
    })
    currentY += ROW_GAP
  }

  // ── 9. Row 3: downstream standalone bridges (below routers) ────────────────
  if (standaloneDownstreamNames.length > 0) {
    standaloneDownstreamNames.forEach((name, idx) => {
      const bridge = bridgeMap.get(name)
      nodes.push({
        id: `bridge-${name}`,
        type: "bridge",
        position: { x: centerRow(standaloneDownstreamNames.length, idx, totalWidth), y: currentY },
        data: { label: name, portCount: bridge?.interfaces.length || undefined },
      })
    })
    currentY += BRIDGE_H + ROW_GAP
  }

  // ── 10. Row 4: regular instances ────────────────────────────────────────────
  vms.forEach((inst, idx) => {
    const running = inst.status === "running"
    nodes.push({
      id: `inst-${inst.id}`,
      type: "instance",
      position: { x: centerRow(vms.length, idx, totalWidth), y: currentY },
      data: { label: inst.name, status: inst.status },
    })
    inst.devices?.networkInterfaces
      ?.filter((n) => n.bridgeName && bridgeMap.has(n.bridgeName))
      .forEach((nic) => {
        edges.push({
          id: `e-bridge-${nic.bridgeName}-inst-${inst.id}`,
          source: `bridge-${nic.bridgeName}`,
          target: `inst-${inst.id}`,
          animated: running,
          style: { stroke: "#cbd5e1", strokeWidth: 1.5 },
        })
      })
  })

  return { nodes, edges }
}

// ── Page ───────────────────────────────────────────────────────────────────────

export default function NetworkTopology() {
  const [nodes, setNodes, onNodesChange] = useNodesState<AppNode>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [loading, setLoading] = useState(true)
  const [rawBridges, setRawBridges] = useState<OVSBridge[]>([])
  const [rawSites, setRawSites] = useState<Site[]>([])
  const [rawInstances, setRawInstances] = useState<Instance[]>([])
  const [selectedNode, setSelectedNode] = useState<AppNode | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const [bridgesRes, sitesRes, instancesRes] = await Promise.allSettled([
          fetch("/api/v1/bridges").then((r) => (r.ok ? r.json() : [])),
          fetch("/api/v1/sites").then((r) => (r.ok ? r.json() : [])),
          fetch("/api/v1/instances").then((r) => (r.ok ? r.json() : [])),
        ])

        const apiBridges: OVSBridge[] =
          bridgesRes.status === "fulfilled" && Array.isArray(bridgesRes.value)
            ? bridgesRes.value : []
        const sites: Site[] =
          sitesRes.status === "fulfilled" && Array.isArray(sitesRes.value)
            ? sitesRes.value : []
        const instances: Instance[] =
          instancesRes.status === "fulfilled" && Array.isArray(instancesRes.value)
            ? instancesRes.value : []

        setRawBridges(apiBridges)
        setRawSites(sites)
        setRawInstances(instances)

        const { nodes: n, edges: e } = buildTopology(apiBridges, sites, instances)
        setNodes(n)
        setEdges(e)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [setNodes, setEdges])

  const handleNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node as AppNode)
  }, [])

  return (
    <SidebarProvider>
      <title>Network Topology</title>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#" className="font-bold">
                  Network Topology
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="relative flex-1" style={{ height: "calc(100vh - 4rem)" }}>
          {loading && (
            <div className="flex h-full items-center justify-center text-muted-foreground">
              Loading topology…
            </div>
          )}
          {!loading && (
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              nodeTypes={nodeTypes}
              onNodeClick={handleNodeClick}
              fitView
              fitViewOptions={{ padding: 0.25 }}
              proOptions={{ hideAttribution: true }}
            >
              <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
              <Controls />
              <MiniMap
                nodeColor={(n) => {
                  if (n.type === "bridge") return "#e2e8f0"
                  if (n.type === "router" || n.type === "instance" || n.type === "physicalIf") return "#dbeafe"
                  return "#f1f5f9"
                }}
                zoomable
                pannable
              />
            </ReactFlow>
          )}
          {selectedNode && (
            <DetailsPanel
              node={selectedNode}
              bridges={rawBridges}
              sites={rawSites}
              instances={rawInstances}
              onClose={() => setSelectedNode(null)}
            />
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
