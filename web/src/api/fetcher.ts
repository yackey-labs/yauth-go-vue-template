// Custom fetch wrapper that the orval-generated client routes every
// call through. The two-argument shape (`apiFetch(url, init)`) matches
// what orval emits for `client: 'fetch'` with a mutator override.
//
// Edit this when you need to add custom headers (e.g. CSRF tokens),
// retry logic, or telemetry. The credentials: 'include' bit is what
// keeps yauth's session cookie on every request.

export class ApiError extends Error {
  status: number
  body: unknown

  constructor(message: string, status: number, body: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

export async function apiFetch<T>(
  url: string,
  init: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    ...(init.headers as Record<string, string> | undefined),
  }
  if (init.body !== undefined && !(init.body instanceof FormData)) {
    headers['Content-Type'] ??= 'application/json'
  }

  const res = await fetch(url, {
    credentials: 'include',
    ...init,
    headers,
  })

  const text = await res.text()
  let parsed: unknown = undefined
  if (text) {
    try {
      parsed = JSON.parse(text)
    } catch {
      parsed = text
    }
  }

  if (!res.ok) {
    const message =
      typeof parsed === 'object' && parsed !== null && 'detail' in parsed
        ? String((parsed as { detail: unknown }).detail)
        : `${res.status} ${res.statusText}`
    throw new ApiError(message, res.status, parsed)
  }

  return parsed as T
}
