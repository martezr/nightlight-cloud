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
import { routerColumns } from "./RoutersColumns"
import { DataTable } from "./InstancesDataTable"
import { useEffect, useState } from "react"
import { type Instance } from "./InstancesColumns"
import { Button } from "./components/ui/button"
import { RefreshCw } from "lucide-react"

async function getData(): Promise<Instance[]> {
  const response = await fetch("/api/v1/instances")
  if (!response.ok) throw new Error("Failed to fetch instances")
  const instances: Instance[] = await response.json()
  console.log("Fetched instances:", instances)
  console.log(instances.filter((i) => i.instanceType === "router"))
  return instances.filter((i) => i.instanceType === "router")
}

export default function Page() {
  const [data, setData] = useState<Instance[]>([])

  useEffect(() => {
    getData().then(setData)
    const interval = setInterval(() => getData().then(setData), 10_000)
    return () => clearInterval(interval)
  }, [])

  return (
    <SidebarProvider>
      <title>Routers</title>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#" className="font-bold">Routers</BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Routers</h2>
            <Button onClick={() => getData().then(setData)}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          </div>
          <DataTable columns={routerColumns} data={data} />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
