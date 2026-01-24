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
import { useState } from "react"
import { Field, FieldDescription, FieldLabel } from "./components/ui/field"
import { Input } from "./components/ui/input"
import { useEffect } from "react"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

type Datastore = {
    id: string
    name: string
}

function useDatastores() {
    const [datastores, setDatastores] = useState<Datastore[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        setLoading(true)
        fetch("/api/v1/datastores")
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch datastores")
                return res.json()
            })
            .then((data) => setDatastores(data))
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [])

    return { datastores, loading, error }
}


type VPC = {
    id: string
    name: string
}


function useVPCs() {
    const [vpcs, setVPCs] = useState<VPC[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        setLoading(true)
        fetch("/api/v1/vpcs")
            .then((res) => {
                if (!res.ok) throw new Error("Failed to fetch vpcs")
                return res.json()
            })
            .then((data) => setVPCs(data))
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false))
    }, [])

    return { vpcs, loading, error }
}

export default function Page() {

    const [enabled, setEnabled] = useState(true)
    const { datastores, loading, error } = useDatastores()
    const { vpcs, loading: vpcsLoading, error: vpcsError } = useVPCs()
    const [name, setName] = useState("")
    const [description, setDescription] = useState("")
    const [datastoreId, setDatastoreId] = useState<string | undefined>()
    const [vpcId, setVpcId] = useState<string | undefined>()
    const [submitting, setSubmitting] = useState(false)
    const [submitError, setSubmitError] = useState<string | null>(null)
    // Removed unused submitSuccess state

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setSubmitting(true)
        setSubmitError(null)
        // Removed setSubmitSuccess(false)
        try {
            const res = await fetch("/api/v1/instances", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    name,
                    description,
                    datastoreId: datastoreId,
                    vpcId: vpcId,
                    enabled,
                }),
            })
            if (!res.ok) throw new Error("Failed to create instance")
            // Removed setSubmitSuccess(true)
        } catch (err: any) {
            setSubmitError(err.message)
        } finally {
            setSubmitting(false)
        }
    }
  return (
    <SidebarProvider>
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
                  New Instance
                </BreadcrumbLink>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-8">
            <h2 className="text-3xl font-bold tracking-tight">Create New Instance</h2>
        <div>
            <form className="w-full max-w-md space-y-6" onSubmit={handleSubmit}>
                <Field>
                    <FieldLabel htmlFor="instance-name">Instance Name</FieldLabel>
                    <Input
                        id="instance-name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="Enter instance name"
                    />
                    <FieldDescription>
                        A unique name for your instance.
                    </FieldDescription>
                </Field>
                <Field>
                    <FieldLabel htmlFor="instance-description">Instance Description</FieldLabel>
                    <Input
                        id="instance-description"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="Enter instance description"
                    />
                    <FieldDescription>
                        A unique name for your instance.
                    </FieldDescription>
                </Field>
                <Field>
                    <FieldLabel htmlFor="instance-datastore">Select Datastore</FieldLabel>
                    <Select
                        value={datastoreId}
                        onValueChange={setDatastoreId}
                    >
                        <SelectTrigger className="w-full rounded border border-gray-300 px-3 py-2" id="instance-datastore">
                            <SelectValue placeholder="Select a datastore" />
                        </SelectTrigger>
                        <SelectContent>
                            {!loading && !error && datastores.map((ds) => (
                                <SelectItem key={ds.id} value={ds.id}>
                                    {ds.name} ({ds.id})
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <FieldDescription>
                        Choose a datastore for this instance.
                    </FieldDescription>
                </Field>
                <Field>
                    <FieldLabel htmlFor="instance-vpc">Select VPC</FieldLabel>
                    <Select
                        value={vpcId}
                        onValueChange={setVpcId}
                    >
                        <SelectTrigger className="w-full rounded border border-gray-300 px-3 py-2" id="instance-vpc">
                            <SelectValue placeholder="Select a VPC" />
                        </SelectTrigger>
                        <SelectContent>
                            {!vpcsLoading && !vpcsError && vpcs.map((vpc) => (
                                <SelectItem key={vpc.id} value={vpc.id}>
                                    {vpc.name} ({vpc.id})
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <FieldDescription>
                        Choose a VPC for this instance.
                    </FieldDescription>
                </Field>
                <Field>
                    <div className="flex items-center space-x-2">
                        <input
                            id="instance-enabled"
                            type="checkbox"
                            checked={enabled}
                            onChange={(e) => setEnabled(e.target.checked)}
                            className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                        />
                        <FieldLabel htmlFor="instance-enabled">Enable Instance</FieldLabel>
                    </div>
                    <FieldDescription>
                        Toggle to enable or disable the instance.
                    </FieldDescription>
                </Field>
                <div className="pt-4">
                    <button
                        type="button"
                        className="mr-2 rounded border border-gray-300 bg-white px-4 py-2 font-semibold text-gray-700 hover:bg-gray-100"
                        onClick={() => window.location.href = "/instances"}
                    >
                        Back
                    </button>
                    <button
                        type="submit"
                        className="rounded bg-primary px-4 py-2 font-semibold text-white hover:bg-primary/90"
                        disabled={submitting}
                    >
                        {submitting ? "Creating..." : "Create Instance"}
                    </button>
                </div>
                {submitError && (
                    <div className="mt-2 text-sm text-red-600" role="alert">
                        {submitError}
                    </div>
                )}
            </form>
        </div>
          </div>
      </SidebarInset>
    </SidebarProvider>
  )
}