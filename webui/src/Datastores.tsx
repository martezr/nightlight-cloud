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
import { columns, type Datastore } from "./DatastoresColumns"
import { DataTable } from "./DatastoresDataTable"
import { useEffect, useState } from "react"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "./components/ui/dialog"
import { Button } from "./components/ui/button"
import { Input } from "./components/ui/input"
import { Label } from "./components/ui/label"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "./components/ui/select"

async function getData(): Promise<Datastore[]> {
  // Fetch data from your API here.
  const response = await fetch("/api/v1/datastores")
  if (!response.ok) {
    throw new Error("Failed to fetch datastores")
  }
  const datastores = await response.json()
  return datastores
}

export default function Page() {
  const [data, setData] = useState<Datastore[]>([])
  const [open, setOpen] = useState(false)
  const [datastoreType, setDatastoreType] = useState("")

  useEffect(() => {
    getData().then(setData)
  }, [])

async function handleAddDatastore(event: React.FormEvent<HTMLFormElement>) {
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
    "type": intData.datastoreType || "",
    "description": intData.description || "",
    "localPath": intData.localPath || "",
    "path": intData.path || "",
  };

  try {
    const response = await fetch("/api/v1/datastores", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(jsonObject),
    });
    if (!response.ok) {
      throw new Error("Failed to add datastore");
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
      <title>Storage</title>
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
                  Storage
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Datastores</h2>
                      <div className="flex items-center gap-2">
            <Dialog open={open} onOpenChange={setOpen}>
                <DialogTrigger asChild>
                <Button
                  type="button"
                  className="ml-auto"
                >
                  Create Datastore
                </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <form id="add-datastore-form" onSubmit={async (event) => {
                    await handleAddDatastore(event);
                  }
                }>
                  <DialogHeader>
                    <DialogTitle>Create Datastore</DialogTitle>
                    <DialogDescription>
                      Create a new datastore by filling out the form below.
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
                      <Label htmlFor="datastore-type-value-1">Datastore Type</Label>
                      <Select
                        name="datastoreType"
                        defaultValue=""
                        onValueChange={(value) => {
                          // handle value change here, e.g. set state or form value
                          console.log("Selected datastore type:", value)
                          setDatastoreType(value)
                        }}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Select a value" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem value="local">Local</SelectItem>
                            <SelectItem value="nfs">NFS</SelectItem>
                            <SelectItem value="iscsi">iSCSI</SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                    {datastoreType === "local" && (
                      <div className="grid gap-3">
                        <Label htmlFor="local-path-1">Path</Label>
                        <Input id="local-path-1" name="localPath" defaultValue="" />
                      </div>          
                    )}
                    {datastoreType === "nfs" && (
                      <div className="grid gap-3">
                        <Label htmlFor="nfs-username-1">Username</Label>
                        <Input id="nfs-username-1" name="nfsUsername" defaultValue="" />
                      </div>          
                    )}
                    {datastoreType === "nfs" && (
                      <div className="grid gap-3">
                        <Label htmlFor="nfs-password-1">Password</Label>
                        <Input id="nfs-password-1" name="nfsPassword" type="password" />
                      </div>          
                    )}
                    {datastoreType === "nfs" && (
                      <div className="grid gap-3">
                        <Label htmlFor="nfs-path-1">Path</Label>
                        <Input id="nfs-path-1" name="path" defaultValue="" />
                      </div>          
                    )}
                  </div>
                                  </form>
                 <DialogFooter>
                    <DialogClose asChild>
                      <Button variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button form="add-datastore-form" type="submit">Create datastore</Button>
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