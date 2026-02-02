import type { ChangeEvent } from 'react'

export const DecomposeSlider = ({
  value,
  onChange,
}: {
  value: number
  onChange: (value: number) => void
}) => {
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange(Number(event.target.value) / 100)
  }

  return (
    <div className="decompose-slider">
      <label>분해도: {Math.round(value * 100)}%</label>
      <input type="range" min={0} max={100} value={Math.round(value * 100)} onChange={handleChange} />
      <div className="decompose-slider__labels">
        <span>조립</span>
        <span>분해</span>
      </div>
    </div>
  )
}
