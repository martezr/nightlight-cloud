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


interface IntegrationDetails {
    id: string
    name: string
    // Add other fields as needed
}

function useIntegrationDetails(id?: string) {
    const [data, setData] = useState<IntegrationDetails | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!id) return
        setLoading(true)
        setError(null)
        fetch(`/api/v1/integrations/${id}`)
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch integration details")
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
const { data, loading, error } = useIntegrationDetails(id);
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
                  <BreadcrumbLink href="/integrations" className="font-bold">
                    Integrations
                  </BreadcrumbLink>
                </BreadcrumbItem>
                 <BreadcrumbSeparator />
                <BreadcrumbItem className="hidden md:block">
                    <BreadcrumbLink href={`/integrations/${id}`} className="font-bold">
                      {data ? data.name : loading ? "Loading..." : "Error"}
                    </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <div className="grid auto-rows-min gap-4 md:grid-cols-3">
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
            <div className="bg-muted/50 aspect-video rounded-xl" />
          </div>
          <div className="bg-muted/50 min-h-[100vh] flex-1 rounded-xl md:min-h-min" />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}