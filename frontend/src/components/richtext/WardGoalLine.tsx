import { useMemo } from 'react'
import { wardGoalLineHtml } from '../../lib/wardGoalHtml'

/** One ward-plan bullet: HTML from the rich editor, or legacy plain / **markdown**. */
export function WardGoalLine({ content }: { content: string }) {
  const html = useMemo(() => wardGoalLineHtml(content), [content])

  return (
    <div
      className="ward-rich-line"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
