import type { ObjectModel, Part, ChatMessage } from '../../types'
import { NoteEditor } from './NoteEditor'
import { AIChat } from './AIChat'
import { PartInfoPanel } from './PartInfoPanel'
import { ProductInfoPanel } from './ProductInfoPanel'

export const SidePanel = ({
  object,
  selectedPart,
  activeTab,
  onTabChange,
  notes,
  onNotesChange,
  aiHistory,
  onSendMessage,
  aiLoading,
}: {
  object: ObjectModel
  selectedPart: Part | null
  activeTab: 'note' | 'ai'
  onTabChange: (tab: 'note' | 'ai') => void
  notes: string
  onNotesChange: (value: string) => void
  aiHistory: ChatMessage[]
  onSendMessage: (message: string) => void
  aiLoading: boolean
}) => {
  return (
    <aside className="side-panel">
      <ProductInfoPanel object={object} />
      <PartInfoPanel part={selectedPart} />
      <div className="side-panel__tabs">
        <button
          type="button"
          className={activeTab === 'note' ? 'active' : ''}
          onClick={() => onTabChange('note')}
        >
          노트
        </button>
        <button
          type="button"
          className={activeTab === 'ai' ? 'active' : ''}
          onClick={() => onTabChange('ai')}
        >
          AI 어시스턴트
        </button>
      </div>
      <div className="side-panel__content">
        {activeTab === 'note' ? (
          <NoteEditor value={notes} onChange={onNotesChange} />
        ) : (
          <AIChat history={aiHistory} isLoading={aiLoading} onSend={onSendMessage} />
        )}
      </div>
    </aside>
  )
}
