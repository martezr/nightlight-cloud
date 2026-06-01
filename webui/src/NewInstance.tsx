"use client"

import { useEffect, useState } from "react"

import { AppSidebar } from "@/components/app-sidebar"
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar"

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { CdromModal } from "@/components/cdrom-modal"

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

import { StorageDiskModal } from "@/components/storage-disk-modal"
import { NetworkInterfaceModal } from "./components/network-interface-modal"
import type { Datastore } from "./DatastoresColumns"
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList } from "./components/ui/breadcrumb"
import { Separator } from "./components/ui/separator"
import { Server, Cpu, Network, HardDrive, Disc, Trash2, Plus, MemoryStick } from "lucide-react"

function useDatastores() {
    const [datastores, setDatastores] = useState<Datastore[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        setLoading(true)
        fetch("/api/v1/datastores")
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch datastores")
                return res.json()
            })
            .then((data) => setDatastores(data))
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [])

    return { datastores, loading, error }
}

function SectionHeader({ step, icon, title, description }: {
  step: number
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="flex items-center gap-3">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-semibold">
        {step}
      </div>
      <div className="flex items-center gap-2">
        {icon}
        <div>
          <CardTitle className="text-base">{title}</CardTitle>
          <CardDescription className="text-xs">{description}</CardDescription>
        </div>
      </div>
    </div>
  )
}

function EmptyState({ icon, message }: { icon: React.ReactNode; message: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 rounded-md border border-dashed py-8 text-muted-foreground">
      <div className="opacity-40">{icon}</div>
      <p className="text-sm">{message}</p>
    </div>
  )
}

export default function CreateInstancePage() {
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [cpuCores, setCpuCores] = useState(2)
  const [memory, setMemory] = useState(2048)
  const [datastoreId, setDatastoreId] = useState<string | undefined>()
  const { datastores, loading, error: datastoreError } = useDatastores()

  const [niDialogOpen, setNiDialogOpen] = useState(false)
  const [networkInterfaces, setNetworkInterfaces] = useState<any[]>([])

  function addNetworkInterface(ni: any) {
    setNetworkInterfaces(prev => [...prev, ni])
  }
  function removeInterface(id: string) {
    setNetworkInterfaces((prev) => prev.filter((i) => i.id !== id))
  }

  const [storageDisks, setStorageDisks] = useState<any[]>([])
  const [storageDiskDialogOpen, setStorageDiskDialogOpen] = useState(false)

  const addStorageDisk = (disk: any) => {
    setStorageDisks(prev => [...prev, disk])
  }
  function removeStorageDisk(id: string) {
    setStorageDisks((prev) => prev.filter((d) => d.id !== id))
  }

  const [cdroms, setCdroms] = useState<any[]>([])
  const [cdromDialogOpen, setCdromDialogOpen] = useState(false)

  const addCdrom = (cdrom: any) => {
    setCdroms(prev => [...prev, cdrom])
  }
  function removeCdrom(id: string) {
    setCdroms((prev) => prev.filter((c) => c.id !== id))
  }

  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleLaunch() {
    setSubmitting(true)
    setError(null)
    try {
      const payload = {
        name,
        description,
        cpuCores,
        cpuSockets: 1,
        bootType: "bios",
        memoryMB: memory,
        devices: {
          networkInterfaces: [...networkInterfaces],
          storageDisks: [...storageDisks],
          cdroms: [...cdroms],
        },
        datastoreId,
      }

      const res = await fetch("/api/v1/instances", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || "Failed to create instance")
      }

      window.location.href = "/instances"
    } catch (err: any) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  const totalStorageGB = storageDisks.reduce((sum, d) => sum + (d.sizeGB || 0), 0)

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <AppSidebar collapsible="icon" variant="sidebar" />

        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="/instances" className="font-bold">Instances</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbItem className="hidden md:block">
                  <span className="text-muted-foreground">/</span>
                </BreadcrumbItem>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="#" className="font-bold">Launch Instance</BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </header>

          <div className="flex min-h-0 flex-1">
            {/* Main form */}
            <div className="flex-1 overflow-y-auto p-8">
              <div className="max-w-3xl space-y-6">
                <div>
                  <h1 className="text-2xl font-bold">Launch Instance</h1>
                  <p className="text-sm text-muted-foreground">Configure compute, networking, and storage</p>
                </div>

                {/* Step 1 — Instance Details */}
                <Card>
                  <CardHeader>
                    <SectionHeader
                      step={1}
                      icon={<Server className="h-4 w-4 text-muted-foreground" />}
                      title="Instance Details"
                      description="Name and description"
                    />
                  </CardHeader>
                  <CardContent className="grid gap-4 md:grid-cols-2">
                    <div className="grid gap-2">
                      <Label>Name</Label>
                      <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="web-server-01" />
                    </div>
                    <div className="grid gap-2">
                      <Label>Description</Label>
                      <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Optional" />
                    </div>
                  </CardContent>
                </Card>

                {/* Step 2 — Compute */}
                <Card>
                  <CardHeader>
                    <SectionHeader
                      step={2}
                      icon={<Cpu className="h-4 w-4 text-muted-foreground" />}
                      title="Compute"
                      description="CPU, memory, and datastore"
                    />
                  </CardHeader>
                  <CardContent className="grid gap-4 md:grid-cols-2">
                    <div className="grid gap-2">
                      <Label>CPU Cores</Label>
                      <Input type="number" value={cpuCores} onChange={(e) => setCpuCores(parseInt(e.target.value))} />
                    </div>
                    <div className="grid gap-2">
                      <Label>Memory (MB)</Label>
                      <Input type="number" value={memory} onChange={(e) => setMemory(parseInt(e.target.value))} />
                    </div>
                    <div className="grid gap-2 md:col-span-2">
                      <Label>Datastore</Label>
                      <Select value={datastoreId} onValueChange={setDatastoreId}>
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Select a datastore" />
                        </SelectTrigger>
                        <SelectContent>
                          {!loading && !datastoreError && datastores.map((ds) => (
                            <SelectItem key={ds.id} value={ds.id}>
                              {ds.name} ({ds.id})
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </CardContent>
                </Card>

                {/* Step 3 — Networking */}
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between">
                    <SectionHeader
                      step={3}
                      icon={<Network className="h-4 w-4 text-muted-foreground" />}
                      title="Networking"
                      description="Configure network interfaces"
                    />
                    <Button size="sm" onClick={() => setNiDialogOpen(true)}>
                      <Plus className="mr-1 h-4 w-4" />
                      Add Interface
                    </Button>
                  </CardHeader>
                  <CardContent>
                    {networkInterfaces.length === 0 ? (
                      <EmptyState icon={<Network className="h-8 w-8" />} message="No interfaces added" />
                    ) : (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Bridge</TableHead>
                            <TableHead>Model</TableHead>
                            <TableHead>MAC</TableHead>
                            <TableHead>Connected</TableHead>
                            <TableHead className="w-10" />
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {networkInterfaces.map((ni) => (
                            <TableRow key={ni.id}>
                              <TableCell className="font-mono text-sm">{ni.bridgeName}</TableCell>
                              <TableCell>{ni.model}</TableCell>
                              <TableCell className="font-mono text-sm">{ni.mac || "—"}</TableCell>
                              <TableCell>
                                <Badge variant={ni.connected ? "default" : "secondary"}>
                                  {ni.connected ? "Connected" : "Disconnected"}
                                </Badge>
                              </TableCell>
                              <TableCell>
                                <Button size="icon" variant="ghost" className="h-8 w-8 text-muted-foreground hover:text-red-500" onClick={() => removeInterface(ni.id)}>
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    )}
                  </CardContent>
                </Card>

                {/* Step 4 — Storage */}
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between">
                    <SectionHeader
                      step={4}
                      icon={<HardDrive className="h-4 w-4 text-muted-foreground" />}
                      title="Storage"
                      description="Configure instance disks"
                    />
                    <Button size="sm" onClick={() => setStorageDiskDialogOpen(true)}>
                      <Plus className="mr-1 h-4 w-4" />
                      Add Disk
                    </Button>
                  </CardHeader>
                  <CardContent>
                    {storageDisks.length === 0 ? (
                      <EmptyState icon={<HardDrive className="h-8 w-8" />} message="No disks added" />
                    ) : (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Bus Type</TableHead>
                            <TableHead>Size (GB)</TableHead>
                            <TableHead>Datastore</TableHead>
                            <TableHead className="w-10" />
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {storageDisks.map((disk) => (
                            <TableRow key={disk.id}>
                              <TableCell><Badge variant="outline">{disk.busType}</Badge></TableCell>
                              <TableCell>{disk.sizeGB} GB</TableCell>
                              <TableCell className="text-sm text-muted-foreground">{disk.datastoreId}</TableCell>
                              <TableCell>
                                <Button size="icon" variant="ghost" className="h-8 w-8 text-muted-foreground hover:text-red-500" onClick={() => removeStorageDisk(disk.id)}>
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    )}
                  </CardContent>
                </Card>

                {/* Step 5 — CD-ROMs */}
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between">
                    <SectionHeader
                      step={5}
                      icon={<Disc className="h-4 w-4 text-muted-foreground" />}
                      title="CD-ROMs"
                      description="Attach ISO images"
                    />
                    <Button size="sm" onClick={() => setCdromDialogOpen(true)}>
                      <Plus className="mr-1 h-4 w-4" />
                      Add CD-ROM
                    </Button>
                  </CardHeader>
                  <CardContent>
                    {cdroms.length === 0 ? (
                      <EmptyState icon={<Disc className="h-8 w-8" />} message="No CD-ROMs added" />
                    ) : (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Path</TableHead>
                            <TableHead>Boot Order</TableHead>
                            <TableHead>Connected</TableHead>
                            <TableHead>Bootable</TableHead>
                            <TableHead className="w-10" />
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {cdroms.map((cdrom) => (
                            <TableRow key={cdrom.id}>
                              <TableCell className="font-mono text-sm">{cdrom.path}</TableCell>
                              <TableCell>{cdrom.bootOrder}</TableCell>
                              <TableCell>
                                <Badge variant={cdrom.connected ? "default" : "secondary"}>
                                  {cdrom.connected ? "Connected" : "Disconnected"}
                                </Badge>
                              </TableCell>
                              <TableCell>
                                <Badge variant={cdrom.bootable ? "default" : "secondary"}>
                                  {cdrom.bootable ? "Bootable" : "No"}
                                </Badge>
                              </TableCell>
                              <TableCell>
                                <Button size="icon" variant="ghost" className="h-8 w-8 text-muted-foreground hover:text-red-500" onClick={() => removeCdrom(cdrom.id)}>
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    )}
                  </CardContent>
                </Card>
              </div>
            </div>

            {/* Sticky summary panel */}
            <div className="hidden w-72 shrink-0 border-l xl:block">
              <div className="sticky top-0 flex h-screen flex-col p-6">
                <h2 className="mb-1 font-semibold">Instance Summary</h2>
                <p className="mb-6 text-xs text-muted-foreground">Review before launching</p>

                <div className="flex-1 space-y-4 text-sm">
                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <Server className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div className="min-w-0">
                      <p className="text-xs text-muted-foreground">Name</p>
                      <p className="truncate font-medium">{name || <span className="text-muted-foreground">—</span>}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <Cpu className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div>
                      <p className="text-xs text-muted-foreground">Compute</p>
                      <p className="font-medium">{cpuCores} cores</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <MemoryStick className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div>
                      <p className="text-xs text-muted-foreground">Memory</p>
                      <p className="font-medium">{memory >= 1024 ? `${(memory / 1024).toFixed(memory % 1024 === 0 ? 0 : 1)} GB` : `${memory} MB`}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <Network className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div>
                      <p className="text-xs text-muted-foreground">Interfaces</p>
                      <p className="font-medium">{networkInterfaces.length} {networkInterfaces.length === 1 ? "interface" : "interfaces"}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <HardDrive className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div>
                      <p className="text-xs text-muted-foreground">Storage</p>
                      <p className="font-medium">
                        {storageDisks.length === 0
                          ? "No disks"
                          : `${storageDisks.length} disk${storageDisks.length > 1 ? "s" : ""} · ${totalStorageGB} GB`}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3 rounded-lg border p-3">
                    <Disc className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div>
                      <p className="text-xs text-muted-foreground">CD-ROMs</p>
                      <p className="font-medium">{cdroms.length} {cdroms.length === 1 ? "drive" : "drives"}</p>
                    </div>
                  </div>
                </div>

                <div className="mt-6 space-y-3 border-t pt-6">
                  <Button className="w-full" size="lg" onClick={handleLaunch} disabled={submitting || !name}>
                    {submitting ? "Launching..." : "Launch Instance"}
                  </Button>
                  {!name && (
                    <p className="text-center text-xs text-muted-foreground">Instance name is required</p>
                  )}
                  {error && (
                    <p className="rounded-md bg-red-50 px-3 py-2 text-xs text-red-600">{error}</p>
                  )}
                </div>
              </div>
            </div>
          </div>
        </SidebarInset>

        <NetworkInterfaceModal open={niDialogOpen} onOpenChange={setNiDialogOpen} addInterface={addNetworkInterface} />
        <CdromModal open={cdromDialogOpen} onOpenChange={setCdromDialogOpen} addCdrom={addCdrom} />
        <StorageDiskModal open={storageDiskDialogOpen} onOpenChange={setStorageDiskDialogOpen} addDisk={addStorageDisk} />
      </div>
    </SidebarProvider>
  )
}
