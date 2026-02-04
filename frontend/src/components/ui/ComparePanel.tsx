import type { Part } from '../../types'

export const ComparePanel = ({
  parts,
  compareIds,
  onRemove,
  onClear,
}: {
  parts: Part[]
  compareIds: string[]
  onRemove: (id: string) => void
  onClear: () => void
}) => {
  const compareParts = compareIds
    .map((id) => parts.find((p) => p.id === id))
    .filter((p): p is Part => p != null)

  if (compareParts.length === 0) {
    return (
      <div className="compare-panel">
        <p className="compare-panel__empty">
          부품 목록에서 비교할 부품을 선택하세요. (최대 2개)
        </p>
      </div>
    )
  }

  const fields: { label: string; key: keyof Part }[] = [
    { label: '부품명', key: 'name' },
    { label: '재질', key: 'material' },
    { label: '역할', key: 'role' },
  ]

  return (
    <div className="compare-panel">
      <div className="compare-panel__header">
        <h4>부품 비교</h4>
        <button type="button" className="compare-panel__clear" onClick={onClear}>
          초기화
        </button>
      </div>
      <table className="compare-panel__table">
        <thead>
          <tr>
            <th />
            {compareParts.map((p) => (
              <th key={p.id}>
                {p.name}
                <button
                  type="button"
                  className="compare-panel__remove"
                  onClick={() => onRemove(p.id)}
                >
                  ✕
                </button>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {fields.map((f) => (
            <tr key={f.key}>
              <td className="compare-panel__label">{f.label}</td>
              {compareParts.map((p) => (
                <td key={p.id}>{String(p[f.key])}</td>
              ))}
            </tr>
          ))}
          <tr>
            <td className="compare-panel__label">분해 거리</td>
            {compareParts.map((p) => (
              <td key={p.id}>{p.decomposeDistance.toFixed(2)}</td>
            ))}
          </tr>
        </tbody>
      </table>
    </div>
  )
}
