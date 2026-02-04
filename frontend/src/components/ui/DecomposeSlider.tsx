import { useState, useEffect, useRef, type ChangeEvent } from 'react'

export const DecomposeSlider = ({
  value,
  onChange,
}: {
  value: number
  onChange: (value: number) => void
}) => {
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(1)
  const valueRef = useRef(value)
  const directionRef = useRef(1)
  const onChangeRef = useRef(onChange)

  useEffect(() => { valueRef.current = value }, [value])
  useEffect(() => { onChangeRef.current = onChange }, [onChange])

  useEffect(() => {
    if (!playing) return

    let raf: number
    let lastTime = 0

    const animate = (time: number) => {
      if (!lastTime) lastTime = time
      const delta = (time - lastTime) / 1000
      lastTime = time

      let next = valueRef.current + directionRef.current * speed * 0.3 * delta

      if (next >= 1) {
        next = 1
        directionRef.current = -1
      } else if (next <= 0) {
        next = 0
        directionRef.current = 1
      }

      onChangeRef.current(next)
      raf = requestAnimationFrame(animate)
    }

    raf = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(raf)
  }, [playing, speed])

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange(Number(event.target.value) / 100)
  }

  const handleTogglePlay = () => {
    if (!playing) {
      directionRef.current = valueRef.current >= 1 ? -1 : 1
    }
    setPlaying(!playing)
  }

  return (
    <div className="decompose-slider">
      <div className="decompose-slider__header">
        <label>분해도: {Math.round(value * 100)}%</label>
        <div className="decompose-slider__actions">
          <button type="button" className="decompose-slider__play" onClick={handleTogglePlay}>
            {playing ? '⏸' : '▶'}
          </button>
          {playing && (
            <select
              className="decompose-slider__speed"
              value={speed}
              onChange={(e) => setSpeed(Number(e.target.value))}
            >
              <option value={0.5}>0.5x</option>
              <option value={1}>1x</option>
              <option value={2}>2x</option>
              <option value={3}>3x</option>
            </select>
          )}
        </div>
      </div>
      <input
        type="range"
        min={0}
        max={100}
        value={Math.round(value * 100)}
        onChange={handleChange}
      />
      <div className="decompose-slider__labels">
        <span>조립</span>
        <span>분해</span>
      </div>
    </div>
  )
}
