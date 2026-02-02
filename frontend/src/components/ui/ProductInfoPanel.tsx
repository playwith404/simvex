import type { ObjectModel } from '../../types'

export const ProductInfoPanel = ({ object }: { object: ObjectModel }) => {
  return (
    <div className="info-panel">
      <h4>완제품 정보</h4>
      <div className="info-panel__item">
        <span>완제품명</span>
        <strong>{object.name}</strong>
      </div>
      <div className="info-panel__item">
        <span>설명</span>
        <strong>{object.description}</strong>
      </div>
      <div className="info-panel__item">
        <span>관련 이론</span>
        <strong>{object.theory}</strong>
      </div>
    </div>
  )
}
