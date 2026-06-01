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
import { columns, type Subnet } from "./SubnetsColumns"
import { SubnetsDataTable } from "./SubnetsDataTable"
import { useEffect, useState } from "react"
import { Button } from "./components/ui/button"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "./components/ui/dialog"
import { Input } from "./components/ui/input"
import { Label } from "./components/ui/label"
import { Switch } from "./components/ui/switch"

async function getData(): Promise<Subnet[]> {
  const response = await fetch("/api/v1/subnets")
  if (!response.ok) throw new Error("Failed to fetch subnets")
  return response.json()
}

export default function Page() {
  const [data, setData] = useState<Subnet[]>([])
  const [open, setOpen] = useState(false)
  const [dhcpEnabled, setDhcpEnabled] = useState(false)

  useEffect(() => {
    getData().then(setData)
  }, [])

  async function handleCreate(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const fd = new FormData(event.currentTarget)
    const body = {
      name: fd.get("name") as string,
      description: fd.get("description") as string,
      cidrBlock: fd.get("cidrBlock") as string,
      gateway: fd.get("gateway") as string,
      bridgeName: fd.get("bridgeName") as string,
      dhcpServer: dhcpEnabled,
      ipPoolRange: dhcpEnabled ? (fd.get("ipPoolRange") as string) : "",
      dnsServers: fd.get("dnsServers")
        ? (fd.get("dnsServers") as string).split(",").map((s) => s.trim()).filter(Boolean)
        : [],
      domainName: fd.get("domainName") as string,
      ntpServers: fd.get("ntpServers")
        ? (fd.get("ntpServers") as string).split(",").map((s) => s.trim()).filter(Boolean)
        : [],
    }
    try {
      const res = await fetch("/api/v1/subnets", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error("Failed to create subnet")
    } catch (err) {
      console.error(err)
    }
    setOpen(false)
    setDhcpEnabled(false)
    window.location.reload()
  }

  return (
    <SidebarProvider>
      <title>Subnets</title>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#" className="font-bold">Subnets</BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Subnets</h2>
            <Dialog open={open} onOpenChange={(o) => { setOpen(o); if (!o) setDhcpEnabled(false) }}>
              <DialogTrigger asChild>
                <Button type="button" className="ml-auto">Create Subnet</Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <form id="create-subnet-form" onSubmit={handleCreate}>
                  <DialogHeader>
                    <DialogTitle>Create Subnet</DialogTitle>
                    <DialogDescription>
                      Define a new subnet within a VNet.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="grid gap-4 py-4">
                    <div className="grid gap-2">
                      <Label htmlFor="name">Name</Label>
                      <Input id="name" name="name" required />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="description">Description</Label>
                      <Input id="description" name="description" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="cidrBlock">CIDR Block</Label>
                      <Input id="cidrBlock" name="cidrBlock" placeholder="10.0.0.0/24" required />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="gateway">Gateway</Label>
                      <Input id="gateway" name="gateway" placeholder="10.0.0.1" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="bridgeName">Bridge Name</Label>
                      <Input id="bridgeName" name="bridgeName" placeholder="br-default" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="dnsServers">DNS Servers (comma-separated)</Label>
                      <Input id="dnsServers" name="dnsServers" placeholder="8.8.8.8, 8.8.4.4" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="domainName">Domain Name</Label>
                      <Input id="domainName" name="domainName" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="ntpServers">NTP Servers (comma-separated)</Label>
                      <Input id="ntpServers" name="ntpServers" />
                    </div>
                    <div className="flex items-center gap-3">
                      <Switch
                        id="dhcpServer"
                        checked={dhcpEnabled}
                        onCheckedChange={setDhcpEnabled}
                      />
                      <Label htmlFor="dhcpServer">Enable DHCP Server</Label>
                    </div>
                    {dhcpEnabled && (
                      <div className="grid gap-2">
                        <Label htmlFor="ipPoolRange">IP Pool Range</Label>
                        <Input id="ipPoolRange" name="ipPoolRange" placeholder="10.0.0.10-10.0.0.254" />
                      </div>
                    )}
                  </div>
                </form>
                <DialogFooter>
                  <DialogClose asChild>
                    <Button variant="outline">Cancel</Button>
                  </DialogClose>
                  <Button form="create-subnet-form" type="submit">Create Subnet</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>
          <SubnetsDataTable columns={columns} data={data} />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
