import type { ChangeEvent } from 'react'

export const NoteEditor = ({
  value,
  onChange,
}: {
  value: string
  onChange: (value: string) => void
}) => {
  const handleChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    onChange(event.target.value)
  }

  return (
    <div className="note-editor">
      <textarea
        value={value}
        onChange={handleChange}
        placeholder="학습한 내용을 정리해보세요."
      />
    </div>
  )
}
