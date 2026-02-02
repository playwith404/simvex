import { useCallback, useEffect, useMemo } from 'react'
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

const STORAGE_KEY = 'simvex:workflow'

const defaultNodes: Node[] = [
  {
    id: '1',
    position: { x: 80, y: 80 },
    data: { label: '학습 목표\n- 구조 이해하기', attachments: [] },
    type: 'editable',
  },
  {
    id: '2',
    position: { x: 380, y: 200 },
    data: { label: '부품 분석\n- 핵심 부품 선택', attachments: [] },
    type: 'editable',
  },
  {
    id: '3',
    position: { x: 680, y: 320 },
    data: { label: '퀴즈 정리\n- 노트 정리', attachments: [] },
    type: 'editable',
  },
]

const defaultEdges: Edge[] = [
  {
    id: 'e1-2',
    source: '1',
    target: '2',
    markerEnd: { type: MarkerType.ArrowClosed },
  },
  {
    id: 'e2-3',
    source: '2',
    target: '3',
    markerEnd: { type: MarkerType.ArrowClosed },
  },
]

export const Workflow = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

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
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      const parsed = JSON.parse(saved) as { nodes: Node[]; edges: Edge[] }
      setNodes(
        parsed.nodes.map((node) => ({
          ...node,
          type: 'editable',
          data: {
            ...node.data,
            attachments: node.data.attachments || [],
            onChange: handleNodeLabelChange,
            onAddAttachment: handleAddAttachment,
            onRemoveAttachment: handleRemoveAttachment,
          },
        })),
      )
      setEdges(parsed.edges)
    } else {
      setNodes(
        defaultNodes.map((node) => ({
          ...node,
          data: {
            ...node.data,
            onChange: handleNodeLabelChange,
            onAddAttachment: handleAddAttachment,
            onRemoveAttachment: handleRemoveAttachment,
          },
        })),
      )
      setEdges(defaultEdges)
    }
  }, [setNodes, setEdges, handleNodeLabelChange, handleAddAttachment, handleRemoveAttachment])

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ nodes, edges }))
  }, [nodes, edges])

  const onConnect = useCallback(
    (connection: Connection) => setEdges((eds) => addEdge(connection, eds)),
    [setEdges],
  )

  const addNode = () => {
    const id = `node-${nodes.length + 1}`
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
        </div>
        {headerButtons}
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
