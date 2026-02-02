import { useState } from 'react'
import jsPDF from 'jspdf'
import type { ChatMessage } from '../types'

export const usePdfExport = () => {
  const [isExporting, setIsExporting] = useState(false)

  const exportPdf = async (options: {
    canvas: HTMLCanvasElement | null
    objectName: string
    notes: string
    chatHistory: ChatMessage[]
  }) => {
    if (!options.canvas) return
    setIsExporting(true)

    try {
      const pdf = new jsPDF('p', 'mm', 'a4')
      const imgData = options.canvas.toDataURL('image/png')
      const pageWidth = 210
      const imgWidth = pageWidth - 20
      const imgHeight = (options.canvas.height / options.canvas.width) * imgWidth

      pdf.setFontSize(16)
      pdf.text(`SIMVEX 학습 리포트 - ${options.objectName}`, 10, 15)
      pdf.addImage(imgData, 'PNG', 10, 20, imgWidth, Math.min(imgHeight, 120))

      pdf.setFontSize(12)
      pdf.text('노트', 10, 150)
      pdf.setFontSize(10)
      pdf.text(options.notes || '작성된 노트가 없습니다.', 10, 158, { maxWidth: 190 })

      const summaryY = 200
      pdf.setFontSize(12)
      pdf.text('AI 대화 요약', 10, summaryY)
      pdf.setFontSize(10)

      const summary = options.chatHistory
        .slice(-6)
        .map((m) => `${m.role === 'user' ? 'Q' : 'A'}: ${m.content}`)
        .join('\n')

      pdf.text(summary || '대화 기록이 없습니다.', 10, summaryY + 8, { maxWidth: 190 })

      pdf.save(`simvex-${options.objectName}.pdf`)
    } finally {
      setIsExporting(false)
    }
  }

  return { exportPdf, isExporting }
}
