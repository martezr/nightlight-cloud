"use client"
import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import type { Datastore } from "@/DatastoresColumns"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select"

interface CdromModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  addCdrom: (cdrom: {
  //  id: string
    bootOrder: number
    connected: boolean
    path: string
  }) => void
}

// This type is used to define the shape of our data.
// You can use a Zod schema here if you want.
export type DatastoreFile = {
  name: string
  path: string
  size: number
}

export function CdromModal({ open, onOpenChange, addCdrom }: CdromModalProps) {
  const [niConnected, setNiConnected] = useState(true)
  const [niDatastore, setNiDatastore] = useState("")
  const [niDatastoreFile, setNiDatastoreFile] = useState("")
  const handleAdd = () => {
    addCdrom({
     // id: crypto.randomUUID(),
      bootOrder: 1,
      connected: niConnected,
      path: niDatastoreFile,
    })
    // Reset modal fields
    setNiConnected(true)
    setNiDatastore("")
    onOpenChange(false)
  }

  const [datastores, setDatastores] = useState<Datastore[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [files, setFiles] = useState<DatastoreFile[]>([])

  // Fetch datastores from API
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

  // Fetch files from API when datastore changes
  useEffect(() => {
    if (!niDatastore) return
    // You can implement fetching files for the selected datastore here if needed
    fetch(`/api/v1/datastores/${niDatastore}/files`)
        .then((res) => {
            if (!res.ok) throw new Error("Failed to fetch files")
            return res.json()
        })
        .then((data) => {
            // Handle files data as needed
            console.log("Files for datastore", niDatastore, data)
            //const isoFiles = data.filter((file: DatastoreFile) => file.name.endsWith('.iso'))
            const isoFiles = data.filter((file: DatastoreFile) => file.name.endsWith('.iso') && !file.name.startsWith('._'))
            if (isoFiles.length === 0) {
                setFiles([])
            } else {
                setFiles(isoFiles)
            }
            //setFiles(data)
        })
        .catch((err) => setError(err.message))
  }, [niDatastore])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add CD-ROM</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {/* Datastore */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Datastore</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niDatastore}
                onValueChange={setNiDatastore}
                disabled={loading || !!error}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={loading ? "Loading..." : "Select datastore"} />
                </SelectTrigger>
                <SelectContent>
                  {!loading && !error
                    ? datastores.map((v) => (
                        <SelectItem key={v.id} value={v.id}>
                          {v.name} ({v.id})
                        </SelectItem>
                      ))
                    : null}
                </SelectContent>
              </Select>
            </div>
          </div>
          {/* ISO Name */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">ISO Name</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niDatastoreFile}
                onValueChange={setNiDatastoreFile}
                disabled={loading || !!error}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={loading ? "Loading..." : "Select file"} />
                </SelectTrigger>
                <SelectContent>
                  {!loading && !error
                    ? files.map((v) => (
                        <SelectItem key={v.name} value={v.path}>
                          {v.name} ({v.size} bytes)
                        </SelectItem>
                      ))
                    : null}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleAdd} disabled={!niDatastore || !niDatastoreFile}>Add CD-ROM</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}