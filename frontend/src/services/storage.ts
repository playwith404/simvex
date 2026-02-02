import type { StoredData } from '../types'

const STORAGE_PREFIX = 'simvex:'

export const loadState = (objectId: string): StoredData | null => {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}${objectId}`)
    if (!raw) return null
    return JSON.parse(raw) as StoredData
  } catch {
    return null
  }
}

export const saveState = (objectId: string, data: StoredData) => {
  try {
    localStorage.setItem(`${STORAGE_PREFIX}${objectId}`, JSON.stringify(data))
  } catch {
    // ignore storage errors
  }
}
