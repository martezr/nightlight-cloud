import { useEffect, useState } from "react"
import { Navigate } from "react-router-dom"
import { checkAuth } from "@/lib/auth"

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<"loading" | "ok" | "unauth">("loading")

  useEffect(() => {
    checkAuth().then((ok) => setStatus(ok ? "ok" : "unauth"))
  }, [])

  if (status === "loading") return null
  if (status === "unauth") return <Navigate to="/login" replace />
  return <>{children}</>
}
