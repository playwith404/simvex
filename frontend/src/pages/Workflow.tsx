import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import ReactFlow, {
  addEdge,
  Background,
  Controls,
  MarkerType,
  useEdgesState,
  useNodesState,
} from 'reactflow'
import type { Connection, Edge, Node } from 'reactflow'
import 'reactflow/dist/style.css'
import { EditableNode } from '../components/ui/EditableNode'
import {
  createProject,
  deleteProject,
  fetchObjects,
  fetchParts,
  getWorkflowFull,
  getPartNote,
  listProjects,
  notionConnect,
  notionDisconnect,
  notionStatus,
  notionSync,
  saveWorkflowFull,
  savePartNote,
} from '../services/api'
import type {
  Part,
  WorkflowAttachment,
  WorkflowChecklist,
  WorkflowNode,
  WorkflowProject,
} from '../types'

const emptyNodes: Node[] = []
const emptyEdges: Edge[] = []

export const Workflow = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState(emptyNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(emptyEdges)
  const [projects, setProjects] = useState<WorkflowProject[]>([])
  const [activeProjectId, setActiveProjectId] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [isHydrating, setIsHydrating] = useState(false)
  const [hasLoadedProject, setHasLoadedProject] = useState(false)
  const [saveStatus, setSaveStatus] = useState<string | null>(null)
  const [notionConnected, setNotionConnected] = useState(false)
  const [notionToken, setNotionToken] = useState('')
  const [notionParentPageId, setNotionParentPageId] = useState('')
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [parts, setParts] = useState<Part[]>([])
  const [noteContent, setNoteContent] = useState('')
  const [noteLoading, setNoteLoading] = useState(false)
  const saveTimer = useRef<number | null>(null)

  const handleSelectNode = useCallback((id: string) => {
    setSelectedNodeId(id)
  }, [])

  const makeId = () => {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
      return crypto.randomUUID()
    }
    return `id-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  }

  const handleNodeLabelChange = useCallback(
    (id: string, value: string) => {
      setNodes((nds) =>
        nds.map((node) =>
          node.id === id ? { ...node, data: { ...node.data, label: value } } : node,
        ),
      )
    },
    [setNodes],
  )

  const handleAddAttachment = useCallback(
    (id: string, attachment: { id?: string; type: 'file' | 'link'; name: string; url: string }) => {
      setNodes((nds) =>
        nds.map((node) =>
          node.id === id
            ? {
                ...node,
                data: {
                  ...node.data,
                  attachments: [...(node.data.attachments || []), attachment],
                },
              }
            : node,
        ),
      )
    },
    [setNodes],
  )

  const handleRemoveAttachment = useCallback(
    (id: string, index: number) => {
      setNodes((nds) =>
        nds.map((node) =>
          node.id === id
            ? {
                ...node,
                data: {
                  ...node.data,
                  attachments: (node.data.attachments || []).filter((_: unknown, i: number) => i !== index),
                },
              }
            : node,
        ),
      )
    },
    [setNodes],
  )

  useEffect(() => {
    let alive = true
    const load = async () => {
      try {
        const [projectList, notionState] = await Promise.all([
          listProjects(),
          notionStatus().catch(() => ({ connected: false })),
        ])
        if (!alive) return
        setProjects(Array.isArray(projectList) ? projectList : [])
        setNotionConnected(notionState.connected)
        if (Array.isArray(projectList) && projectList.length > 0) {
          setActiveProjectId(projectList[0].id)
        }
      } catch {
        if (!alive) return
        setProjects([])
        setActiveProjectId('')
      } finally {
        if (alive) setLoading(false)
      }
    }
    load()
    return () => {
      alive = false
    }
  }, [])

  useEffect(() => {
    let alive = true
    const loadParts = async () => {
      try {
        const objects = await fetchObjects()
        const allParts: Part[] = []
        for (const obj of objects) {
          const list = await fetchParts(obj.id)
          allParts.push(...list)
        }
        if (alive) setParts(allParts)
      } catch {
        if (alive) setParts([])
      }
    }
    loadParts()
    return () => {
      alive = false
    }
  }, [])

  useEffect(() => {
    if (!activeProjectId) return
    let alive = true
    const fetchFull = async () => {
      setHasLoadedProject(false)
      setIsHydrating(true)
      const full = await getWorkflowFull(activeProjectId)
      if (!alive) return
      const mappedNodes: Node[] = full.nodes.map((n) => ({
        id: n.id,
        position: { x: n.positionX, y: n.positionY },
        data: {
          label: n.title || '새 노드',
          title: n.title || '새 노드',
          description: n.description || '',
          scheduledDate: n.scheduledDate,
          progress: n.progress,
          color: n.color || '',
          linkedPartId: n.linkedPartId || '',
          linkedNoteId: n.linkedNoteId || '',
          notionPageId: n.notionPageId || '',
          attachments: full.attachments.filter((a) => a.nodeId === n.id),
          checklists: full.checklists.filter((c) => c.nodeId === n.id),
          onChange: handleNodeLabelChange,
          onAddAttachment: handleAddAttachment,
          onRemoveAttachment: handleRemoveAttachment,
          onSelect: handleSelectNode,
        },
        type: 'editable',
      }))
      const mappedEdges: Edge[] = full.edges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        markerEnd: { type: MarkerType.ArrowClosed, color: '#8fa3b8' },
        style: { stroke: '#8fa3b8' },
      }))
      setNodes(mappedNodes)
      setEdges(mappedEdges)
      setHasLoadedProject(true)
      setIsHydrating(false)
    }
    fetchFull().catch(() => {
      setNodes([])
      setEdges([])
      setSelectedNodeId(null)
      setHasLoadedProject(false)
      setIsHydrating(false)
    })
    return () => {
      alive = false
    }
  }, [activeProjectId, handleAddAttachment, handleNodeLabelChange, handleRemoveAttachment, setEdges, setNodes])

  useEffect(() => {
    if (!selectedNodeId) return
    const exists = nodes.some((n) => n.id === selectedNodeId)
    if (!exists) setSelectedNodeId(null)
  }, [nodes, selectedNodeId])

  const onConnect = useCallback(
    (connection: Connection) =>
      setEdges((eds) =>
        addEdge(
          {
            ...connection,
            id: makeId(),
            markerEnd: { type: MarkerType.ArrowClosed, color: '#8fa3b8' },
            style: { stroke: '#8fa3b8' },
          },
          eds,
        ),
      ),
    [setEdges],
  )

  const addNode = () => {
    const id = makeId()
    const newNode: Node = {
      id,
      position: { x: 120 + nodes.length * 40, y: 120 + nodes.length * 30 },
      data: {
        label: '새 노드',
        color: '#f8c86a',
        attachments: [],
        onChange: handleNodeLabelChange,
        onAddAttachment: handleAddAttachment,
        onRemoveAttachment: handleRemoveAttachment,
        onSelect: handleSelectNode,
      },
      type: 'editable',
    }
    setNodes((nds) => [...nds, newNode])
    setSelectedNodeId(id)
  }

  const clearAll = () => {
    setNodes([])
    setEdges([])
  }

  const handleCreateProject = async () => {
    const title = window.prompt('프로젝트 이름을 입력하세요')
    if (!title) return
    const project = await createProject(title)
    setProjects((prev) => [project, ...prev])
    setActiveProjectId(project.id)
  }

  const handleDeleteProject = async () => {
    if (!activeProjectId) return
    const target = projects.find((p) => p.id === activeProjectId)
    const ok = window.confirm(`프로젝트 "${target?.title ?? ''}"를 삭제할까요?`)
    if (!ok) return
    await deleteProject(activeProjectId)
    const next = projects.filter((p) => p.id !== activeProjectId)
    setProjects(next)
    setActiveProjectId(next[0]?.id ?? '')
  }

  const handleConnectNotion = async () => {
    if (!notionToken.trim() || !notionParentPageId.trim()) return
    try {
      await notionConnect(notionToken.trim(), notionParentPageId.trim())
      setNotionToken('')
      setNotionParentPageId('')
      setNotionConnected(true)
    } catch {
      setNotionConnected(false)
    }
  }

  const handleDisconnectNotion = async () => {
    await notionDisconnect()
    setNotionConnected(false)
  }

  const handleSyncNotion = async () => {
    try {
      setSaveStatus('Notion 동기화 중...')
      await notionSync()
      setSaveStatus('Notion 동기화 완료')
    } catch {
      setSaveStatus('Notion 동기화 실패')
    }
  }

  const updateSelectedNode = (partial: Record<string, unknown>) => {
    if (!selectedNodeId) return
    setNodes((nds) =>
      nds.map((node) =>
        node.id === selectedNodeId
          ? {
              ...node,
              data: {
                ...node.data,
                ...partial,
              },
            }
          : node,
      ),
    )
  }

  const selectedNode = nodes.find((n) => n.id === selectedNodeId) || null

  useEffect(() => {
    const partId = selectedNode?.data?.linkedPartId as string | undefined
    if (!partId) {
      setNoteContent('')
      return
    }
    let alive = true
    setNoteLoading(true)
    getPartNote(partId)
      .then((res) => {
        if (alive) setNoteContent(res?.content ?? '')
      })
      .finally(() => {
        if (alive) setNoteLoading(false)
      })
    return () => {
      alive = false
    }
  }, [selectedNode?.data?.linkedPartId])

  const buildPayload = () => {
    const today = new Date().toISOString().slice(0, 10)
    const payloadNodes: WorkflowNode[] = nodes.map((n) => ({
      id: n.id,
      projectId: activeProjectId,
      title: String(n.data?.title ?? n.data?.label ?? '새 노드'),
      description: String(n.data?.description ?? ''),
      scheduledDate: String(n.data?.scheduledDate ?? today),
      progress: Number(n.data?.progress ?? 0),
      color: String(n.data?.color ?? ''),
      positionX: n.position.x,
      positionY: n.position.y,
      linkedPartId: String(n.data?.linkedPartId ?? ''),
      linkedNoteId: String(n.data?.linkedNoteId ?? ''),
      notionPageId: String(n.data?.notionPageId ?? ''),
    }))
    const payloadEdges = edges.map((e) => ({
      id: e.id,
      projectId: activeProjectId,
      source: e.source,
      target: e.target,
    }))
    const attachments: WorkflowAttachment[] = []
    nodes.forEach((n) => {
      const list = (n.data?.attachments || []) as WorkflowAttachment[]
      list.forEach((a) => {
        const id = a.id || makeId()
        attachments.push({ ...a, id, nodeId: n.id })
      })
    })
    const checklists: WorkflowChecklist[] = []
    nodes.forEach((n) => {
      const list = (n.data?.checklists || []) as WorkflowChecklist[]
      list.forEach((c) => {
        const id = c.id || makeId()
        checklists.push({ ...c, id, nodeId: n.id })
      })
    })
    return { nodes: payloadNodes, edges: payloadEdges, checklists, attachments }
  }

  useEffect(() => {
    if (!activeProjectId) return
    if (!hasLoadedProject || isHydrating) return
    if (saveTimer.current) {
      window.clearTimeout(saveTimer.current)
    }
    saveTimer.current = window.setTimeout(async () => {
      const payload = buildPayload()
      try {
        setSaveStatus('저장 중...')
        await saveWorkflowFull(activeProjectId, payload)
        setSaveStatus('저장됨')
      } catch {
        setSaveStatus('저장 실패')
      }
    }, 800)
    return () => {
      if (saveTimer.current) {
        window.clearTimeout(saveTimer.current)
      }
    }
  }, [activeProjectId, edges, nodes])

  const handleSaveNow = async () => {
    if (!activeProjectId) return
    try {
      setSaveStatus('저장 중...')
      await saveWorkflowFull(activeProjectId, buildPayload())
      setSaveStatus('저장됨')
    } catch {
      setSaveStatus('저장 실패')
    }
  }

  const headerButtons = useMemo(
    () => (
      <div className="workflow-actions">
        <button type="button" onClick={addNode}>노드 추가</button>
        <button type="button" onClick={clearAll} className="ghost">전체 삭제</button>
      </div>
    ),
    [nodes.length],
  )

  return (
    <section className="workflow-page">
      <div className="workflow-header">
        <div>
          <h2>워크플로우 차트</h2>
          <p>학습 단계를 노드로 구성하고 연결하세요.</p>
          {saveStatus && <p className="muted">{saveStatus}</p>}
        </div>
        <div className="workflow-actions">
          <div className="workflow-projects">
            <select
              value={activeProjectId}
              onChange={(e) => setActiveProjectId(e.target.value)}
              disabled={loading || projects.length === 0}
            >
              {projects.length === 0 && <option value="">프로젝트 없음</option>}
              {projects.map((p) => (
                <option key={p.id} value={p.id}>{p.title}</option>
              ))}
            </select>
            <button type="button" className="ghost" onClick={handleCreateProject}>프로젝트 추가</button>
            <button type="button" className="ghost" onClick={handleDeleteProject} disabled={!activeProjectId}>프로젝트 삭제</button>
          </div>
          <div className="workflow-notion">
            {notionConnected ? (
              <>
                <button type="button" onClick={handleSyncNotion}>Notion 동기화</button>
                <button type="button" className="ghost" onClick={handleDisconnectNotion}>연결 해제</button>
              </>
            ) : (
              <>
                <input
                  placeholder="Notion 토큰 입력"
                  value={notionToken}
                  onChange={(e) => setNotionToken(e.target.value)}
                  type="password"
                />
                <input
                  placeholder="Parent Page ID"
                  value={notionParentPageId}
                  onChange={(e) => setNotionParentPageId(e.target.value)}
                />
                <button type="button" onClick={handleConnectNotion} disabled={!notionToken.trim() || !notionParentPageId.trim()}>연결</button>
                <button
                  type="button"
                  className="ghost"
                  onClick={() => {
                    window.alert(
                      'Notion 연결 방법:\n' +
                        '1) Notion → Settings & members → Integrations\n' +
                        '2) New integration 생성\n' +
                        '3) Internal Integration Token 복사\n' +
                        '4) 동기화할 페이지를 만들고 Integration을 연결\n' +
                        '5) 페이지 URL에서 Page ID 복사 (notion.so/이름-{32자리ID})\n',
                    )
                  }}
                >
                  연결 방법
                </button>
              </>
            )}
          </div>
          {headerButtons}
          <button type="button" onClick={handleSaveNow}>저장</button>
        </div>
      </div>
      <div className="workflow-body">
        <div className="workflow-canvas">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={(_, node) => setSelectedNodeId(node.id)}
            fitView
            nodeTypes={{ editable: EditableNode }}
          >
            <Background gap={16} color="#24333b" />
            <Controls />
          </ReactFlow>
        </div>
        <aside className="workflow-detail">
          {selectedNode ? (
            <div className="workflow-detail__card">
              <h3>노드 상세</h3>
              <label>
                제목
                <input
                  value={String(selectedNode.data?.title ?? selectedNode.data?.label ?? '')}
                  onChange={(e) =>
                    updateSelectedNode({ title: e.target.value, label: e.target.value })
                  }
                />
              </label>
              <label>
                날짜
                <input
                  type="date"
                  value={String(selectedNode.data?.scheduledDate ?? '')}
                  onChange={(e) => updateSelectedNode({ scheduledDate: e.target.value })}
                />
              </label>
              <label>
                진행률
                <input
                  type="range"
                  min={0}
                  max={100}
                  value={Number(selectedNode.data?.progress ?? 0)}
                  onChange={(e) => updateSelectedNode({ progress: Number(e.target.value) })}
                />
              </label>
              <label>
                색상
                <input
                  type="color"
                  value={String(selectedNode.data?.color || '#f8c86a')}
                  onChange={(e) => updateSelectedNode({ color: e.target.value })}
                />
              </label>
              <label>
                설명
                <textarea
                  value={String(selectedNode.data?.description ?? '')}
                  onChange={(e) => updateSelectedNode({ description: e.target.value })}
                />
              </label>
              <label>
                부품 연결
                <select
                  value={String(selectedNode.data?.linkedPartId ?? '')}
                  onChange={(e) => updateSelectedNode({ linkedPartId: e.target.value })}
                >
                  <option value="">선택 없음</option>
                  {parts.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </label>
              <div className="workflow-note">
                <div className="workflow-checklist__header">
                  <span>부품 노트</span>
                  <button
                    type="button"
                    className="ghost"
                    disabled={!selectedNode.data?.linkedPartId || noteLoading}
                    onClick={async () => {
                      const partId = String(selectedNode.data?.linkedPartId || '')
                      if (!partId) return
                      setNoteLoading(true)
                      try {
                        await savePartNote(partId, noteContent)
                      } finally {
                        setNoteLoading(false)
                      }
                    }}
                  >
                    저장
                  </button>
                </div>
                <textarea
                  placeholder={selectedNode.data?.linkedPartId ? '부품 노트를 입력하세요' : '부품을 선택하면 노트를 작성할 수 있습니다.'}
                  value={noteContent}
                  onChange={(e) => setNoteContent(e.target.value)}
                  disabled={!selectedNode.data?.linkedPartId || noteLoading}
                />
              </div>
              <div className="workflow-checklist">
                <div className="workflow-checklist__header">
                  <span>체크리스트</span>
                  <button
                    type="button"
                    className="ghost"
                    onClick={() => {
                      const list = (selectedNode.data?.checklists || []) as WorkflowChecklist[]
                      updateSelectedNode({
                        checklists: [...list, { id: makeId(), nodeId: selectedNode.id, text: '새 항목', done: false }],
                      })
                    }}
                  >
                    항목 추가
                  </button>
                </div>
                {(selectedNode.data?.checklists || []).map((item: WorkflowChecklist, idx: number) => (
                  <div key={item.id} className="workflow-checklist__item">
                    <input
                      type="checkbox"
                      checked={item.done}
                      onChange={(e) => {
                        const list = [...(selectedNode.data?.checklists || [])] as WorkflowChecklist[]
                        list[idx] = { ...list[idx], done: e.target.checked }
                        updateSelectedNode({ checklists: list })
                      }}
                    />
                    <input
                      value={item.text}
                      onChange={(e) => {
                        const list = [...(selectedNode.data?.checklists || [])] as WorkflowChecklist[]
                        list[idx] = { ...list[idx], text: e.target.value }
                        updateSelectedNode({ checklists: list })
                      }}
                    />
                    <button
                      type="button"
                      onClick={() => {
                        const list = [...(selectedNode.data?.checklists || [])] as WorkflowChecklist[]
                        list.splice(idx, 1)
                        updateSelectedNode({ checklists: list })
                      }}
                    >
                      삭제
                    </button>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="workflow-detail__empty">
              노드를 선택하면 상세 정보가 표시됩니다.
            </div>
          )}
        </aside>
      </div>
    </section>
  )
}
