"use client"

import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"

type VPC = {
  id: string
  name: string
}

interface NetworkInterfaceModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  addInterface: (ni: any) => void
}

export function NetworkInterfaceModal({
  open,
  onOpenChange,
  addInterface,
}: NetworkInterfaceModalProps) {
  const [niConnected, setNiConnected] = useState(true)
  const [niBridge, setNiBridge] = useState("")
  const [niModel, setNiModel] = useState("virtio")
  const [niMac, setNiMac] = useState("")

  const [vpcs, setVPCs] = useState<VPC[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Fetch VPCs from API
  useEffect(() => {
    setLoading(true)
    fetch("/api/v1/vpcs")
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch VPCs")
        return res.json()
      })
      .then((data) => setVPCs(data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  function handleAddInterface() {
    addInterface({
//      id: crypto.randomUUID(),
//      bootOrder: 1,
      connected: niConnected,
//      bridgeName: niBridge,
      bridgeName: "nightlight",
      model: niModel,
      mac: niMac,
    })
    // Reset fields
    setNiBridge("")
    setNiModel("virtio")
    setNiMac("")
    setNiConnected(true)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Add Network Interface</DialogTitle>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {/* Bridge */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Bridge</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niBridge}
                onValueChange={setNiBridge}
                disabled={loading || !!error}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={loading ? "Loading..." : "Select bridge"} />
                </SelectTrigger>
                <SelectContent>
                  {!loading && !error
                    ? vpcs.map((v) => (
                        <SelectItem key={v.id} value={v.id}>
                          {v.name} ({v.id})
                        </SelectItem>
                      ))
                    : null}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Model */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Model</Label>
            <div className="col-span-2 w-full">
              <Select
                value={niModel}
                onValueChange={setNiModel}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="virtio">virtio</SelectItem>
                  <SelectItem value="e1000">e1000</SelectItem>
                  <SelectItem value="rtl8139">rtl8139</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* MAC Address */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">MAC Address</Label>
            <Input
              className="col-span-2 w-full"
              value={niMac}
              onChange={(e) => setNiMac(e.target.value)}
              placeholder="Optional"
            />
          </div>

          {/* Connected */}
          <div className="grid grid-cols-3 items-center gap-2">
            <Label className="text-right">Connected</Label>
            <div className="col-span-2">
              <Switch checked={niConnected} onCheckedChange={setNiConnected} />
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleAddInterface} disabled={!niBridge}>
            Add Interface
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}