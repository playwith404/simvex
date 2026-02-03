export type ObjectModel = {
  id: string
  name: string
  description: string
  thumbnail: string
  modelPath: string
  theory: string
  category: string
  createdAt: string
}

export type Part = {
  id: string
  objectId: string
  name: string
  material: string
  role: string
  modelPath: string
  localPosX: number
  localPosY: number
  localPosZ: number
  decomposeDirX: number
  decomposeDirY: number
  decomposeDirZ: number
  decomposeDistance: number
}

export type ChatMessage = {
  role: 'user' | 'assistant'
  content: string
}

export type ViewState = {
  cameraPosition: { x: number; y: number; z: number }
  cameraTarget: { x: number; y: number; z: number }
  zoom: number
  decompositionLevel: number
}

export type StoredData = {
  objectId: string
  viewState: ViewState
  selectedPart?: string
  notes: string
  aiHistory: ChatMessage[]
  lastUpdated: string
}

export type WorkflowProject = {
  id: string
  userId: string
  title: string
  notionPageId?: string
  createdAt: string
  updatedAt: string
}

export type WorkflowNode = {
  id: string
  projectId: string
  title: string
  description?: string
  scheduledDate: string
  progress: number
  color?: string
  positionX: number
  positionY: number
  linkedPartId?: string
  linkedNoteId?: string
  notionPageId?: string
}

export type WorkflowEdge = {
  id: string
  projectId: string
  source: string
  target: string
}

export type WorkflowChecklist = {
  id: string
  nodeId: string
  text: string
  done: boolean
}

export type WorkflowAttachment = {
  id: string
  nodeId: string
  type: 'link' | 'file'
  name: string
  url: string
}

export type WorkflowFull = {
  project: WorkflowProject
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  checklists: WorkflowChecklist[]
  attachments: WorkflowAttachment[]
}
