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
import { CdromModal } from "@/components/cdrom-modal" // <- new import

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

export default function CreateInstancePage() {
  // Instance details
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [cpuCores, setCpuCores] = useState(2)
  const [memory, setMemory] = useState(2048)
  const [datastoreId, setDatastoreId] = useState<string | undefined>()
  const { datastores, loading, error: datastoreError } = useDatastores()

  // Network Interfaces
const [niDialogOpen, setNiDialogOpen] = useState(false)
const [networkInterfaces, setNetworkInterfaces] = useState<any[]>([])

  function addNetworkInterface(ni: any) {
    setNetworkInterfaces(prev => [...prev, ni])
  }

  function removeInterface(id: string) {
    setNetworkInterfaces((prev) => prev.filter((i) => i.id !== id))
  }

  // Storage / Disks
  const [storageDisks, setStorageDisks] = useState<any[]>([])
  const [storageDiskDialogOpen, setStorageDiskDialogOpen] = useState(false)

  const addStorageDisk = (disk: any) => {
    setStorageDisks(prev => [...prev, disk])
  }

  function removeStorageDisk(id: string) {
    setStorageDisks((prev) => prev.filter((d) => d.id !== id))
  }

  // CD-ROM state
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
  // Launch handler
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
    console.log("Launch payload:", payload)

      const res = await fetch("/api/v1/instances", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (!res.ok) {
        const text = await res.text()
        throw new Error(text || "Failed to create instance")
      }

      alert("Instance created successfully!")
      // redirect to /instances
      window.location.href = "/instances"

    } catch (err: any) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }

  }

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        {/* Sidebar */}
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
                  Instance
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
          <div className="flex">
            {/* Center Form */}
            <div className="flex-1 max-w-4xl p-8 space-y-6">
              <div>
                <h1 className="text-2xl font-bold">Launch Instance</h1>
                <p className="text-muted-foreground text-sm">
                  Configure compute, networking, and storage
                </p>
              </div>

              {/* Instance Details */}
              <Card>
                <CardHeader>
                  <CardTitle>Instance Details</CardTitle>
                  <CardDescription>Basic information</CardDescription>
                </CardHeader>
                <CardContent className="grid gap-4 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label>Name</Label>
                    <Input
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      placeholder="web-server-01"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label>Description</Label>
                    <Input
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      placeholder="Optional"
                    />
                  </div>
                </CardContent>
              </Card>

              {/* Compute */}
              <Card>
                <CardHeader>
                  <CardTitle>Compute</CardTitle>
                  <CardDescription>CPU and memory</CardDescription>
                </CardHeader>
                <CardContent className="grid gap-4 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label>CPU Cores</Label>
                    <Input
                      type="number"
                      value={cpuCores}
                      onChange={(e) =>
                        setCpuCores(parseInt(e.target.value))
                      }
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label>Memory (MB)</Label>
                    <Input
                      type="number"
                      value={memory}
                      onChange={(e) =>
                        setMemory(parseInt(e.target.value))
                      }
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label>Datastore</Label>
                    <Select
                        value={datastoreId}
                        onValueChange={setDatastoreId}
                    >
                        <SelectTrigger className="w-full rounded border border-gray-300 px-3 py-2" id="instance-datastore">
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

              {/* Networking */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                  <div>
                    <CardTitle>Networking</CardTitle>
                    <CardDescription>Configure interfaces</CardDescription>
                  </div>
                  <Button onClick={() => setNiDialogOpen(true)}>
                    Add Interface
                  </Button>
                </CardHeader>
                <CardContent>
                  {networkInterfaces.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                      No interfaces added
                    </p>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Bridge</TableHead>
                          <TableHead>Model</TableHead>
                          <TableHead>MAC</TableHead>
                          <TableHead>Connected</TableHead>
                          <TableHead></TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {networkInterfaces.map((ni) => (
                          <TableRow key={ni.id}>
                            <TableCell>{ni.bridgeName}</TableCell>
                            <TableCell>{ni.model}</TableCell>
                            <TableCell>{ni.mac || "-"}</TableCell>
                            <TableCell>{ni.connected ? "Yes" : "No"}</TableCell>
                            <TableCell>
                              <Button
                                size="sm"
                                variant="destructive"
                                onClick={() => removeInterface(ni.id)}
                              >
                                Remove
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>

              {/* Storage / Disks */}
              <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                  <div>
                    <CardTitle>Storage</CardTitle>
                    <CardDescription>Configure instance disks</CardDescription>
                  </div>
                  <Button onClick={() => setStorageDiskDialogOpen(true)}>
                    Add Disk
                  </Button>
                </CardHeader>
                <CardContent>
                  {storageDisks.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                      No disks added
                    </p>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Disk Type</TableHead>
                          <TableHead>Size (GB)</TableHead>
                          <TableHead>Datastore</TableHead>
                          <TableHead></TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {storageDisks.map((disk) => (
                          <TableRow key={disk.id}>
                            <TableCell>{disk.busType}</TableCell>
                            <TableCell>{disk.sizeGB}</TableCell>
                            <TableCell>{disk.datastoreId}</TableCell>
                            <TableCell>
                              <Button
                                size="sm"
                                variant="destructive"
                                onClick={() => removeStorageDisk(disk.id)}
                              >
                                Remove
                              </Button>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>

      {/* CD-ROMs Card */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>CD-ROMs</CardTitle>
            <CardDescription>Attach ISO or CD-ROM to your instance</CardDescription>
          </div>
          <Button size="sm" onClick={() => setCdromDialogOpen(true)}>
            Add CD-ROM
          </Button>
        </CardHeader>
        <CardContent>
          {cdroms.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No CD-ROMs added
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Path</TableHead>
                  <TableHead>Boot Order</TableHead>
                  <TableHead>Connected</TableHead>
                  <TableHead>Bootable</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {cdroms.map((cdrom) => (
                  <TableRow key={cdrom.id}>
                    <TableCell>{cdrom.path}</TableCell>
                    <TableCell>{cdrom.bootOrder}</TableCell>
                    <TableCell>{cdrom.connected ? "Yes" : "No"}</TableCell>
                    <TableCell>{cdrom.bootable ? "Yes" : "No"}</TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => removeCdrom(cdrom.id)}
                      >
                        Remove
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

            {/* Right Summary Panel */}
            <div className="w-80 border-l bg-muted/20 p-6">
              <h2 className="font-semibold mb-4">Instance Summary</h2>
              <div className="space-y-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Name</p>
                  <p>{name || "-"}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">CPU</p>
                  <p>{cpuCores} cores</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Memory</p>
                  <p>{memory} MB</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Interfaces</p>
                  <p>{networkInterfaces.length}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Disks</p>
                  <p>{storageDisks.length}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">CD-ROMs</p>
                  <p>{cdroms.length}</p>
                </div>
                <div className="pt-4 border-t">
                    <Button className="w-full" size="lg" onClick={handleLaunch} disabled={submitting}>
                        {submitting ? "Launching..." : "Launch Instance"}
                    </Button>

                    {error && <div className="text-red-600 mt-2">{error}</div>}
                </div>
              </div>
            </div>
          </div>
        </SidebarInset>

{/* Network Interface Dialog */}

<NetworkInterfaceModal
  open={niDialogOpen}
  onOpenChange={setNiDialogOpen}
  addInterface={addNetworkInterface}
/>

      <CdromModal
        open={cdromDialogOpen}
        onOpenChange={setCdromDialogOpen}
        addCdrom={addCdrom}
      />

    <StorageDiskModal
      open={storageDiskDialogOpen}
      onOpenChange={setStorageDiskDialogOpen}
      addDisk={addStorageDisk}
    />

      </div>
    </SidebarProvider>
  )
}