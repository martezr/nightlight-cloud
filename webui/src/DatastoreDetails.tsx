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
import { DataTable } from "./DatastoreDetailsDataTable"
import { columns, type DatastoreFile } from "./DatastoreDetailsColumns"

function useDatastoreDetails(id?: string, refreshKey?: number) {
    const [data, setData] = useState<DatastoreFile[] | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!id) return
        setLoading(true)
        setError(null)
        fetch(`/api/v1/datastores/${id}/files`)
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch datastore details")
                return res.json()
            })
            .then(setData)
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [id, refreshKey])

    return { data, loading, error }
}

export default function Page() {
const { id } = useParams();
const [refreshKey, setRefreshKey] = useState(0);
const { data, loading, error } = useDatastoreDetails(id, refreshKey);
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
                  <BreadcrumbLink href="/datastores" className="font-bold">
                    Datastores
                  </BreadcrumbLink>
                </BreadcrumbItem>
                 <BreadcrumbSeparator />
                <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink href={`/datastores/${id}`} className="font-bold">
                      {data ? `${id}` : loading ? "Loading..." : "Error"}
                    </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {/* <h2 className="text-3xl font-bold tracking-tight">Datastore Details: {id}</h2>
          <h2 className="text-2xl font-bold tracking-tight">Files</h2> */}
            <DataTable
              columns={columns}
              data={data ?? []}
              datastoreId={id}
              onUploadComplete={() => setRefreshKey((k) => k + 1)}
            />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}