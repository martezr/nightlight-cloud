import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { BadgeCheckIcon, ChevronRightIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item"
import { Label } from "./components/ui/label"
import { Checkbox } from "./components/ui/checkbox"
import { useState } from "react"

export default function Page() {
        const [enabled, setEnabled] = useState(true)
      const [url, setUrl] = useState("")
      const [username, setUsername] = useState("")
      const [password, setPassword] = useState("")
  return (
    <SidebarProvider>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="#" className="font-bold">
                    Settings
                  </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      <Item variant="outline">
        <ItemContent>
          <ItemTitle>Basic Item</ItemTitle>
          <ItemDescription>
            A simple item with title and description.
          </ItemDescription>
        </ItemContent>
        <ItemActions>
          <Button variant="outline" size="sm">
            Action
          </Button>
        </ItemActions>
      </Item>
      <Item variant="outline" size="sm" asChild>
        <a href="#">
          <ItemMedia>
            <BadgeCheckIcon className="size-5" />
          </ItemMedia>
          <ItemContent>
            <ItemTitle>Your profile has been verified.</ItemTitle>
          </ItemContent>
          <ItemActions>
            <ChevronRightIcon className="size-4" />
          </ItemActions>
        </a>
      </Item>

      <h2 className="text-2xl font-bold">Image Repository</h2>
      <p className="text-muted-foreground">
        Manage your image repository settings and configurations.
      </p>
      <Label className="hover:bg-accent/50 flex items-start gap-3 rounded-lg border p-3 has-[[aria-checked=true]]:border-blue-600 has-[[aria-checked=true]]:bg-blue-50 dark:has-[[aria-checked=true]]:border-blue-900 dark:has-[[aria-checked=true]]:bg-blue-950">
        <Checkbox
          id="toggle-2"
          checked={enabled}
          onCheckedChange={checked => setEnabled(checked === true)}
          className="data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600 data-[state=checked]:text-white dark:data-[state=checked]:border-blue-700 dark:data-[state=checked]:bg-blue-700"
        />
        <div className="grid gap-1.5 font-normal flex-1">
          <p className="text-sm leading-none font-medium">
            Enable custom image repository
          </p>
          <p className="text-muted-foreground text-sm">
            You can enable or disable a custom image repository at any time.
          </p>

        </div>
      </Label>
                {enabled && (
            <div className="mt-3 grid gap-2">
              <div>
                <Label htmlFor="repo-url" className="text-xs mb-1">Repository URL</Label>
                <input
                  id="repo-url"
                  type="text"
                  value={url}
                  onChange={e => setUrl(e.target.value)}
                  className="w-full rounded border px-2 py-1 text-sm"
                  placeholder="https://your-repo-url"
                />
              </div>
              <div>
                <Label htmlFor="repo-username" className="text-xs mb-1">Username</Label>
                <input
                  id="repo-username"
                  type="text"
                  value={username}
                  onChange={e => setUsername(e.target.value)}
                  className="w-full rounded border px-2 py-1 text-sm"
                  placeholder="Username"
                />
              </div>
              <div>
                <Label htmlFor="repo-password" className="text-xs mb-1">Password</Label>
                <input
                  id="repo-password"
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  className="w-full rounded border px-2 py-1 text-sm"
                  placeholder="Password"
                />
              </div>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}