import { useEffect, useState } from "react"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { KeyRound, User, Info, Moon } from "lucide-react"

function useCurrentUser() {
  const [username, setUsername] = useState<string | null>(null)
  useEffect(() => {
    fetch("/api/v1/auth/me")
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => d && setUsername(d.username))
      .catch(() => {})
  }, [])
  return username
}

function usePlatformVersion() {
  const [version, setVersion] = useState<string | null>(null)
  useEffect(() => {
    fetch("/api/v1/version")
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => d && setVersion(d.version))
      .catch(() => {})
  }, [])
  return version
}

function ChangePasswordForm() {
  const [current, setCurrent] = useState("")
  const [next, setNext] = useState("")
  const [confirm, setConfirm] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSuccess(false)
    if (next !== confirm) {
      setError("New passwords do not match")
      return
    }
    if (next.length < 8) {
      setError("New password must be at least 8 characters")
      return
    }
    setSaving(true)
    try {
      const res = await fetch("/api/v1/auth/password", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ currentPassword: current, newPassword: next }),
      })
      if (!res.ok) {
        const text = await res.text()
        throw new Error(text.trim() || "Failed to change password")
      }
      setSuccess(true)
      setCurrent("")
      setNext("")
      setConfirm("")
    } catch (err: any) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor="current-password">Current password</Label>
        <Input
          id="current-password"
          type="password"
          autoComplete="current-password"
          value={current}
          onChange={(e) => setCurrent(e.target.value)}
          disabled={saving}
          required
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="new-password">New password</Label>
        <Input
          id="new-password"
          type="password"
          autoComplete="new-password"
          value={next}
          onChange={(e) => setNext(e.target.value)}
          disabled={saving}
          required
        />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="confirm-password">Confirm new password</Label>
        <Input
          id="confirm-password"
          type="password"
          autoComplete="new-password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          disabled={saving}
          required
        />
      </div>
      {error && (
        <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>
      )}
      {success && (
        <p className="rounded-md bg-green-50 px-3 py-2 text-sm text-green-700">
          Password changed successfully.
        </p>
      )}
      <div>
        <Button type="submit" disabled={saving}>
          {saving ? "Saving..." : "Change password"}
        </Button>
      </div>
    </form>
  )
}

export default function Page() {
  const username = useCurrentUser()
  const version = usePlatformVersion()

  return (
    <SidebarProvider>
      <AppSidebar collapsible="icon" variant="sidebar" />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden md:block">
                  <BreadcrumbLink href="/settings" className="font-bold">
                    Settings
                  </BreadcrumbLink>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-8">
          <div>
            <h1 className="text-2xl font-bold">Settings</h1>
            <p className="text-sm text-muted-foreground">Manage your account and platform configuration</p>
          </div>

          <Tabs defaultValue="account" className="w-full max-w-2xl">
            <TabsList className="mb-6">
              <TabsTrigger value="account">Account</TabsTrigger>
              <TabsTrigger value="platform">Platform</TabsTrigger>
            </TabsList>

            {/* Account tab */}
            <TabsContent value="account" className="flex flex-col gap-6">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <CardTitle className="text-base">Your Account</CardTitle>
                  </div>
                  <CardDescription>Current session information</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-4 rounded-lg border p-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground font-semibold text-sm">
                      {username ? username[0].toUpperCase() : "?"}
                    </div>
                    <div>
                      <p className="font-medium">{username ?? "—"}</p>
                      <p className="text-xs text-muted-foreground">Administrator</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <KeyRound className="h-4 w-4 text-muted-foreground" />
                    <CardTitle className="text-base">Change Password</CardTitle>
                  </div>
                  <CardDescription>Must be at least 8 characters</CardDescription>
                </CardHeader>
                <CardContent>
                  <ChangePasswordForm />
                </CardContent>
              </Card>
            </TabsContent>

            {/* Platform tab */}
            <TabsContent value="platform" className="flex flex-col gap-6">
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Info className="h-4 w-4 text-muted-foreground" />
                    <CardTitle className="text-base">Platform Information</CardTitle>
                  </div>
                  <CardDescription>Runtime details for this installation</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <div className="grid grid-cols-2 gap-y-4 text-sm">
                    <div className="text-muted-foreground">Platform</div>
                    <div className="flex items-center gap-2 font-medium">
                      <Moon className="h-4 w-4" />
                      Nightlight Cloud
                    </div>

                    <div className="text-muted-foreground">Version</div>
                    <div>
                      <Badge variant="outline">{version ?? "—"}</Badge>
                    </div>

                    <div className="text-muted-foreground">API</div>
                    <div className="font-mono text-xs text-muted-foreground">v1</div>

                    <div className="text-muted-foreground">Default credentials</div>
                    <div className="font-mono text-xs text-muted-foreground">root / nightlight</div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
