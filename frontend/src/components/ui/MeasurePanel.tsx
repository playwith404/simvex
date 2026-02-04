import type { Measurement } from '../../types'

export const MeasurePanel = ({
  measurements,
  onClear,
}: {
  measurements: Measurement[]
  onClear: () => void
}) => {
  if (measurements.length === 0) return null

  return (
    <div className="measure-panel">
      <div className="measure-panel__header">
        <h4>측정 결과</h4>
        <button type="button" className="measure-panel__clear" onClick={onClear}>
          초기화
        </button>
      </div>
      <ul className="measure-panel__list">
        {measurements.map((m, i) => (
          <li key={m.id} className="measure-panel__item">
            <span>측정 {i + 1}</span>
            <strong>{m.distance.toFixed(4)}</strong>
          </li>
        ))}
      </ul>
    </div>
  )
}
