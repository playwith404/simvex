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
  getWorkflowFull,
  listProjects,
  notionConnect,
  notionDisconnect,
  notionStatus,
  notionSync,
  saveWorkflowFull,
} from '../services/api'
import type {
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
  const [saveStatus, setSaveStatus] = useState<string | null>(null)
  const [notionConnected, setNotionConnected] = useState(false)
  const [notionToken, setNotionToken] = useState('')
  const saveTimer = useRef<number | null>(null)
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
    (id: string, attachment: { type: 'file' | 'link'; name: string; url: string }) => {
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
    if (!activeProjectId) return
    let alive = true
    const fetchFull = async () => {
      const full = await getWorkflowFull(activeProjectId)
      if (!alive) return
      const mappedNodes: Node[] = full.nodes.map((n) => ({
        id: n.id,
        position: { x: n.positionX, y: n.positionY },
        data: {
          label: n.title || '새 노드',
          attachments: full.attachments.filter((a) => a.nodeId === n.id),
          onChange: handleNodeLabelChange,
          onAddAttachment: handleAddAttachment,
          onRemoveAttachment: handleRemoveAttachment,
        },
        type: 'editable',
      }))
      const mappedEdges: Edge[] = full.edges.map((e) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        markerEnd: { type: MarkerType.ArrowClosed },
      }))
      setNodes(mappedNodes)
      setEdges(mappedEdges)
    }
    fetchFull().catch(() => {
      setNodes([])
      setEdges([])
    })
    return () => {
      alive = false
    }
  }, [activeProjectId, handleAddAttachment, handleNodeLabelChange, handleRemoveAttachment, setEdges, setNodes])

  const onConnect = useCallback(
    (connection: Connection) =>
      setEdges((eds) =>
        addEdge(
          {
            ...connection,
            id: makeId(),
            markerEnd: { type: MarkerType.ArrowClosed },
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
        attachments: [],
        onChange: handleNodeLabelChange,
        onAddAttachment: handleAddAttachment,
        onRemoveAttachment: handleRemoveAttachment,
      },
      type: 'editable',
    }
    setNodes((nds) => [...nds, newNode])
  }

  const clearAll = () => {
    setNodes([])
    setEdges([])
  }

  const handleCreateProject = async () => {
    const title = window.prompt('새 프로젝트 제목을 입력하세요')
    if (!title) return
    const project = await createProject(title)
    setProjects((prev) => [project, ...prev])
    setActiveProjectId(project.id)
  }

  const handleConnectNotion = async () => {
    if (!notionToken.trim()) return
    try {
      await notionConnect(notionToken.trim())
      setNotionToken('')
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
    await notionSync()
  }

  useEffect(() => {
    if (!activeProjectId) return
    if (saveTimer.current) {
      window.clearTimeout(saveTimer.current)
    }
    saveTimer.current = window.setTimeout(async () => {
      const today = new Date().toISOString().slice(0, 10)
      const payloadNodes: WorkflowNode[] = nodes.map((n) => ({
        id: n.id,
        projectId: activeProjectId,
        title: String(n.data?.label ?? '새 노드'),
        scheduledDate: today,
        progress: 0,
        positionX: n.position.x,
        positionY: n.position.y,
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
      try {
        setSaveStatus('저장 중...')
        await saveWorkflowFull(activeProjectId, {
          nodes: payloadNodes,
          edges: payloadEdges,
          checklists,
          attachments,
        })
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
                <button type="button" onClick={handleConnectNotion}>연결</button>
              </>
            )}
          </div>
          {headerButtons}
        </div>
      </div>
      <div className="workflow-canvas">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
          nodeTypes={{ editable: EditableNode }}
        >
          <Background gap={16} color="#24333b" />
          <Controls />
        </ReactFlow>
      </div>
    </section>
  )
}
