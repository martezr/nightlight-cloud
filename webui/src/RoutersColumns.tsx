"use client"

import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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
import { ArrowUpDown, MoreHorizontal } from "lucide-react"
import { Link } from "react-router-dom"
import { type Instance } from "./InstancesColumns"

export type { Instance }

const powerColours: Record<string, string> = {
  running:         "bg-emerald-50 text-emerald-700 border-emerald-200",
  stopped:         "bg-gray-100 text-gray-500 border-gray-200",
  paused:          "bg-yellow-50 text-yellow-700 border-yellow-200",
  "shutting-down": "bg-orange-50 text-orange-700 border-orange-200",
  crashed:         "bg-red-50 text-red-700 border-red-200",
}

const powerDots: Record<string, string> = {
  running:         "bg-emerald-500",
  stopped:         "bg-gray-400",
  paused:          "bg-yellow-500",
  "shutting-down": "bg-orange-500",
  crashed:         "bg-red-500",
}

export const routerColumns: ColumnDef<Instance>[] = [
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
      <Link to={`/instances/${row.getValue("id")}`} className="hover:underline font-mono text-xs">
        {row.getValue("id")}
      </Link>
    ),
  },
  {
    accessorKey: "name",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Name
        <ArrowUpDown />
      </Button>
    ),
    cell: ({ row }) => <div>{row.getValue("name")}</div>,
  },
  {
    accessorKey: "powerState",
    header: "Power State",
    cell: ({ row }) => {
      const state: string = row.getValue("powerState") ?? "unknown"
      const cls = powerColours[state] ?? "bg-gray-100 text-gray-500 border-gray-200"
      const dot = powerDots[state] ?? "bg-gray-400"
      return (
        <span className={`inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs font-medium ${cls}`}>
          <span className={`size-1.5 rounded-full ${dot}`} />
          {state}
        </span>
      )
    },
  },
  {
    accessorKey: "primaryIPAddress",
    header: "IP Address",
    cell: ({ row }) => <div className="font-mono">{row.getValue("primaryIPAddress") || "—"}</div>,
  },
  {
    accessorKey: "cpuCores",
    header: "vCPU",
    cell: ({ row }) => <div>{row.getValue("cpuCores")}</div>,
  },
  {
    accessorKey: "memoryMB",
    header: "Memory",
    cell: ({ row }) => <div>{row.getValue("memoryMB")} MB</div>,
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
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link to={`/instances/${instance.id}`}>View Details</Link>
            </DropdownMenuItem>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <DropdownMenuItem
                  onSelect={(e) => e.preventDefault()}
                  className="text-red-600"
                >
                  Delete Router
                </DropdownMenuItem>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete the
                    router <span className="font-medium">{instance.name}</span>.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    className="bg-red-600 hover:bg-red-700"
                    onClick={async () => {
                      await fetch(`/api/v1/instances/${instance.id}`, { method: "DELETE" })
                      window.location.reload()
                    }}
                  >
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </DropdownMenuContent>
        </DropdownMenu>
      )
    },
  },
]
