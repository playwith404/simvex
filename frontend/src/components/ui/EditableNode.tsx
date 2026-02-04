import { memo } from 'react'
import { Handle, Position } from 'reactflow'
import type { NodeProps } from 'reactflow'

type NodeData = {
  label: string
  color?: string
  onChange: (id: string, value: string) => void
  onSelect: (id: string) => void
}

const EditableNodeComponent = ({ id, data }: NodeProps<NodeData>) => {
  const accent = data.color || '#f8c86a'

  return (
    <div
      className="flow-node"
      style={{ borderColor: accent, boxShadow: `0 0 0 1px ${accent}40` }}
      onClick={() => data.onSelect(id)}
    >
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
      <textarea
        value={data.label}
        onChange={(e) => data.onChange(id, e.target.value)}
      />
    </div>
  )
}

export const EditableNode = memo(EditableNodeComponent)
