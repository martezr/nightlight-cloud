"use client"
import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import type { Datastore } from "@/DatastoresColumns"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select"
import { Input } from "./ui/input"

interface StorageDiskModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  addDisk: (disk: {
  //  id: string
    bootOrder: number
    sizeGB: number
    busType: string
    datastoreId: string
  }) => void
}

export function StorageDiskModal({ open, onOpenChange, addDisk }: StorageDiskModalProps) {
//  const [niConnected, setNiConnected] = useState(true)
const [niBootOrder, setNiBootOrder] = useState(1)
  const [niSizeGB, setNiSizeGB] = useState(10)
  const [niBusType, setNiBusType] = useState("virtio")
  const [niDatastoreId, setNiDatastoreId] = useState("")
  const handleAdd = () => {
    addDisk({
     // id: crypto.randomUUID(),
      bootOrder: niBootOrder,
      sizeGB: niSizeGB,
      busType: niBusType,
      datastoreId: niDatastoreId,
    })
    // Reset modal fields
//    setNiConnected(true)
    setNiDatastoreId("")
    setNiBootOrder(1)
    setNiSizeGB(10)
    setNiBusType("virtio")
    onOpenChange(false)
  }

  const [datastores, setDatastores] = useState<Datastore[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add Disk</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {/* Datastore */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Datastore</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niDatastoreId}
                onValueChange={setNiDatastoreId}
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
          {/* Bus Type */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Bus Type</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niBusType}
                onValueChange={setNiBusType}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="virtio">virtio</SelectItem>
                  <SelectItem value="sata">sata</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

            {/* Size */}
            <div className="grid grid-cols-3 items-center gap-2">
              <Label className="text-right">Size (GB)</Label>
              <div className="col-span-2 w-full">
                <Input
                  type="number"
                  value={niSizeGB}
                  onChange={(e) => setNiSizeGB(Number(e.target.value))}
                  min={1}
                />
              </div>
            </div>

            {/* Boot Order */}
            <div className="grid grid-cols-3 items-center gap-2">
              <Label className="text-right">Boot Order</Label>
              <div className="col-span-2 w-full">
                <Input
                  type="number"
                  value={niBootOrder}
                  onChange={(e) => setNiBootOrder(Number(e.target.value))}
                  min={1}
                />
              </div>
            </div>

         </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleAdd} disabled={!niDatastoreId}>Add Disk</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}