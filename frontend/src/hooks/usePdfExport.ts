import { useState } from 'react'
import jsPDF from 'jspdf'
import type { ChatMessage } from '../types'
import type { PdfImageData } from '../types/pdf'

type TextSection = {
  title: string
  body: string
}

const buildTextCanvas = (title: string, sections: TextSection[]) => {
  const width = 1000
  const padding = 32
  const titleFont = '700 32px "Noto Sans KR", sans-serif'
  const headingFont = '600 24px "Noto Sans KR", sans-serif'
  const bodyFont = '400 18px "Noto Sans KR", sans-serif'
  const titleLineHeight = 40
  const headingLineHeight = 32
  const bodyLineHeight = 26

  const measureCanvas = document.createElement('canvas')
  const mctx = measureCanvas.getContext('2d')
  if (!mctx) return null

  const maxWidth = width - padding * 2
  const content: { lines: string[]; font: string; lineHeight: number; gapAfter: number }[] = []
  let height = padding

  mctx.font = titleFont
  const titleLines = wrapText(mctx, title, maxWidth)
  content.push({ lines: titleLines, font: titleFont, lineHeight: titleLineHeight, gapAfter: 12 })
  height += titleLines.length * titleLineHeight + 12

  sections.forEach((section, index) => {
    mctx.font = headingFont
    const headLines = wrapText(mctx, section.title, maxWidth)
    content.push({ lines: headLines, font: headingFont, lineHeight: headingLineHeight, gapAfter: 6 })
    height += headLines.length * headingLineHeight + 6

    mctx.font = bodyFont
    const bodyLines = wrapText(mctx, section.body, maxWidth)
    content.push({ lines: bodyLines, font: bodyFont, lineHeight: bodyLineHeight, gapAfter: index === sections.length - 1 ? 0 : 16 })
    height += bodyLines.length * bodyLineHeight + (index === sections.length - 1 ? 0 : 16)
  })

  height += padding

  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const ctx = canvas.getContext('2d')
  if (!ctx) return null

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  ctx.fillStyle = '#111111'

  let y = padding
  content.forEach((block) => {
    ctx.font = block.font
    block.lines.forEach((line) => {
      ctx.fillText(line, padding, y)
      y += block.lineHeight
    })
    y += block.gapAfter
  })

  return canvas
}

const wrapText = (ctx: CanvasRenderingContext2D, text: string, maxWidth: number) => {
  const lines: string[] = []
  let line = ''

  for (const char of text) {
    if (char === '\n') {
      if (line) lines.push(line)
      line = ''
      continue
    }

    const testLine = line + char
    if (ctx.measureText(testLine).width > maxWidth && line !== '') {
      lines.push(line)
      line = char
      continue
    }
    line = testLine
  }

  if (line) lines.push(line)
  if (lines.length === 0) lines.push('')
  return lines
}

export const usePdfExport = () => {
  const [isExporting, setIsExporting] = useState(false)

  const exportPdf = (options: {
    canvas?: HTMLCanvasElement | null
    getImageData?: () => PdfImageData | null
    objectName: string
    notes: string
    chatHistory: ChatMessage[]
    onSaved?: () => void
  }) => {
    if (!options.canvas && !options.getImageData) return
    setIsExporting(true)

    try {
      const pdf = new jsPDF('p', 'mm', 'a4')
      const payload =
        options.getImageData?.() ??
        (options.canvas
          ? {
              dataUrl: options.canvas.toDataURL('image/png'),
              width: options.canvas.width,
              height: options.canvas.height,
            }
          : null)
      if (!payload) return
      const pageWidth = 210
      const pageHeight = 297
      const marginX = 10
      const marginTop = 12
      const maxImgWidth = pageWidth - marginX * 2
      const maxImgHeight = pageHeight - marginTop * 2
      const aspect = payload.width / payload.height
      let imgWidth = maxImgWidth
      let imgHeight = imgWidth / aspect
      if (imgHeight > maxImgHeight) {
        imgHeight = maxImgHeight
        imgWidth = imgHeight * aspect
      }
      const imgX = marginX + (maxImgWidth - imgWidth) / 2

      pdf.addImage(payload.dataUrl, 'PNG', imgX, marginTop, imgWidth, imgHeight)

      const summary = options.chatHistory
        .slice(-6)
        .map((m) => `${m.role === 'user' ? 'Q' : 'A'}: ${m.content}`)
        .join('\n')

      const textCanvas = buildTextCanvas(`SIMVEX 학습 리포트 - ${options.objectName}`, [
        {
          title: '노트',
          body: options.notes || '작성된 노트가 없습니다.',
        },
        {
          title: 'AI 대화 요약',
          body: summary || '대화 기록이 없습니다.',
        },
      ])

      if (textCanvas) {
        const textImg = textCanvas.toDataURL('image/png')
        const textWidth = imgWidth
        const textHeight = (textCanvas.height / textCanvas.width) * textWidth
        let textY = marginTop + imgHeight + 10

        if (textY + textHeight > pageHeight - marginTop) {
          pdf.addPage()
          textY = marginTop
        }

        pdf.addImage(textImg, 'PNG', marginX, textY, textWidth, textHeight)
      }

      pdf.save(`simvex-${options.objectName}.pdf`)
      options.onSaved?.()
    } finally {
      setTimeout(() => setIsExporting(false), 0)
    }
  }

  return { exportPdf, isExporting }
}
