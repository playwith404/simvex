import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ThreeScene } from '../components/viewer/ThreeScene'
import { SidePanel } from '../components/ui/SidePanel'
import { DecomposeSlider } from '../components/ui/DecomposeSlider'
import { fetchObject, fetchParts, sendChat } from '../services/api'
import { loadState, saveState } from '../services/storage'
import { usePdfExport } from '../hooks/usePdfExport'
import type { PdfImageData } from '../types/pdf'
import type { ChatMessage, ObjectModel, Part, StoredData } from '../types'

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
  const [aiHistory, setAiHistory] = useState<ChatMessage[]>([])
  const [activeTab, setActiveTab] = useState<'note' | 'ai'>('note')
  const [viewState, setViewState] = useState<StoredData['viewState'] | undefined>(undefined)
  const [aiLoading, setAiLoading] = useState(false)
  const { exportPdf, isExporting } = usePdfExport()
  const [canvasEl, setCanvasEl] = useState<HTMLCanvasElement | null>(null)
  const [captureImage, setCaptureImage] = useState<(() => PdfImageData | null) | null>(
    null,
  )

  useEffect(() => {
    if (!objectId) return
    setLoading(true)
    Promise.all([fetchObject(objectId), fetchParts(objectId)])
      .then(([obj, partsData]) => {
        setObject(obj)
        setParts(partsData)
        const saved = loadState(objectId)
        if (saved) {
          setSelectedPartId(saved.selectedPart ?? null)
          setDecompositionLevel(saved.viewState?.decompositionLevel ?? 0)
          setNotes(saved.notes)
          setAiHistory(saved.aiHistory ?? [])
          setViewState(saved.viewState)
        }
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
    } catch (err) {
      setAiHistory([
        ...nextHistory,
        { role: 'assistant' as const, content: 'AI 응답에 실패했습니다.' },
      ])
    } finally {
      setAiLoading(false)
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
          <Link to="/objects" className="ghost">← 목록</Link>
          <h2>{object.name}</h2>
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
          <Link to="/workflow" className="ghost">워크플로우</Link>
        </div>
      </div>

      <div className="viewer-body">
        <div className="viewer-canvas">
          <ThreeScene
            parts={parts}
            decompositionLevel={decompositionLevel}
            selectedPartId={selectedPartId}
            hoveredPartId={hoveredPartId}
            onSelectPart={setSelectedPartId}
            onHoverPart={setHoveredPartId}
            viewState={viewState}
            onViewStateChange={(state) => setViewState({ ...state, decompositionLevel })}
            onCanvasReady={setCanvasEl}
            onCaptureReady={(capture) => setCaptureImage(() => capture)}
          />
          <div className="viewer-controls">
            <DecomposeSlider value={decompositionLevel} onChange={setDecompositionLevel} />
          </div>
        </div>
        <SidePanel
          object={object}
          selectedPart={selectedPart}
          activeTab={activeTab}
          onTabChange={setActiveTab}
          notes={notes}
          onNotesChange={setNotes}
          aiHistory={aiHistory}
          onSendMessage={handleSendMessage}
          aiLoading={aiLoading}
        />
      </div>
    </div>
  )
}
