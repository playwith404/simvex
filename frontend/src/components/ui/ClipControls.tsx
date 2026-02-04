import type { ChangeEvent } from 'react'

export const ClipControls = ({
  enabled,
  axis,
  position,
  onToggle,
  onAxisChange,
  onPositionChange,
}: {
  enabled: boolean
  axis: 'x' | 'y' | 'z'
  position: number
  onToggle: () => void
  onAxisChange: (axis: 'x' | 'y' | 'z') => void
  onPositionChange: (position: number) => void
}) => {
  const handlePosition = (e: ChangeEvent<HTMLInputElement>) => {
    onPositionChange(Number(e.target.value) / 100)
  }

  return (
    <div className="clip-controls">
      <label className="clip-controls__toggle">
        <input type="checkbox" checked={enabled} onChange={onToggle} />
        단면 절단
      </label>
      {enabled && (
        <div className="clip-controls__options">
          <div className="clip-controls__axis">
            {(['x', 'y', 'z'] as const).map((a) => (
              <button
                key={a}
                type="button"
                className={axis === a ? 'active' : ''}
                onClick={() => onAxisChange(a)}
              >
                {a.toUpperCase()}축
              </button>
            ))}
          </div>
          <div className="clip-controls__slider">
            <input
              type="range"
              min={-500}
              max={500}
              value={Math.round(position * 100)}
              onChange={handlePosition}
            />
            <span>{position.toFixed(2)}</span>
          </div>
        </div>
      )}
    </div>
  )
}
