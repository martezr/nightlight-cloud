"use client"

import type { ColumnDef } from "@tanstack/react-table"
import { Checkbox } from "@/components/ui/checkbox"
import { type InstanceDisk } from "./InstancesColumns"

export const instanceDisksColumns: ColumnDef<InstanceDisk>[] = [
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
    cell: ({ row }) => <div>{row.getValue("id")}</div>,
  },
  {
    accessorKey: "sizeGB",
    header: "Size (GB)",
    cell: ({ row }) => <div>{row.getValue("sizeGB")}</div>,
  },
  {
    accessorKey: "busType",
    header: "Bus Type",
    cell: ({ row }) => <div>{row.getValue("busType")}</div>,
  },
  {
    accessorKey: "bootOrder",
    header: "Boot Order",
    cell: ({ row }) => <div>{row.getValue("bootOrder")}</div>,
  },
]
