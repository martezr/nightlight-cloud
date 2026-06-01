"use client"

import * as React from "react"
import {
  Command,
  House,
  Send,
  LifeBuoy,
  Settings2,
  Database,
  Server,
  Network,
  Share2,
  Building2,
  SquareTerminal,
  X,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"
import { NavSecondary } from "./nav-secondary"
import { Button } from "@/components/ui/button"
import TerminalComponent from "./terminal"

const data = {
  user: {
    name: "Martez Reed",
    email: "mreed@kalvar.io",
    avatar: "https://avatars.githubusercontent.com/u/5444942?v=4",
  },
  navMain: [
    {
      title: "Home",
      url: "/",
      icon: House,
      items: [],
    },
    {
      title: "Sites",
      url: "/sites",
      icon: Building2,
      items: [],
    },
    {
      title: "Instances",
      url: "/instances",
      icon: Server,
      items: [],
    },
    {
      title: "Storage",
      url: "/datastores",
      icon: Database,
      items: [],
    },
    {
      title: "Networking",
      url: "#",
      icon: Network,
      items: [
        { title: "Networks", url: "/networks" },
        { title: "Routers", url: "/networking/routers" },
        { title: "Switches", url: "/networking/switches" },
        { title: "Subnets", url: "/networking/subnets" },
      ],
    },
    {
      title: "Topology",
      url: "/topology",
      icon: Share2,
      items: [],
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Settings2,
      items: [],
    },
  ],
  navSecondary: [
    {
      title: "Support",
      url: "#",
      icon: LifeBuoy,
    },
    {
      title: "Feedback",
      url: "#",
      icon: Send,
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const [shellOpen, setShellOpen] = React.useState(false)

  return (
    <>
      <Sidebar variant="inset" {...props}>
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <a href="#">
                  <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                    <Command className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">Kalvar</span>
                    <span className="truncate text-xs">Nightlight</span>
                  </div>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          {typeof window !== "undefined" && (
            <NavMain
              items={data.navMain.map((item) => {
                const subItems = item.items.map((sub) => ({
                  ...sub,
                  isActive:
                    window.location.pathname === sub.url ||
                    window.location.pathname.startsWith(sub.url + "/"),
                }))
                const isActive =
                  item.url !== "#"
                    ? window.location.pathname === item.url ||
                      window.location.pathname.startsWith(item.url + "/")
                    : subItems.some((s) => s.isActive)
                return { ...item, isActive, items: subItems }
              })}
            />
          )}
          <NavSecondary items={data.navSecondary} className="mt-auto" />
        </SidebarContent>
        <SidebarFooter>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                tooltip="Cloud Shell"
                onClick={() => setShellOpen((o) => !o)}
              >
                <SquareTerminal />
                <span>Cloud Shell</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
          <NavUser user={data.user} />
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>

      {/*
        Hand-rolled fixed overlay instead of a vaul Drawer.
        Vaul uses flex+height:100% chains that prevent FitAddon from measuring
        an explicit pixel height, causing xterm to compute ~20 cols.
        With fixed+inset-x-0 the overlay has definite viewport-width dimensions,
        and the inner absolute+inset-0 wrapper gives terminalRef.current an
        explicit pixel bounding box that FitAddon can measure correctly.
      */}
      {shellOpen && (
        <div className="fixed inset-x-0 bottom-0 z-50 flex h-[480px] flex-col border-t bg-background shadow-2xl">
          {/* Header */}
          <div className="flex shrink-0 items-center justify-between border-b px-4 py-2">
            <div className="flex items-center gap-2">
              <SquareTerminal className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-medium">Cloud Shell</span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => setShellOpen(false)}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          {/*
            relative parent → gives the absolute child a positioned ancestor
            whose pixel dimensions are always known (filled by flex-1).
            absolute inset-0 → terminalRef.current's parent has an explicit
            clientWidth/clientHeight that getComputedStyle() resolves correctly.
          */}
          <div className="relative min-h-0 flex-1">
            <div className="absolute inset-0 p-2">
              <TerminalComponent />
            </div>
          </div>
        </div>
      )}
    </>
  )
}
