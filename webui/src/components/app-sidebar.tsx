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
    },
    {
      title: "Instances",
      url: "/instances",
      icon: Server,
    },
    {
      title: "Storage",
      url: "/datastores",
      icon: Database,
    },
    {
      title: "Networking",
      url: "/networks",
      icon: Network,
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Settings2,
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
  return (
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
        {/* Determine the current path */}
        {typeof window !== "undefined" && (
          <NavMain
            items={data.navMain.map((item) => ({
              ...item,
              isActive: item.url === window.location.pathname,
              items: [],
            }))}
          />
        )}
      {/*  <NavMain items={data.navMain} />*/}
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
