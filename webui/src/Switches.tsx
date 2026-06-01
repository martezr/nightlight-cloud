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
import { columns, type Switch } from "./SwitchesColumns"
import { DataTable } from "./SwitchesDataTable"
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

async function getData(): Promise<Switch[]> {
  const response = await fetch("/api/v1/switches")
  if (!response.ok) throw new Error("Failed to fetch switches")
  return response.json()
}

export default function Page() {
  const [data, setData] = useState<Switch[]>([])
  const [open, setOpen] = useState(false)
  const [switchType, setSwitchType] = useState("")

  useEffect(() => {
    getData().then(setData)
  }, [])

  async function handleCreateSwitch(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const payload: Record<string, any> = {}
    formData.forEach((value, key) => { payload[key] = value })

    const jsonObject = {
      name: payload.name || "",
      description: payload.description || "",
      bridgeName: payload.bridgeName || "",
      siteId: payload.siteId || "",
      type: switchType || "",
      tags: [],
    }

    try {
      const response = await fetch("/api/v1/switches", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(jsonObject),
      })
      if (!response.ok) throw new Error("Failed to create switch")
    } catch (error) {
      console.error(error)
    }
    setOpen(false)
    window.location.reload()
  }

  return (
    <SidebarProvider>
      <title>Switches</title>
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
                  Switches
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Switches</h2>
            <div className="flex items-center gap-2">
              <Dialog open={open} onOpenChange={setOpen}>
                <DialogTrigger asChild>
                  <Button type="button" className="ml-auto">
                    Create Switch
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <form
                    id="create-switch-form"
                    onSubmit={handleCreateSwitch}
                  >
                    <DialogHeader>
                      <DialogTitle>Create Switch</DialogTitle>
                      <DialogDescription>
                        Create a new network switch backed by an OVS bridge.
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
                        <Label htmlFor="bridgeName">Bridge Name</Label>
                        <Input id="bridgeName" name="bridgeName" defaultValue="" placeholder="Defaults to generated ID if blank" />
                      </div>
                      <div className="grid gap-3">
                        <Label htmlFor="siteId">Site ID</Label>
                        <Input id="siteId" name="siteId" defaultValue="" />
                      </div>
                      <div className="grid gap-3">
                        <Label>Type</Label>
                        <Select onValueChange={setSwitchType} value={switchType}>
                          <SelectTrigger>
                            <SelectValue placeholder="Select type" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="core">Core</SelectItem>
                            <SelectItem value="leaf">Leaf</SelectItem>
                            <SelectItem value="access">Access</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                  </form>
                  <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button form="create-switch-form" type="submit">
                      Create Switch
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
