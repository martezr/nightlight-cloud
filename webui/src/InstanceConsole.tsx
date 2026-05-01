import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { useParams } from "react-router-dom"
import { useEffect, useState, useRef } from "react"

import { type Instance } from "./InstancesColumns"
import { VncScreen } from "react-vnc"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./components/ui/alert-dialog"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./components/ui/dropdown-menu"

import { Button } from "./components/ui/button"

import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./components/ui/dialog"

import { Field, FieldGroup } from "./components/ui/field"
import { useForm } from "react-hook-form"
import { Label } from "./components/ui/label"
import { Textarea } from "./components/ui/textarea"

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
  const { id } = useParams()
  const { data, loading } = useInstanceDetails(id)

  const ref = useRef<any>(null)

  const sendCtrlAltDel = () => {
    if (ref.current) {
      ref.current.sendCtrlAltDel()
    }
  }

  // Dialog state
  const [pasteOpen, setPasteOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)

  const form = useForm()

  // Actions
  const handleStart = async () => {
    if (!id) return
    await fetch(`/api/v1/instances/${id}/start`, { method: "POST" })
  }

  const handleStop = async () => {
    if (!id) return
    await fetch(`/api/v1/instances/${id}/stop`, { method: "POST" })
  }

  const handleRestart = async () => {
    if (!id) return
    await fetch(`/api/v1/instances/${id}/restart`, { method: "POST" })
  }

  const handleDelete = async () => {
    if (!id) return
    await fetch(`/api/v1/instances/${id}`, { method: "DELETE" })
    window.location.reload()
  }

  const onSubmit = async (formData: any) => {
    if (!id) return

    await fetch(`/api/v1/instances/${id}/sendkeys`, {
      method: "POST",
      body: JSON.stringify({ keyCode: formData.text }),
    })

    setPasteOpen(false)
  }

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
<header className="flex h-16 items-center px-4">
    <div className="flex items-center gap-2">

        <Separator orientation="vertical" className="h-4" />
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink href="/instances" className="font-bold">Instances</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbLink href={`/instances/${id}`} className="font-bold">
                {data ? data.name : loading ? "Loading..." : "Error"}
              </BreadcrumbLink>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </div>
      
        {/* Actions Menu */}
          <div className="ml-auto">

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button>Actions</Button>
          </DropdownMenuTrigger>

          <DropdownMenuContent>
            <DropdownMenuGroup>
              <DropdownMenuLabel>Power Management</DropdownMenuLabel>

              <DropdownMenuItem onSelect={handleStart}>
                Start
              </DropdownMenuItem>

              <DropdownMenuItem onSelect={handleStop}>
                Stop
              </DropdownMenuItem>

              <DropdownMenuSeparator />

              <DropdownMenuLabel>Instance Management</DropdownMenuLabel>

              <DropdownMenuItem onSelect={sendCtrlAltDel}>
                Send Ctrl+Alt+Del
              </DropdownMenuItem>

              <DropdownMenuItem
                onSelect={() => {
                  setPasteOpen(true)
                }}
              >
                Paste Text
              </DropdownMenuItem>

              <DropdownMenuItem onSelect={handleRestart}>
                Restart Instance
              </DropdownMenuItem>

              <DropdownMenuSeparator />

              <DropdownMenuItem
                className="text-red-600"
                onSelect={() => {
                  setDeleteOpen(true)
                }}
              >
                Delete Instance
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      </header>

      {/* VNC */}
      <div className="flex-1 overflow-hidden" style={{ padding: "10px" }}>
        <VncScreen
          url={`ws://10.0.0.237/ws/${id}`}
          scaleViewport
          background="#000000"
          ref={ref}
          style={{ width: "100%", height: "100%" }}
        />
      </div>

      {/* Paste Dialog */}
      <Dialog open={pasteOpen} onOpenChange={setPasteOpen}>
        <DialogContent className="sm:max-w-sm flex flex-col max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>Paste Text</DialogTitle>
            <DialogDescription>
              Paste text into the instance console.
            </DialogDescription>
          </DialogHeader>

          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col flex-1 overflow-hidden"
          >
            {/* Scrollable content */}
            <div className="flex-1 overflow-y-auto">
              <FieldGroup>
                <Field>
                  <Label>Text</Label>
                  <Textarea
                    className="min-h-[120px]"
                    {...form.register("text")}
                  />
                </Field>
              </FieldGroup>
            </div>

            {/* Footer stays fixed */}
            <DialogFooter className="mt-4">
              <DialogClose asChild>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
              <Button type="submit">Send</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete: {data?.name}
            </AlertDialogDescription>
          </AlertDialogHeader>

          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}