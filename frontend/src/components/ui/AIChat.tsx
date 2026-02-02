import { useState } from 'react'
import type { ChatMessage } from '../../types'

export const AIChat = ({
  history,
  isLoading,
  onSend,
}: {
  history: ChatMessage[]
  isLoading: boolean
  onSend: (message: string) => void
}) => {
  const [message, setMessage] = useState('')

  const handleSend = () => {
    if (!message.trim()) return
    onSend(message.trim())
    setMessage('')
  }

  return (
    <div className="ai-chat">
      <div className="ai-chat__history">
        {history.length === 0 ? (
          <p className="ai-chat__empty">AI에게 질문해보세요.</p>
        ) : (
          history.map((item, index) => (
            <div key={`${item.role}-${index}`} className={`ai-chat__bubble ${item.role}`}>
              <span>{item.content}</span>
            </div>
          ))
        )}
        {isLoading && <div className="ai-chat__bubble assistant">답변 생성 중...</div>}
      </div>
      <div className="ai-chat__input">
        <input
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="질문을 입력하세요"
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSend()
          }}
        />
        <button type="button" onClick={handleSend} disabled={isLoading}>
          전송
        </button>
      </div>
    </div>
  )
}
