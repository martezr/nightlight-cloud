export async function checkAuth(): Promise<boolean> {
  try {
    const res = await fetch("/api/v1/auth/me")
    return res.ok
  } catch {
    return false
  }
}

export async function login(username: string, password: string): Promise<void> {
  const res = await fetch("/api/v1/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text.trim() || "Login failed")
  }
}

export async function logout(): Promise<void> {
  await fetch("/api/v1/auth/logout", { method: "POST" })
}
