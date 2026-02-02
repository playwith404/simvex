import { memo } from 'react'
import type { NodeProps } from 'reactflow'

export type Attachment = {
  type: 'file' | 'link'
  name: string
  url: string
}

type NodeData = {
  label: string
  attachments?: Attachment[]
  onChange: (id: string, value: string) => void
  onAddAttachment: (id: string, attachment: Attachment) => void
  onRemoveAttachment: (id: string, index: number) => void
}

const EditableNodeComponent = ({ id, data }: NodeProps<NodeData>) => {
  return (
    <div className="flow-node">
      <textarea
        value={data.label}
        onChange={(e) => data.onChange(id, e.target.value)}
      />
      <div className="flow-node__attachments">
        {(data.attachments || []).map((att, index) => (
          <div key={`${att.type}-${index}`} className="flow-node__attachment">
            <a href={att.url} target="_blank" rel="noreferrer">
              {att.name}
            </a>
            <button type="button" onClick={() => data.onRemoveAttachment(id, index)}>
              삭제
            </button>
          </div>
        ))}
      </div>
      <div className="flow-node__actions">
        <button
          type="button"
          onClick={() => {
            const url = window.prompt('첨부할 링크 URL을 입력하세요')
            if (!url) return
            const name = window.prompt('표시 이름을 입력하세요') || '링크'
            data.onAddAttachment(id, { type: 'link', name, url })
          }}
        >
          링크 추가
        </button>
        <label className="file-button">
          파일 추가
          <input
            type="file"
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (!file) return
              const reader = new FileReader()
              reader.onload = () => {
                const url = String(reader.result)
                data.onAddAttachment(id, { type: 'file', name: file.name, url })
              }
              reader.readAsDataURL(file)
              e.currentTarget.value = ''
            }}
          />
        </label>
      </div>
    </div>
  )
}

export const EditableNode = memo(EditableNodeComponent)
