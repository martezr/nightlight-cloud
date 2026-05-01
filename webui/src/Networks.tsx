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
import { columns, type Network } from "./NetworksColumns"
import { DataTable } from "./NetworksDataTable"
import { useEffect, useState } from "react"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "./components/ui/dialog"
import { Button } from "./components/ui/button"
import { Input } from "./components/ui/input"
import { Label } from "./components/ui/label"

async function getData(): Promise<Network[]> {
  // Fetch data from your API here.
  const response = await fetch("/api/v1/vpcs")
  if (!response.ok) {
    throw new Error("Failed to fetch vpcs")
  }
  const vpcs = await response.json()
  return vpcs
}

export default function Page() {
  const [data, setData] = useState<Network[]>([])
  const [open, setOpen] = useState(false)

  useEffect(() => {
    getData().then(setData)
  }, [])

async function handleAddVPC(event: React.FormEvent<HTMLFormElement>) {
  event.preventDefault();
  console.log("Submitting form");
  const formData = new FormData(event.currentTarget);
  const intData: Record<string, any> = {};
  formData.forEach((value, key) => {
    intData[key] = value;
  });
  console.log("Form data:", intData);
  // Create a JSON object from the form data
  const jsonObject = {
    "name": intData.name || "",
    "cidrBlock": intData.cidrBlock || "",
    "description": intData.description || "",
    "dnsServers": intData.dnsServers ? intData.dnsServers.split(",").map((s: string) => s.trim()) : [],
    "domainName": intData.domainName || "",
    "ntpServers": intData.ntpServers ? intData.ntpServers.split(",").map((s: string) => s.trim()) : [],
  };
  console.log("JSON object to send:", jsonObject)

  try {
    const response = await fetch("/api/v1/vpcs", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(jsonObject),
    });
    if (!response.ok) {
      throw new Error("Failed to add VPC");
    }
    // Optionally handle success (e.g., refresh data, close dialog)
  } catch (error) {
    // Optionally handle error (e.g., show error message)
    console.error(error);
  }
  setOpen(false)
  window.location.reload()
}

  return (
    <SidebarProvider>
      <title>Networking</title>
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
                  Networking
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Networks</h2>
                      <div className="flex items-center gap-2">
            <Dialog open={open} onOpenChange={setOpen}>
                <DialogTrigger asChild>
                <Button
                  type="button"
                  className="ml-auto"
                >
                  Create VPC
                </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <form id="add-vpc-form" onSubmit={async (event) => {
                    await handleAddVPC(event);
                  }
                }>
                  <DialogHeader>
                    <DialogTitle>Create VPC</DialogTitle>
                    <DialogDescription>
                      Create a new VPC by filling out the form below.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="grid gap-4" style={{paddingTop: "5px"}}>
                    <div className="grid gap-3">
                      <Label htmlFor="name-1">Name</Label>
                      <Input id="name-1" name="name" defaultValue="" />
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="description-1">Description</Label>
                      <Input id="description-1" name="description" defaultValue="" />
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="cidr-block-1">CIDR Block</Label>
                      <Input id="cidr-block-1" name="cidrBlock" defaultValue="" />
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="dns-servers-1">DNS Servers</Label>
                      <Input id="dns-servers-1" name="dnsServers" defaultValue="" />
                    </div>
                    <div className="grid gap-3">
                      <Label htmlFor="domain-name-1">Domain Name</Label>
                      <Input id="domain-name-1" name="domainName" defaultValue="" />
                    </div>                    
                    <div className="grid gap-3">
                      <Label htmlFor="ntp-servers-1">NTP Servers</Label>
                      <Input id="ntp-servers-1" name="ntpServers" defaultValue="" />
                    </div>
                  </div>
                                  </form>
                 <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button form="add-vpc-form" type="submit">Create VPC</Button>
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