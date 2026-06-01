"use client"

import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { ArrowUpDown, MoreHorizontal } from "lucide-react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Link } from "react-router-dom"
//import { Input } from "@/components/ui/input"

// This type is used to define the shape of our data.
// You can use a Zod schema here if you want.
export type Instance = {
  id: string
  name: string
  type: string
  status: string
  version: string
  description: string
  initializationStatus: string
  powerState: string
  instanceType: string
  cpuCores: string
  cpuSockets: string
  memoryMB: string
  devices: InstanceDevices
  primaryIPAddress: string
  primaryMacAddress: string
}

export type InstanceDevices = {
  networkInterfaces: InstanceNetworkInterface[]
  cdroms: InstanceCDROM[]
  storageDisks: InstanceDisk[]
}

export type InstanceNetworkInterface = {
  id: string
  subnetId: string
  macAddress: string
  bridgeName: string
  connected: boolean
  bootOrder: number
  model: string
}

export type InstanceCDROM = {
  id: string
  name: string
  connected: boolean
}

export type InstanceDisk = {
  id: string
  sizeGB: number
  busType: string
  bootOrder: number
}

export const columns: ColumnDef<Instance>[] = [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "id",
    header: "ID",
    cell: ({ row }) => (
      <div>
        <Link
          to={`/instances/${row.getValue("id")}`}
          className="hover:underline"
        >
          {row.getValue("id")}
        </Link>
      </div>
    ),
  },
  {
    accessorKey: "name",
    header: ({ column }) => {
      return (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Name
          <ArrowUpDown />
        </Button>
      )
    },
    cell: ({ row }) => (
      <div >{row.getValue("name")}</div>
    ),
  },
  {
    accessorKey: "powerState",
    header: "Power State",
    cell: ({ row }) => {
      const state: string = row.getValue("powerState") ?? "unknown"
      const colours: Record<string, string> = {
        running:       "bg-emerald-50 text-emerald-700 border-emerald-200",
        stopped:       "bg-gray-100 text-gray-500 border-gray-200",
        paused:        "bg-yellow-50 text-yellow-700 border-yellow-200",
        "shutting-down": "bg-orange-50 text-orange-700 border-orange-200",
        crashed:       "bg-red-50 text-red-700 border-red-200",
      }
      const dotColours: Record<string, string> = {
        running:       "bg-emerald-500",
        stopped:       "bg-gray-400",
        paused:        "bg-yellow-500",
        "shutting-down": "bg-orange-500",
        crashed:       "bg-red-500",
      }
      const cls = colours[state] ?? "bg-gray-100 text-gray-500 border-gray-200"
      const dot = dotColours[state] ?? "bg-gray-400"
      return (
        <span className={`inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs font-medium ${cls}`}>
          <span className={`size-1.5 rounded-full ${dot}`} />
          {state}
        </span>
      )
    },
  },
  {
    accessorKey: "description",
    header: "Description",
    cell: ({ row }) => (
      <div >{row.getValue("description")}</div>
    ),
  },
  {
    accessorKey: "cpuSockets",
    header: "CPU Sockets",
    cell: ({ row }) => (
      <div >{row.getValue("cpuSockets")}</div>
    ),
  },
  {
    accessorKey: "cpuCores",
    header: "CPU Cores",
    cell: ({ row }) => (
      <div >{row.getValue("cpuCores")}</div>
    ),
  },
  {
    accessorKey: "memoryMB",
    header: "Memory",
    cell: ({ row }) => (
      <div >{row.getValue("memoryMB")}</div>
    ),
  },
  {
    accessorKey: "primaryIPAddress",
    header: "IP Address",
    cell: ({ row }) => (
      <div >{row.getValue("primaryIPAddress")}</div>
    ),
  },
  {
    id: "actions",
    header: "Actions",
    enableHiding: false,
    cell: ({ row }) => {
      const instance = row.original
      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>Edit Instance</DropdownMenuItem>
    <AlertDialog>
      <AlertDialogTrigger asChild>
            <DropdownMenuItem
              onClick={async (event) => {
               event.preventDefault();
              await fetch(`/api/v1/instances/${instance.id}`, {
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
            This action cannot be undone. This will permanently delete the following instance: {instance.name}.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction>Delete</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

          </DropdownMenuContent>
        </DropdownMenu>
      )
    },
  },
]