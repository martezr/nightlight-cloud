"use client"

import type { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { ArrowUpDown } from "lucide-react"
import { Link } from "react-router-dom"

// This type is used to define the shape of our data.
// You can use a Zod schema here if you want.
export type NetworkFlow = {
  bytes: string
  dst_ip: string
  dst_mac: string
  src_ip: string
  src_mac: string
  timestamp: string
  icmp_code: string
  icmp_type: string[]
  dns_query: string
  http_method: string
  http_url: string
  http_user_agent: string
  protocol: string
}

export const columns: ColumnDef<NetworkFlow>[] = [
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
          to={`/networks/${row.getValue("id")}`}
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
      <div>
        {row.getValue("name")}
      </div>
    ),
  },
  {
    accessorKey: "src_ip",
    header: "Source IP",
    cell: ({ row }) => (
      <div >{row.getValue("src_ip")}</div>
    ),
  },
  {
    accessorKey: "dst_ip",
    header: "Destination IP",
    cell: ({ row }) => (
      <div >{row.getValue("dst_ip")}</div>
    ),
  },
]