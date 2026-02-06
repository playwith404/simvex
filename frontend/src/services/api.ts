import type {
  ObjectModel,
  Part,
  ChatMessage,
  WorkflowProject,
  WorkflowFull,
  WorkflowNode,
  WorkflowEdge,
  WorkflowChecklist,
  WorkflowAttachment,
} from '../types'

const API_BASE = '/api'

const handle = async <T>(res: Response): Promise<T> => {
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body?.message || body?.error?.message || '요청에 실패했습니다')
  }
  return res.json() as Promise<T>
}

const fetchJson = (input: RequestInfo, init?: RequestInit) => {
  const headers = new Headers(init?.headers)
  if (init?.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  return fetch(input, {
    credentials: 'include',
    ...init,
    headers,
  })
}

const requestJson = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const res = await fetchJson(`${API_BASE}${path}`, init)
  return handle<T>(res)
}

export const fetchObjects = async (): Promise<ObjectModel[]> => requestJson<ObjectModel[]>('/objects')

export const fetchObject = async (id: string): Promise<ObjectModel> =>
  requestJson<ObjectModel>(`/objects/${id}`)

export const fetchParts = async (id: string): Promise<Part[]> =>
  requestJson<Part[]>(`/objects/${id}/parts`)

export const fetchObjectVersions = async (id: string): Promise<ObjectModel[]> =>
  requestJson<ObjectModel[]>(`/objects/${id}/versions`)

export const fetchPart = async (partId: string): Promise<Part> =>
  requestJson<Part>(`/parts/${partId}`)

export const sendChat = async (payload: {
  objectId: string
  partId?: string
  userMessage: string
  history: ChatMessage[]
}): Promise<{ assistantMessage: string }> => {
  return requestJson<{ assistantMessage: string }>('/ai/chat', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const register = async (payload: { email: string; password: string }) => {
  return requestJson<{ message: string }>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const verifyEmail = async (payload: { email: string; code: string }) => {
  return requestJson<{ message: string }>('/auth/verify', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const login = async (payload: { email: string; password: string }) => {
  return requestJson<{ message: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const logout = async () => {
  return requestJson<{ message: string }>('/auth/logout', { method: 'POST' })
}

export const requestPasswordReset = async (payload: { email: string }) => {
  return requestJson<{ message: string }>('/auth/password/reset-request', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const confirmPasswordReset = async (payload: {
  email: string
  code: string
  newPassword: string
}) => {
  return requestJson<{ message: string }>('/auth/password/reset-confirm', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export const listProjects = async (): Promise<WorkflowProject[]> =>
  requestJson<WorkflowProject[]>('/workflow/projects')

export const createProject = async (title: string): Promise<WorkflowProject> => {
  return requestJson<WorkflowProject>('/workflow/projects', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })
}

export const updateProject = async (id: string, title: string): Promise<WorkflowProject> => {
  return requestJson<WorkflowProject>(`/workflow/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ title }),
  })
}

export const deleteProject = async (id: string) => {
  return requestJson<{ message: string }>(`/workflow/projects/${id}`, { method: 'DELETE' })
}

export const getWorkflowFull = async (projectId: string): Promise<WorkflowFull> =>
  requestJson<WorkflowFull>(`/workflow/projects/${projectId}/full`)

export const saveWorkflowFull = async (
  projectId: string,
  payload: {
    nodes: WorkflowNode[]
    edges: WorkflowEdge[]
    checklists: WorkflowChecklist[]
    attachments: WorkflowAttachment[]
  },
) => {
  return requestJson<{ message: string }>(`/workflow/projects/${projectId}/full`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export const notionStatus = async (): Promise<{ connected: boolean }> => {
  return requestJson<{ connected: boolean }>('/notion/status')
}

export const notionConnect = async (token: string, parentPageId: string) => {
  return requestJson<{ message: string }>('/notion/connect', {
    method: 'POST',
    body: JSON.stringify({ token, parentPageId }),
  })
}

export const notionDisconnect = async () => {
  return requestJson<{ message: string }>('/notion/disconnect', { method: 'DELETE' })
}

export const notionSync = async () => {
  return requestJson<{ message: string }>('/notion/sync', { method: 'POST' })
}

export const getMe = async (): Promise<{ id: string; email: string } | null> => {
  try {
    return await requestJson<{ id: string; email: string }>('/auth/me')
  } catch {
    return null
  }
}

export const getPartNote = async (partId: string) => {
  return requestJson<{ content?: string }>(`/parts/${partId}/note`)
}

export const savePartNote = async (partId: string, content: string) => {
  return requestJson<{ id: string; content: string }>(`/parts/${partId}/note`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
  })
}
