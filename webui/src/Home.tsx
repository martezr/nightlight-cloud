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
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "./components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Server,
  Network,
  Database,
  Plus,
  ArrowRight,
  Cpu,
  MemoryStick,
  HardDrive,
  CirclePlay,
} from "lucide-react"
import { Link } from "react-router-dom"
import * as React from "react"
import { useState } from "react"
import { type Instance } from "./InstancesColumns"

function useResourceCounts() {
  const [instances, setInstances] = useState<Instance[]>([])
  const [vnetCount, setVNetCount] = useState<number | null>(null)
  const [datastoreCount, setDatastoreCount] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)

  React.useEffect(() => {
    Promise.all([
      fetch("/api/v1/instances").then((r) => (r.ok ? r.json() : [])),
      fetch("/api/v1/vnets").then((r) => (r.ok ? r.json() : [])),
      fetch("/api/v1/datastores").then((r) => (r.ok ? r.json() : [])),
    ])
      .then(([inst, vnets, ds]) => {
        setInstances(Array.isArray(inst) ? inst : [])
        setVNetCount(Array.isArray(vnets) ? vnets.length : 0)
        setDatastoreCount(Array.isArray(ds) ? ds.length : 0)
      })
      .catch(() => {
        setInstances([])
        setVNetCount(0)
        setDatastoreCount(0)
      })
      .finally(() => setLoading(false))
  }, [])

  return { instances, vnetCount, datastoreCount, loading }
}

function StatCard({
  icon,
  label,
  value,
  href,
  accent,
}: {
  icon: React.ReactNode
  label: string
  value: number | null
  href: string
  accent?: string
}) {
  return (
    <Link to={href}>
      <Card className="transition-shadow hover:shadow-md cursor-pointer">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardDescription className="text-sm font-medium">{label}</CardDescription>
          <div className={`rounded-md p-2 ${accent ?? "bg-muted"}`}>{icon}</div>
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-bold tabular-nums">
            {value === null ? <span className="text-muted-foreground text-xl">—</span> : value}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}

function InstanceStatusBadge({ status }: { status: string }) {
  const s = status?.toLowerCase()
  if (s === "running")
    return <Badge className="bg-green-500 text-white hover:bg-green-500">Running</Badge>
  if (s === "stopped")
    return <Badge variant="secondary">Stopped</Badge>
  if (s === "error" || s === "failed")
    return <Badge variant="destructive">{status}</Badge>
  return <Badge variant="outline">{status || "Unknown"}</Badge>
}

export default function Page() {
  const { instances, vnetCount, datastoreCount, loading } = useResourceCounts()

  const runningCount = instances.filter(
    (i) => i.status?.toLowerCase() === "running"
  ).length
  const stoppedCount = instances.filter(
    (i) => i.status?.toLowerCase() === "stopped"
  ).length
  const otherCount = instances.length - runningCount - stoppedCount

  const statusGroups = [
    { label: "Running", count: runningCount, color: "bg-green-500" },
    { label: "Stopped", count: stoppedCount, color: "bg-muted-foreground/30" },
    ...(otherCount > 0 ? [{ label: "Other", count: otherCount, color: "bg-yellow-400" }] : []),
  ]

  const recentInstances = instances.slice(0, 6)

  return (
    <SidebarProvider>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#" className="font-bold">Dashboard</BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

        </header>

        <div className="flex flex-1 flex-col gap-6 p-8">

          {/* Stat cards */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={<Server className="h-4 w-4 text-primary" />}
              label="Total Instances"
              value={loading ? null : instances.length}
              href="/instances"
              accent="bg-primary/10"
            />
            <StatCard
              icon={<CirclePlay className="h-4 w-4 text-green-600" />}
              label="Running"
              value={loading ? null : runningCount}
              href="/instances"
              accent="bg-green-500/10"
            />
            <StatCard
              icon={<Network className="h-4 w-4 text-blue-600" />}
              label="Networks"
              value={loading ? null : vnetCount}
              href="/vnets"
              accent="bg-blue-500/10"
            />
            <StatCard
              icon={<Database className="h-4 w-4 text-orange-600" />}
              label="Datastores"
              value={loading ? null : datastoreCount}
              href="/datastores"
              accent="bg-orange-500/10"
            />
          </div>

          <div className="grid gap-4 lg:grid-cols-3">
            {/* Instance Status Breakdown */}
            <Card className="lg:col-span-2">
              <CardHeader>
                <CardTitle className="text-base">Instance Status</CardTitle>
                <CardDescription>Current state across all instances</CardDescription>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <p className="text-sm text-muted-foreground">Loading...</p>
                ) : instances.length === 0 ? (
                  <div className="flex flex-col items-center justify-center gap-2 rounded-md border border-dashed py-10 text-muted-foreground">
                    <Server className="h-8 w-8 opacity-30" />
                    <p className="text-sm">No instances found</p>
                    <Link to="/instances/createinstance">
                      <Button size="sm" className="mt-2">
                        <Plus className="mr-1 h-4 w-4" />
                        Launch Instance
                      </Button>
                    </Link>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {/* Visual bar */}
                    <div className="flex h-4 w-full overflow-hidden rounded-full bg-muted">
                      {statusGroups.map((g) => (
                        <div
                          key={g.label}
                          className={`${g.color} transition-all`}
                          style={{ width: `${(g.count / instances.length) * 100}%` }}
                        />
                      ))}
                    </div>
                    {/* Legend */}
                    <div className="flex flex-wrap gap-4">
                      {statusGroups.map((g) => (
                        <div key={g.label} className="flex items-center gap-2 text-sm">
                          <span className={`h-3 w-3 rounded-full ${g.color}`} />
                          <span className="text-muted-foreground">{g.label}</span>
                          <span className="font-semibold">{g.count}</span>
                          <span className="text-muted-foreground text-xs">
                            ({Math.round((g.count / instances.length) * 100)}%)
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Quick Actions</CardTitle>
                <CardDescription>Common tasks</CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-2">
                <Link to="/instances/createinstance">
                  <Button variant="outline" className="w-full justify-between">
                    <span className="flex items-center gap-2">
                      <Server className="h-4 w-4" />
                      Launch Instance
                    </span>
                    <Plus className="h-4 w-4" />
                  </Button>
                </Link>
                <Link to="/vnets">
                  <Button variant="outline" className="w-full justify-between">
                    <span className="flex items-center gap-2">
                      <Network className="h-4 w-4" />
                      Create Network
                    </span>
                    <Plus className="h-4 w-4" />
                  </Button>
                </Link>
                <Link to="/datastores">
                  <Button variant="outline" className="w-full justify-between">
                    <span className="flex items-center gap-2">
                      <HardDrive className="h-4 w-4" />
                      Add Datastore
                    </span>
                    <Plus className="h-4 w-4" />
                  </Button>
                </Link>
                <Separator className="my-1" />
                <Link to="/instances">
                  <Button variant="ghost" className="w-full justify-between text-muted-foreground">
                    View all instances
                    <ArrowRight className="h-4 w-4" />
                  </Button>
                </Link>
              </CardContent>
            </Card>
          </div>

          {/* Recent Instances */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-base">Recent Instances</CardTitle>
                <CardDescription>Last {recentInstances.length} instances</CardDescription>
              </div>
              <Link to="/instances">
                <Button variant="ghost" size="sm" className="text-muted-foreground">
                  View all
                  <ArrowRight className="ml-1 h-4 w-4" />
                </Button>
              </Link>
            </CardHeader>
            <CardContent className="p-0">
              {loading ? (
                <p className="px-6 py-4 text-sm text-muted-foreground">Loading...</p>
              ) : recentInstances.length === 0 ? (
                <p className="px-6 py-4 text-sm text-muted-foreground">No instances yet.</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>
                        <span className="flex items-center gap-1">
                          <Cpu className="h-3 w-3" /> CPU
                        </span>
                      </TableHead>
                      <TableHead>
                        <span className="flex items-center gap-1">
                          <MemoryStick className="h-3 w-3" /> Memory
                        </span>
                      </TableHead>
                      <TableHead>IP Address</TableHead>
                      <TableHead className="w-10" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentInstances.map((instance) => (
                      <TableRow key={instance.id}>
                        <TableCell className="font-medium">
                          <Link to={`/instances/${instance.id}`} className="hover:underline">
                            {instance.name}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <InstanceStatusBadge status={instance.status} />
                        </TableCell>
                        <TableCell>{instance.cpuCores} cores</TableCell>
                        <TableCell>
                          {parseInt(instance.memoryMB) >= 1024
                            ? `${(parseInt(instance.memoryMB) / 1024).toFixed(parseInt(instance.memoryMB) % 1024 === 0 ? 0 : 1)} GB`
                            : `${instance.memoryMB} MB`}
                        </TableCell>
                        <TableCell className="font-mono text-sm text-muted-foreground">
                          {instance.primaryIPAddress || "—"}
                        </TableCell>
                        <TableCell>
                          <Link to={`/instances/${instance.id}`}>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <ArrowRight className="h-4 w-4" />
                            </Button>
                          </Link>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
