import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useGLTF } from '@react-three/drei'
import { ThreeScene } from '../components/viewer/ThreeScene'
import { SidePanel } from '../components/ui/SidePanel'
import { DecomposeSlider } from '../components/ui/DecomposeSlider'
import { ClipControls } from '../components/ui/ClipControls'
import {
  fetchObject,
  fetchParts,
  fetchObjectVersions,
  getPartNote,
  savePartNote,
  sendChat,
} from '../services/api'
import { loadState, saveState } from '../services/storage'
import { usePdfExport } from '../hooks/usePdfExport'
import type { PdfImageData } from '../types/pdf'
import type {
  ChatMessage,
  ObjectModel,
  Part,
  StoredData,
  Measurement,
  ViewerMode,
} from '../types'

let measureIdCounter = 0

export const Viewer = () => {
  const { objectId } = useParams<{ objectId: string }>()
  const [object, setObject] = useState<ObjectModel | null>(null)
  const [parts, setParts] = useState<Part[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedPartId, setSelectedPartId] = useState<string | null>(null)
  const [hoveredPartId, setHoveredPartId] = useState<string | null>(null)
  const [decompositionLevel, setDecompositionLevel] = useState(0)
  const [notes, setNotes] = useState('')
  const [noteSaving, setNoteSaving] = useState(false)
  const [aiHistory, setAiHistory] = useState<ChatMessage[]>([])
  const [activeTab, setActiveTab] = useState<'note' | 'ai' | 'compare'>('note')
  const [viewState, setViewState] = useState<StoredData['viewState'] | undefined>(undefined)
  const [aiLoading, setAiLoading] = useState(false)
  const { exportPdf, isExporting } = usePdfExport()
  const [canvasEl, setCanvasEl] = useState<HTMLCanvasElement | null>(null)
  const [captureImage, setCaptureImage] = useState<(() => PdfImageData | null) | null>(null)

  // #1 파트 가시성
  const [hiddenPartIds, setHiddenPartIds] = useState<Set<string>>(new Set())

  // #5 측정
  const [mode, setMode] = useState<ViewerMode>('select')
  const [measurements, setMeasurements] = useState<Measurement[]>([])
  const [pendingMeasurePoint, setPendingMeasurePoint] = useState<{
    x: number
    y: number
    z: number
  } | null>(null)

  // #4 단면 절단
  const [clipEnabled, setClipEnabled] = useState(false)
  const [clipAxis, setClipAxis] = useState<'x' | 'y' | 'z'>('x')
  const [clipPosition, setClipPosition] = useState(0)

  // #7 파트 비교
  const [compareIds, setCompareIds] = useState<string[]>([])

  // #8 버전 관리
  const [versions, setVersions] = useState<ObjectModel[]>([])

  // #6 모델 프리로딩
  useEffect(() => {
    if (parts.length === 0) return
    parts.forEach((part) => {
      useGLTF.preload(part.modelPath)
    })
  }, [parts])

  useEffect(() => {
    if (!objectId) return
    setLoading(true)
    setNotes('')
    Promise.all([fetchObject(objectId), fetchParts(objectId)])
      .then(([obj, partsData]) => {
        setObject(obj)
        setParts(partsData)
        const saved = loadState(objectId)
        if (saved) {
          setSelectedPartId(saved.selectedPart ?? null)
          setDecompositionLevel(saved.viewState?.decompositionLevel ?? 0)
          setAiHistory(saved.aiHistory ?? [])
          setViewState(saved.viewState)
        }
        fetchObjectVersions(objectId)
          .then((v) => setVersions(v))
          .catch(() => setVersions([]))
        return getPartNote(objectId).catch(() => null)
      })
      .then((note) => {
        setNotes(note?.content ?? '')
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [objectId])

  useEffect(() => {
    if (!objectId || !object) return
    const payload: StoredData = {
      objectId,
      viewState: {
        cameraPosition: viewState?.cameraPosition ?? { x: 4, y: 4, z: 8 },
        cameraTarget: viewState?.cameraTarget ?? { x: 0, y: 0, z: 0 },
        zoom: viewState?.zoom ?? 1,
        decompositionLevel,
      },
      selectedPart: selectedPartId ?? undefined,
      notes,
      aiHistory,
      lastUpdated: new Date().toISOString(),
    }
    saveState(objectId, payload)
  }, [objectId, object, viewState, decompositionLevel, selectedPartId, notes, aiHistory])

  const selectedPart = useMemo(
    () => parts.find((part) => part.id === selectedPartId) || null,
    [parts, selectedPartId],
  )

  // #1 가시성 토글 핸들러
  const handleToggleVisibility = useCallback((partId: string) => {
    setHiddenPartIds((prev) => {
      const next = new Set(prev)
      if (next.has(partId)) {
        next.delete(partId)
      } else {
        next.add(partId)
      }
      return next
    })
  }, [])

  const handleToggleAllVisibility = useCallback(
    (visible: boolean) => {
      if (visible) {
        setHiddenPartIds(new Set())
      } else {
        setHiddenPartIds(new Set(parts.map((p) => p.id)))
      }
    },
    [parts],
  )

  // #5 측정 핸들러
  const handleMeasurePoint = useCallback(
    (point: { x: number; y: number; z: number }) => {
      if (!pendingMeasurePoint) {
        setPendingMeasurePoint(point)
      } else {
        const dx = point.x - pendingMeasurePoint.x
        const dy = point.y - pendingMeasurePoint.y
        const dz = point.z - pendingMeasurePoint.z
        const distance = Math.sqrt(dx * dx + dy * dy + dz * dz)
        const measurement: Measurement = {
          id: String(++measureIdCounter),
          start: pendingMeasurePoint,
          end: point,
          distance,
        }
        setMeasurements((prev) => [...prev, measurement])
        setPendingMeasurePoint(null)
      }
    },
    [pendingMeasurePoint],
  )

  const handleClearMeasurements = useCallback(() => {
    setMeasurements([])
    setPendingMeasurePoint(null)
  }, [])

  // #7 비교 핸들러
  const handleAddCompare = useCallback((partId: string) => {
    setCompareIds((prev) => {
      if (prev.includes(partId)) return prev
      if (prev.length >= 2) return [prev[1], partId]
      return [...prev, partId]
    })
  }, [])

  const handleRemoveCompare = useCallback((partId: string) => {
    setCompareIds((prev) => prev.filter((id) => id !== partId))
  }, [])

  const handleClearCompare = useCallback(() => {
    setCompareIds([])
  }, [])

  // 파트 선택 시 비교 탭이면 비교에 추가
  const handleSelectPart = useCallback(
    (partId: string) => {
      setSelectedPartId(partId)
      if (activeTab === 'compare') {
        handleAddCompare(partId)
      }
    },
    [activeTab, handleAddCompare],
  )

  const handleSendMessage = async (message: string) => {
    if (!objectId) return
    const nextHistory: ChatMessage[] = [
      ...aiHistory,
      { role: 'user' as const, content: message },
    ]
    setAiHistory(nextHistory)
    setAiLoading(true)
    try {
      const res = await sendChat({
        objectId,
        partId: selectedPartId ?? undefined,
        userMessage: message,
        history: aiHistory,
      })
      setAiHistory([
        ...nextHistory,
        { role: 'assistant' as const, content: res.assistantMessage },
      ])
    } catch {
      setAiHistory([
        ...nextHistory,
        { role: 'assistant' as const, content: 'AI 응답에 실패했습니다.' },
      ])
    } finally {
      setAiLoading(false)
    }
  }

  const handleModeChange = (newMode: ViewerMode) => {
    setMode(newMode)
    if (newMode !== 'measure') {
      setPendingMeasurePoint(null)
    }
  }

  if (loading) {
    return <div className="viewer-page">로딩 중...</div>
  }

  if (error || !object) {
    return <div className="viewer-page">{error || '오브젝트를 찾을 수 없습니다.'}</div>
  }

  return (
    <div className="viewer-page">
      <div className="viewer-header">
        <div className="viewer-header__left">
          <Link to="/objects" className="ghost" reloadDocument>
            목록
          </Link>
          <h2>{object.name}</h2>
          {versions.length > 1 && (
            <select
              className="viewer-version-select"
              value={objectId}
              onChange={(e) => {
                window.location.href = `/viewer/${e.target.value}`
              }}
            >
              {versions.map((v) => (
                <option key={v.id} value={v.id}>
                  v{v.version}
                </option>
              ))}
            </select>
          )}
          {versions.length <= 1 && object.version && (
            <span className="viewer-version-badge">v{object.version}</span>
          )}
        </div>
        <div className="viewer-header__actions">
          <button
            type="button"
            className="ghost"
            onClick={() =>
              exportPdf({
                canvas: canvasEl,
                getImageData: captureImage ?? undefined,
                objectName: object.name,
                notes,
                chatHistory: aiHistory,
                onSaved: () => {
                  window.focus()
                },
              })
            }
            disabled={isExporting}
          >
            PDF 저장
          </button>
          <Link to="/workflow" className="ghost" reloadDocument>
            워크플로우
          </Link>
        </div>
      </div>

      <div className="viewer-body">
        <div className="viewer-canvas">
          <ThreeScene
            parts={parts}
            decompositionLevel={decompositionLevel}
            selectedPartId={selectedPartId}
            hoveredPartId={hoveredPartId}
            onSelectPart={handleSelectPart}
            onHoverPart={setHoveredPartId}
            viewState={viewState}
            onViewStateChange={(state) => setViewState({ ...state, decompositionLevel })}
            onCanvasReady={setCanvasEl}
            onCaptureReady={(capture) => setCaptureImage(() => capture)}
            hiddenPartIds={hiddenPartIds}
            mode={mode}
            onMeasurePoint={handleMeasurePoint}
            measurements={measurements}
            pendingPoint={pendingMeasurePoint}
            clipEnabled={clipEnabled}
            clipAxis={clipAxis}
            clipPosition={clipPosition}
          />
          <div className="viewer-controls">
            <div className="viewer-toolbar">
              <div className="viewer-toolbar__modes">
                <button
                  type="button"
                  className={mode === 'select' ? 'active' : ''}
                  onClick={() => handleModeChange('select')}
                >
                  선택
                </button>
                <button
                  type="button"
                  className={mode === 'measure' ? 'active' : ''}
                  onClick={() => handleModeChange('measure')}
                >
                  측정
                </button>
              </div>
              {mode === 'measure' && measurements.length > 0 && (
                <button
                  type="button"
                  className="viewer-toolbar__clear"
                  onClick={handleClearMeasurements}
                >
                  측정 초기화
                </button>
              )}
            </div>
            <DecomposeSlider value={decompositionLevel} onChange={setDecompositionLevel} />
            <ClipControls
              enabled={clipEnabled}
              axis={clipAxis}
              position={clipPosition}
              onToggle={() => setClipEnabled(!clipEnabled)}
              onAxisChange={setClipAxis}
              onPositionChange={setClipPosition}
            />
          </div>
        </div>
        <SidePanel
          object={object}
          parts={parts}
          selectedPart={selectedPart}
          activeTab={activeTab}
          onTabChange={setActiveTab}
          notes={notes}
          onNotesChange={setNotes}
          onSaveNotes={async () => {
            if (!objectId) return
            setNoteSaving(true)
            try {
              await savePartNote(objectId, notes)
            } finally {
              setNoteSaving(false)
            }
          }}
          noteSaving={noteSaving}
          aiHistory={aiHistory}
          onSendMessage={handleSendMessage}
          aiLoading={aiLoading}
          hiddenPartIds={hiddenPartIds}
          onSelectPart={handleSelectPart}
          onToggleVisibility={handleToggleVisibility}
          onToggleAllVisibility={handleToggleAllVisibility}
          measurements={measurements}
          onClearMeasurements={handleClearMeasurements}
          compareIds={compareIds}
          onRemoveCompare={handleRemoveCompare}
          onClearCompare={handleClearCompare}
        />
      </div>
    </div>
  )
}
