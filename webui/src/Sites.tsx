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
import { columns, type Site } from "./SitesColumns"
import { DataTable } from "./SitesDataTable"
import { useEffect, useState } from "react"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./components/ui/dialog"
import { Button } from "./components/ui/button"
import { Input } from "./components/ui/input"
import { Label } from "./components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./components/ui/select"

async function getData(): Promise<Site[]> {
  const response = await fetch("/api/v1/sites")
  if (!response.ok) throw new Error("Failed to fetch sites")
  return response.json()
}

export default function Page() {
  const [data, setData] = useState<Site[]>([])
  const [open, setOpen] = useState(false)
  const [topology, setTopology] = useState("")

  useEffect(() => {
    getData().then(setData)
  }, [])

  async function handleCreateSite(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const payload: Record<string, any> = {}
    formData.forEach((value, key) => { payload[key] = value })

    const jsonObject = {
      name: payload.name || "",
      description: payload.description || "",
      location: payload.location || "",
      type: payload.type || "",
      topology: topology || "single-bridge",
      wanBandwidth: payload.wanBandwidth ? parseInt(payload.wanBandwidth, 10) : 0,
    }

    try {
      const response = await fetch("/api/v1/sites", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(jsonObject),
      })
      if (!response.ok) throw new Error("Failed to create site")
    } catch (error) {
      console.error(error)
    }
    setOpen(false)
    window.location.reload()
  }

  return (
    <SidebarProvider>
      <title>Sites</title>
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
                  Sites
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Sites</h2>
            <div className="flex items-center gap-2">
              <Dialog open={open} onOpenChange={setOpen}>
                <DialogTrigger asChild>
                  <Button type="button" className="ml-auto">
                    Create Site
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <form id="create-site-form" onSubmit={handleCreateSite}>
                    <DialogHeader>
                      <DialogTitle>Create Site</DialogTitle>
                      <DialogDescription>
                        Define a new physical or logical site. The selected topology
                        determines which switches are provisioned automatically.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4" style={{ paddingTop: "5px" }}>
                      <div className="grid gap-3">
                        <Label htmlFor="name">Name</Label>
                        <Input id="name" name="name" defaultValue="" />
                      </div>
                      <div className="grid gap-3">
                        <Label htmlFor="description">Description</Label>
                        <Input id="description" name="description" defaultValue="" />
                      </div>
                      <div className="grid gap-3">
                        <Label htmlFor="location">Location</Label>
                        <Input id="location" name="location" defaultValue="" />
                      </div>
                      <div className="grid gap-3">
                        <Label htmlFor="type">Type</Label>
                        <Input id="type" name="type" defaultValue="" placeholder="e.g. datacenter, branch, robo" />
                      </div>
                      <div className="grid gap-3">
                        <Label>Topology</Label>
                        <Select onValueChange={setTopology} value={topology} defaultValue="single-bridge">
                          <SelectTrigger>
                            <SelectValue placeholder="Select topology" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="single-bridge">Single Bridge</SelectItem>
                            <SelectItem value="spine-leaf">Spine-Leaf</SelectItem>
                            <SelectItem value="robo">ROBO</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="grid gap-3">
                        <Label htmlFor="wanBandwidth">WAN Bandwidth (Mbps)</Label>
                        <Input id="wanBandwidth" name="wanBandwidth" type="number" min="0" defaultValue="" />
                      </div>
                    </div>
                  </form>
                  <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button form="create-site-form" type="submit">
                      Create Site
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>
          <DataTable columns={columns} data={data} />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
