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
import { Tabs, TabsContent, TabsTrigger, TabsList } from "@/components/ui/tabs"
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "./components/ui/dropdown-menu"
import { Button } from "./components/ui/button"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "./components/ui/alert-dialog"
import { type Instance } from "./InstancesColumns"
import { useRef } from 'react';
import { VncScreen } from 'react-vnc';
import { Link } from "react-router-dom";
import { InstanceNetworkInterfacesDataTable } from "./InstanceNetworkInterfacesDataTable"
import { instanceNetworkInterfacsColumns } from "./InstanceNetworkInterfacesColumns"

function useInstanceDetails(id?: string) {
    const [data, setData] = useState<Instance | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!id) return
        setLoading(true)
        setError(null)
        fetch(`/api/v1/instances/${id}`)
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch instance details")
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
const { data, loading, error } = useInstanceDetails(id);
  const ref = useRef<any>(null);

const sendCtrlAltDel = () => {
    if (ref.current) {
      ref.current.sendCtrlAltDel();
    } else {
      console.error("RFB object not available yet.");
    }
  }

console.log("Instance data:", error);
  return (
    <SidebarProvider>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 border-b">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="/instances" className="font-bold">
                    Instances
                  </BreadcrumbLink>
                </BreadcrumbItem>
                 <BreadcrumbSeparator />
                <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink href={`/instances/${id}`} className="font-bold">
                      {data ? data.name : loading ? "Loading..." : "Error"}
                    </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-4">
                    <div className="flex items-center justify-between space-y-2">

                <h2 className="text-3xl font-bold tracking-tight">{data ? data.name : loading ? "Loading..." : "Error"}</h2>
                                      <div className="flex items-center gap-2">
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button >Actions</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuGroup>
          <DropdownMenuLabel>Power Management</DropdownMenuLabel>
            <DropdownMenuItem
            onClick={async () => {
              if (!id) return;
              try {
              await fetch(`/api/v1/instances/${id}/start`, { method: "POST" });
              // Optionally, show a notification or refresh data here
              } catch (err) {
              // Optionally, handle error here
              console.error("Failed to start instance", err);
              }
            }}
            >
            Start
            </DropdownMenuItem>
            <DropdownMenuItem
            onClick={async () => {
              if (!id) return;
              try {
              await fetch(`/api/v1/instances/${id}/stop`, { method: "POST" });
              // Optionally, show a notification or refresh data here
              } catch (err) {
              // Optionally, handle error here
              console.error("Failed to stop instance", err);
              }
            }}
            >
            Stop
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          <DropdownMenuLabel>Instance Management</DropdownMenuLabel>
            <DropdownMenuItem onClick={sendCtrlAltDel}>
            Send Ctrl+Alt+Del
            </DropdownMenuItem>
            <DropdownMenuItem>
                      <Link
          to={`/instances/${id}/console`}
        >
          Open Console
        </Link>
            </DropdownMenuItem>
          <DropdownMenuItem
            onClick={async () => {
              if (!id) return;
              try {
              await fetch(`/api/v1/instances/${id}/restart`, { method: "POST" });
              // Optionally, show a notification or refresh data here
              } catch (err) {
              // Optionally, handle error here
              console.error("Failed to restart instance", err);
              }
            }}
            >
            Restart Instance
            </DropdownMenuItem>
              <DropdownMenuSeparator />
        <AlertDialog>
          <AlertDialogTrigger asChild>
                <DropdownMenuItem
                  onClick={async (event) => {
                  event.preventDefault();
                  await fetch(`/api/v1/instances/${id}`, {
                    method: "DELETE",
                  })
                  window.location.reload()
                  }}
                  className="text-red-600"
                >
                  Delete Instance
                </DropdownMenuItem>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. This will permanently delete the following instance: {data?.name}.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction>Delete</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
</div>
</div>
      <Tabs defaultValue="overview" className="relative mr-auto w-full">
        <TabsList className="w-full justify-start rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            value="overview"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Overview
          </TabsTrigger>
          <TabsTrigger
            value="console"
            className="relative rounded-none border-b-2 border-b-transparent bg-transparent px-4 pb-3 pt-2 font-semibold text-muted-foreground shadow-none transition-none focus-visible:ring-0 data-[state=active]:border-b-primary data-[state=active]:text-foreground data-[state=active]:shadow-none "
          >
            Console
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
          <TabsContent style={{ padding: "1rem" }} value="overview">
            {loading && <p>Loading instance details...</p>}
            {error && <p className="text-red-500">{error}</p>}

            {data && (
              <div className="grid gap-6">
                {/* Top Summary Cards */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="rounded-2xl border p-4 shadow-sm">
                    <p className="text-sm text-muted-foreground">Status</p>
                    <p className="text-xl font-semibold">{data.status}</p>
                  </div>

                  <div className="rounded-2xl border p-4 shadow-sm">
                    <p className="text-sm text-muted-foreground">Instance Type</p>
                    <p className="text-xl font-semibold">Test</p>
                  </div>

                  <div className="rounded-2xl border p-4 shadow-sm">
                    <p className="text-sm text-muted-foreground">Region</p>
                    <p className="text-xl font-semibold">Test</p>
                  </div>
                </div>

                {/* Details Section */}
                <div className="rounded-2xl border p-6 shadow-sm">
                  <h3 className="text-lg font-semibold mb-4">Instance Details</h3>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-muted-foreground">Instance ID</p>
                      <p className="font-medium">{data.id}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Name</p>
                      <p className="font-medium">{data.name}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Image / OS</p>
                      <p className="font-medium">{data.status || "N/A"}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Created At</p>
                      <p className="font-medium">
                        {new Date(data.status).toLocaleString()}
                      </p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Availability Zone</p>
                      <p className="font-medium">{data.status || "N/A"}</p>
                    </div>
                  </div>
                </div>

                {/* Compute Section */}
                <div className="rounded-2xl border p-6 shadow-sm">
                  <h3 className="text-lg font-semibold mb-4">Compute</h3>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                    <div>
                      <p className="text-muted-foreground">vCPU</p>
                      <p className="font-medium">{data.cpuCores}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Memory</p>
                      <p className="font-medium">{data.memoryMB} MB</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Architecture</p>
                      <p className="font-medium">x86_64</p>
                    </div>
                  </div>
                </div>

                {/* Networking Section */}
                <div className="rounded-2xl border p-6 shadow-sm">
                  <h3 className="text-lg font-semibold mb-4">Networking</h3>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-muted-foreground">Private IP</p>
                      <p className="font-medium">{data.status}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Public IP</p>
                      <p className="font-medium">{data.primaryIPAddress || "N/A"}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">VPC</p>
                      <p className="font-medium">{data.devices.networkInterfaces[0].vpcId || "N/A"}</p>
                    </div>

                    <div>
                      <p className="text-muted-foreground">Subnet</p>
                      <p className="font-medium">{data.status || "N/A"}</p>
                    </div>
                  </div>
                </div>

                {/* Tags Section */}
                {/* {data.tags && (
                  <div className="rounded-2xl border p-6 shadow-sm">
                    <h3 className="text-lg font-semibold mb-4">Tags</h3>

                    <div className="flex flex-wrap gap-2">
                      {Object.entries(data.tags).map(([key, value]) => (
                        <span
                          key={key}
                          className="rounded-full border px-3 py-1 text-xs"
                        >
                          {key}: {value}
                        </span>
                      ))}
                    </div>
                  </div>
                )} */}
              </div>
            )}
          </TabsContent>
        <TabsContent style={{padding: '1rem'}} value="console">
          <div>
            <VncScreen
              url={`ws://10.0.0.237/ws/${id}`}
              scaleViewport
              background="#000000"
              style={{
                width: '75vw',
                height: '75vh',
              }}
              ref={ref}
            />
          </div>
        </TabsContent>
        <TabsContent style={{padding: '1rem'}} value="networking">
            <InstanceNetworkInterfacesDataTable columns={instanceNetworkInterfacsColumns} data={data?.devices.networkInterfaces || []} />
        </TabsContent>
        <TabsContent style={{padding: '1rem'}} value="storage">
        </TabsContent>
        <TabsContent style={{padding: '1rem'}} value="snapshots">
        </TabsContent>
      </Tabs>

        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}