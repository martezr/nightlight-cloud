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
import { columns, type Instance } from "./InstancesColumns"
import { DataTable } from "./InstancesDataTable"
import { useEffect, useState } from "react"
import { Button } from "./components/ui/button"
import { Link } from 'react-router-dom'
import { RefreshCw } from "lucide-react"

async function getData(): Promise<Instance[]> {
  // Fetch data from your API here.
  const response = await fetch("/api/v1/instances")
  if (!response.ok) {
    throw new Error("Failed to fetch instances")
  }
  const instances = await response.json()
  return instances
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
      <title>Instances</title>
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
                  Instances
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
          <div className="flex items-center justify-between space-y-2">
            <h2 className="text-3xl font-bold tracking-tight">Instances</h2>
              <div className="flex items-center justify-center gap-2">
                <Button onClick={() => window.location.reload()}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
              <Link to="/instances/createinstance">
                <Button
                  type="button"
                  className="ml-auto"
                >
                  Create Instance
                </Button>
              </Link>
          </div>
          </div>
            <DataTable columns={columns} data={data} />
          </div>
      </SidebarInset>
    </SidebarProvider>
  )
}