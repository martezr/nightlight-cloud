"use client"

import type { ColumnDef } from "@tanstack/react-table"
import { Checkbox } from "@/components/ui/checkbox"
import { Link } from "react-router-dom"
import { type InstanceNetworkInterface } from "./InstancesColumns"

export const instanceNetworkInterfacsColumns: ColumnDef<InstanceNetworkInterface>[] = [
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
    accessorKey: "macAddress",
    header: "MAC Address",
    cell: ({ row }) => (
      <div >{row.getValue("macAddress")}</div>
    ),
  },  
  {
    accessorKey: "vpcId",
    header: "VPC",
    cell: ({ row }) => (
      <div >{row.getValue("vpcId")}</div>
    ),
  },
  {
    accessorKey: "bridgeName",
    header: "Bridge Name",
    cell: ({ row }) => (
      <div >{row.getValue("bridgeName")}</div>
    ),
  },
  {
    accessorKey: "model",
    header: "Model",
    cell: ({ row }) => (
      <div >{row.getValue("model")}</div>
    ),
  },
]