import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { useParams } from 'react-router-dom'
import { useEffect, useState } from "react"
import { type Network } from "./NetworksColumns"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./components/ui/tabs"
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "./components/ui/dropdown-menu"
import { Button } from "./components/ui/button"


function useNetworkDetails(id?: string) {
    const [data, setData] = useState<Network | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!id) return
        setLoading(true)
        setError(null)
        fetch(`/api/v1/vpcs/${id}`)
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch network details")
                return res.json()
            })
            .then(setData)
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [id])

    return { data, loading, error }
}

export default function Page() {
const { id } = useParams();
const { data, loading, error } = useNetworkDetails(id);
if (error) {
    return <div>Error: {error}</div>
}
  return (
    <SidebarProvider>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="/networks" className="font-bold">
                    Networks
                  </BreadcrumbLink>
                </BreadcrumbItem>
                 <BreadcrumbSeparator />
                <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink href={`/networks/${id}`} className="font-bold">
                      {data ? data.name : loading ? "Loading..." : "Error"}
                    </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
                    <div className="flex items-center justify-between space-y-2">

                <h2 className="text-3xl font-bold tracking-tight">Network</h2>
                                      <div className="flex items-center gap-2">

    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button >Actions</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuGroup>
          <DropdownMenuLabel>Power Management</DropdownMenuLabel>
          <DropdownMenuItem>Start Instance</DropdownMenuItem>
          <DropdownMenuItem>Stop Instance</DropdownMenuItem>
          <DropdownMenuItem>Restart Instance</DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem>GitHub</DropdownMenuItem>
        <DropdownMenuItem>Support</DropdownMenuItem>
        <DropdownMenuItem disabled>API</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
</div>
</div>
      <Tabs defaultValue="account" className="relative mr-auto w-full">
        <TabsList className="w-full justify-start rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            value="account"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Overview
          </TabsTrigger>
          <TabsTrigger
            value="networking"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Networking
          </TabsTrigger>
          <TabsTrigger
            value="storage"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Storage
          </TabsTrigger>
          <TabsTrigger
            value="snapshots"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Snapshots
          </TabsTrigger>
        </TabsList>
        <TabsContent value="account">
          View and update your account settings here.
        </TabsContent>
        <TabsContent value="networking">
          Manage your networking preferences here.
        </TabsContent>
        <TabsContent value="storage">
          <div className="grid auto-rows-min gap-4 md:grid-cols-3">
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
          </div>
        </TabsContent>
        <TabsContent value="snapshots">
                    <div className="grid auto-rows-min gap-4 md:grid-cols-3">
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
          </div>
        </TabsContent>
      </Tabs>

        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}