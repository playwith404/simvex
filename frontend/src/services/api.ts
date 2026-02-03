import type { ObjectModel, Part, ChatMessage } from '../types'

const API_BASE = '/api'

const handle = async <T>(res: Response): Promise<T> => {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body?.message || body?.error?.message || '요청에 실패했습니다')
  }
  return res.json() as Promise<T>
}

const fetchJson = (input: RequestInfo, init?: RequestInit) =>
  fetch(input, {
    credentials: 'include',
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  })

export const fetchObjects = async (): Promise<ObjectModel[]> => {
  const res = await fetchJson(`${API_BASE}/objects`, { method: 'GET' })
  return handle<ObjectModel[]>(res)
}

export const fetchObject = async (id: string): Promise<ObjectModel> => {
  const res = await fetchJson(`${API_BASE}/objects/${id}`, { method: 'GET' })
  return handle<ObjectModel>(res)
}

export const fetchParts = async (id: string): Promise<Part[]> => {
  const res = await fetchJson(`${API_BASE}/objects/${id}/parts`, { method: 'GET' })
  return handle<Part[]>(res)
}

export const fetchPart = async (partId: string): Promise<Part> => {
  const res = await fetchJson(`${API_BASE}/parts/${partId}`, { method: 'GET' })
  return handle<Part>(res)
}

export const sendChat = async (payload: {
  objectId: string
  partId?: string
  userMessage: string
  history: ChatMessage[]
}): Promise<{ assistantMessage: string }> => {
  const res = await fetchJson(`${API_BASE}/ai/chat`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ assistantMessage: string }>(res)
}

export const register = async (payload: { email: string; password: string }) => {
  const res = await fetchJson(`${API_BASE}/auth/register`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ message: string }>(res)
}

export const verifyEmail = async (payload: { email: string; code: string }) => {
  const res = await fetchJson(`${API_BASE}/auth/verify`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ message: string }>(res)
}

export const login = async (payload: { email: string; password: string }) => {
  const res = await fetchJson(`${API_BASE}/auth/login`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ message: string }>(res)
}

export const logout = async () => {
  const res = await fetchJson(`${API_BASE}/auth/logout`, { method: 'POST' })
  return handle<{ message: string }>(res)
}

export const requestPasswordReset = async (payload: { email: string }) => {
  const res = await fetchJson(`${API_BASE}/auth/password/reset-request`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ message: string }>(res)
}

export const confirmPasswordReset = async (payload: {
  email: string
  code: string
  newPassword: string
}) => {
  const res = await fetchJson(`${API_BASE}/auth/password/reset-confirm`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return handle<{ message: string }>(res)
}
