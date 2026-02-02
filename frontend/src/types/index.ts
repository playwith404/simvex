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
