import type { ObjectModel, Part, ChatMessage, Measurement } from '../../types'
import { NoteEditor } from './NoteEditor'
import { AIChat } from './AIChat'
import { PartInfoPanel } from './PartInfoPanel'
import { ProductInfoPanel } from './ProductInfoPanel'
import { PartListPanel } from './PartListPanel'
import { MeasurePanel } from './MeasurePanel'
import { ComparePanel } from './ComparePanel'

export const SidePanel = ({
  object,
  parts,
  selectedPart,
  activeTab,
  onTabChange,
  notes,
  onNotesChange,
  onSaveNotes,
  noteSaving,
  aiHistory,
  onSendMessage,
  aiLoading,
  hiddenPartIds,
  onSelectPart,
  onToggleVisibility,
  onToggleAllVisibility,
  measurements,
  onClearMeasurements,
  compareIds,
  onRemoveCompare,
  onClearCompare,
}: {
  object: ObjectModel
  parts: Part[]
  selectedPart: Part | null
  activeTab: 'note' | 'ai' | 'compare'
  onTabChange: (tab: 'note' | 'ai' | 'compare') => void
  notes: string
  onNotesChange: (value: string) => void
  onSaveNotes?: () => void
  noteSaving?: boolean
  aiHistory: ChatMessage[]
  onSendMessage: (message: string) => void
  aiLoading: boolean
  hiddenPartIds: Set<string>
  onSelectPart: (id: string) => void
  onToggleVisibility: (id: string) => void
  onToggleAllVisibility: (visible: boolean) => void
  measurements: Measurement[]
  onClearMeasurements: () => void
  compareIds: string[]
  onRemoveCompare: (id: string) => void
  onClearCompare: () => void
}) => {
  return (
    <aside className="side-panel">
      <PartListPanel
        parts={parts}
        selectedPartId={selectedPart?.id ?? null}
        hiddenPartIds={hiddenPartIds}
        onSelectPart={onSelectPart}
        onToggleVisibility={onToggleVisibility}
        onToggleAll={onToggleAllVisibility}
      />
      <ProductInfoPanel object={object} />
      <PartInfoPanel part={selectedPart} />
      <MeasurePanel measurements={measurements} onClear={onClearMeasurements} />
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
        <button
          type="button"
          className={activeTab === 'compare' ? 'active' : ''}
          onClick={() => onTabChange('compare')}
        >
          비교
        </button>
      </div>
      <div className="side-panel__content">
        {activeTab === 'note' ? (
          <>
            <NoteEditor value={notes} onChange={onNotesChange} />
            {onSaveNotes && (
              <div className="note-save">
                <button
                  type="button"
                  className="ghost"
                  onClick={onSaveNotes}
                  disabled={noteSaving}
                >
                  저장
                </button>
              </div>
            )}
          </>
        ) : activeTab === 'ai' ? (
          <AIChat history={aiHistory} isLoading={aiLoading} onSend={onSendMessage} />
        ) : (
          <ComparePanel
            parts={parts}
            compareIds={compareIds}
            onRemove={onRemoveCompare}
            onClear={onClearCompare}
          />
        )}
      </div>
    </aside>
  )
}
