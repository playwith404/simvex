import { useState } from 'react'
import type { Part } from '../../types'

export const PartListPanel = ({
  parts,
  selectedPartId,
  hiddenPartIds,
  onSelectPart,
  onToggleVisibility,
  onToggleAll,
}: {
  parts: Part[]
  selectedPartId: string | null
  hiddenPartIds: Set<string>
  onSelectPart: (id: string) => void
  onToggleVisibility: (id: string) => void
  onToggleAll: (visible: boolean) => void
}) => {
  const [search, setSearch] = useState('')

  const filtered = parts.filter((p) => {
    const q = search.toLowerCase()
    return (
      p.name.toLowerCase().includes(q) ||
      p.material.toLowerCase().includes(q) ||
      p.role.toLowerCase().includes(q)
    )
  })

  const allVisible = hiddenPartIds.size === 0

  return (
    <div className="part-list">
      <div className="part-list__header">
        <h4>부품 목록</h4>
        <button
          type="button"
          className="part-list__toggle-all"
          onClick={() => onToggleAll(!allVisible)}
        >
          {allVisible ? '전체 숨기기' : '전체 보기'}
        </button>
      </div>
      <input
        className="part-list__search"
        type="text"
        placeholder="부품 검색 (이름, 재질, 역할)"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />
      <ul className="part-list__items">
        {filtered.map((part) => (
          <li
            key={part.id}
            className={`part-list__item${selectedPartId === part.id ? ' selected' : ''}`}
          >
            <input
              type="checkbox"
              checked={!hiddenPartIds.has(part.id)}
              onChange={() => onToggleVisibility(part.id)}
            />
            <button type="button" onClick={() => onSelectPart(part.id)}>
              {part.name}
            </button>
          </li>
        ))}
        {filtered.length === 0 && (
          <li className="part-list__empty">검색 결과가 없습니다.</li>
        )}
      </ul>
    </div>
  )
}
