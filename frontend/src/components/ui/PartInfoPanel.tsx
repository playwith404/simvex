import type { Part } from '../../types'

export const PartInfoPanel = ({ part }: { part: Part | null }) => {
  return (
    <div className="info-panel">
      <h4>부품 정보</h4>
      {part ? (
        <>
          <div className="info-panel__item">
            <span>부품명</span>
            <strong>{part.name}</strong>
          </div>
          <div className="info-panel__item">
            <span>재질</span>
            <strong>{part.material}</strong>
          </div>
          <div className="info-panel__item">
            <span>역할</span>
            <strong>{part.role}</strong>
          </div>
        </>
      ) : (
        <p className="info-panel__empty">부품을 선택하면 정보가 표시됩니다.</p>
      )}
    </div>
  )
}
