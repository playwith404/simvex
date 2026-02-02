import type { ObjectModel, Part, ChatMessage } from '../types'

const API_BASE = '/api'

const handle = async <T>(res: Response): Promise<T> => {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body?.error?.message || '요청에 실패했습니다')
  }
  return res.json() as Promise<T>
}

export const fetchObjects = async (): Promise<ObjectModel[]> => {
  const res = await fetch(`${API_BASE}/objects`)
  return handle<ObjectModel[]>(res)
}

export const fetchObject = async (id: string): Promise<ObjectModel> => {
  const res = await fetch(`${API_BASE}/objects/${id}`)
  return handle<ObjectModel>(res)
}

export const fetchParts = async (id: string): Promise<Part[]> => {
  const res = await fetch(`${API_BASE}/objects/${id}/parts`)
  return handle<Part[]>(res)
}

export const fetchPart = async (partId: string): Promise<Part> => {
  const res = await fetch(`${API_BASE}/parts/${partId}`)
  return handle<Part>(res)
}

export const sendChat = async (payload: {
  objectId: string
  partId?: string
  userMessage: string
  history: ChatMessage[]
}): Promise<{ assistantMessage: string }> => {
  const res = await fetch(`${API_BASE}/ai/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  return handle<{ assistantMessage: string }>(res)
}
