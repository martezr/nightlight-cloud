import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { useParams } from "react-router-dom"
import { useEffect, useState, useCallback } from "react"
import { type Network } from "./NetworksColumns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./components/ui/tabs"
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  applyNodeChanges,
  applyEdgeChanges,
  addEdge,
} from "@xyflow/react"
import "@xyflow/react/dist/style.css"
import dagre from "dagre"

import { columns, type NetworkFlow } from "./NetworkFlowLogsColumns"
import { NetworkFlowsDataTable } from "./NetworkFlowLogsDataTable"
// =========================
// NETWORK DETAILS
// =========================
function useNetworkDetails(id?: string) {
  const [data, setData] = useState<Network | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return

    fetch(`/api/v1/vnets/${id}`)
      .then((res) => res.json())
      .then(setData)
      .finally(() => setLoading(false))
  }, [id])

  return { data, loading }
}

function useNetworkFlowDetails(id?: string) {
    const [data, setData] = useState<NetworkFlow[] | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!id) return
        setLoading(true)
        setError(null)
        fetch(`/api/v1/vnets/${id}/flowlogs`)
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch network flow details")
                return res.json()
            })
            .then(setData)
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [id])

    return { data, loading, error }
}

// =========================
// GRAPH HELPERS
// =========================
const NODE_WIDTH = 190
const NODE_HEIGHT = 80

function getEdgeColor(protocol: string) {
  switch (protocol) {
    case "TCP":
      return "#4dabf7"
    case "UDP":
      return "#51cf66"
    case "ICMP":
      return "#ffa94d"
    case "ARP":
      return "#adb5bd"
    default:
      return "#ced4da"
  }
}

function layoutGraph(nodes: any[], edges: any[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))

  g.setGraph({
    rankdir: "LR",
    nodesep: 60,
    ranksep: 120,
  })

  nodes.forEach((node) =>
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  )

  edges.forEach((edge) => g.setEdge(edge.source, edge.target))

  dagre.layout(g)

  return nodes.map((node) => {
    const pos = g.node(node.id)
    return {
      ...node,
      position: {
        x: pos.x - NODE_WIDTH / 2,
        y: pos.y - NODE_HEIGHT / 2,
      },
      sourcePosition: "right",
      targetPosition: "left",
    }
  })
}

function isolateNode(nodeId: string, nodes: any[], edges: any[]) {
  const connectedEdges = edges.filter(
    (e) => e.source === nodeId || e.target === nodeId
  )

  const connectedNodeIds = new Set<string>()
  connectedNodeIds.add(nodeId)

  connectedEdges.forEach((e) => {
    connectedNodeIds.add(e.source)
    connectedNodeIds.add(e.target)
  })

  return {
    nodes: nodes.filter((n) => connectedNodeIds.has(n.id)),
    edges: connectedEdges,
  }
}

function transformGraph(data: any) {
  let rfNodes: any[] = []
  let rfEdges: any[] = []

  data.nodes.forEach((n: any) => {
    rfNodes.push({
      id: n.id,
      data: {
        label: (
          <div style={{ lineHeight: 1.2 }}>
            <div style={{ fontWeight: 600 }}>
              {n.host || n.ip}
            </div>
            <div style={{ fontSize: 11, opacity: 0.7 }}>
              {n.ip}
            </div>
            <div style={{ fontSize: 10, opacity: 0.5 }}>
              {n.mac}
            </div>
          </div>
        ),
      },
      position: { x: 0, y: 0 },
      style: {
        background: "#1e293b",
        color: "#e2e8f0",
        border: "1px solid #334155",
        borderRadius: 10,
        padding: 10,
      },
    })
  })

  data.edges.forEach((e: any, i: number) => {
    rfEdges.push({
      id: `${e.src}-${e.dst}-${i}`,
      source: e.src,
      target: e.dst,
      label: `${e.protocol}`,
      animated: e.bytes > 50000,
      style: {
        stroke: getEdgeColor(e.protocol),
        strokeWidth: Math.max(1, Math.log(e.bytes + 1)),
      },
    })
  })

  rfNodes = layoutGraph(rfNodes, rfEdges)

  return { nodes: rfNodes, edges: rfEdges }
}

// =========================
// MAIN PAGE
// =========================
export default function Page() {
  const { id } = useParams()
  const { data, loading } = useNetworkDetails(id)
  const { data: flowData, loading: flowLoading, error: flowError } = useNetworkFlowDetails(id)
  console.log("Network details:", data, loading)
  console.log("Network flow details:", flowData, flowLoading, flowError)
  const [allNodes, setAllNodes] = useState<any[]>([])
  const [allEdges, setAllEdges] = useState<any[]>([])

  const [nodes, setNodes] = useState<any[]>([])
  const [edges, setEdges] = useState<any[]>([])

  const [focusedNode, setFocusedNode] = useState<string | null>(null)

  // =========================
  // GRAPH FETCH
  // =========================
  useEffect(() => {
    if (!id) return

    let mounted = true

    const loadGraph = () => {
      fetch(`/api/v1/vnets/${id}/graph`)
        .then((res) => res.json())
        .then((data) => {
          if (!mounted) return

          const { nodes, edges } = transformGraph(data)

          setAllNodes(nodes)
          setAllEdges(edges)

          setNodes(nodes)
          setEdges(edges)
        })
        .catch(console.error)
    }

    loadGraph()

    const interval = setInterval(loadGraph, 30 * 60 * 1000)

    return () => {
      mounted = false
      clearInterval(interval)
    }
  }, [id])

  // =========================
  // INTERACTIONS
  // =========================
  const onNodesChange = useCallback(
    (changes: any) =>
      setNodes((nds) => applyNodeChanges(changes, nds)),
    []
  )

  const onEdgesChange = useCallback(
    (changes: any) =>
      setEdges((eds) => applyEdgeChanges(changes, eds)),
    []
  )

  const onConnect = useCallback(
    (params: any) =>
      setEdges((eds) => addEdge(params, eds)),
    []
  )

  const onNodeClick = useCallback(
    (_: any, node: any) => {
      if (focusedNode === node.id) {
        setNodes(allNodes)
        setEdges(allEdges)
        setFocusedNode(null)
        return
      }

      const { nodes: n, edges: e } = isolateNode(
        node.id,
        allNodes,
        allEdges
      )

      setNodes(n)
      setEdges(e)
      setFocusedNode(node.id)
    },
    [focusedNode, allNodes, allEdges]
  )

  // =========================
  // UI
  // =========================
  return (
    <SidebarProvider>
      <AppSidebar />

      <SidebarInset>
        <header className="flex h-16 items-center gap-2 px-4">
          <SidebarTrigger />
          <Separator orientation="vertical" className="h-4" />

          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href="/networks">networks</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbLink href={`/networks/${id}`}>
                  {data ? data.name : "Loading..."}
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col p-4 gap-4">
          <h2 className="text-2xl font-bold">
            {data ? data.name : "Loading..."}
          </h2>

          <Tabs defaultValue="topology">
            <TabsList>
              <TabsTrigger value="topology">Topology</TabsTrigger>
              <TabsTrigger value="flows">Flows</TabsTrigger>
            </TabsList>

            <TabsContent value="topology">
              <div
                style={{
                  width: "100%",
                  height: "75vh",
                  background: "#020617",
                  borderRadius: 12,
                }}
              >
                <ReactFlow
                  nodes={nodes}
                  edges={edges}
                  onNodesChange={onNodesChange}
                  onEdgesChange={onEdgesChange}
                  onConnect={onConnect}
                  onNodeClick={onNodeClick}
                  fitView
                >
                  <Background gap={20} size={1} color="#1e293b" />
                  <MiniMap />
                  <Controls />
                </ReactFlow>
              </div>
            </TabsContent>

            <TabsContent value="flows">
              Flow logs coming soon
                          <NetworkFlowsDataTable columns={columns} data={flowData ?? []} />
              
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}