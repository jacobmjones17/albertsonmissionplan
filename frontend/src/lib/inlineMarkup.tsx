import type { ReactNode } from 'react'

/**
 * Minimal safe inline markup for ward-plan lines edited as plain text:
 * - **phrase** → <strong>
 * - [label](/path) or [label](https://…) → <a> (only safe hrefs)
 */
export function renderInlineMarkup(text: string): ReactNode {
  const re = /\*\*([\s\S]*?)\*\*|\[([^\]]+)\]\(([^)]+)\)/g
  const out: ReactNode[] = []
  let last = 0
  let k = 0
  for (;;) {
    const m = re.exec(text)
    if (!m) break
    if (m.index > last) {
      out.push(text.slice(last, m.index))
    }
    if (m[1] !== undefined) {
      out.push(<strong key={k++}>{m[1]}</strong>)
    } else {
      const href = safeHref(m[3].trim())
      const label = m[2]
      out.push(
        href ? (
          <a key={k++} href={href} className="inline-markup-link">
            {label}
          </a>
        ) : (
          <span key={k++}>{m[0]}</span>
        ),
      )
    }
    last = re.lastIndex
  }
  if (last < text.length) {
    out.push(text.slice(last))
  }
  if (out.length === 0) {
    return text
  }
  return <>{out}</>
}

function safeHref(raw: string): string | null {
  if (!raw || raw.includes('\n') || raw.includes('\r')) {
    return null
  }
  const lower = raw.toLowerCase()
  if (lower.startsWith('javascript:') || lower.startsWith('data:')) {
    return null
  }
  if (raw.startsWith('/')) {
    return raw
  }
  if (raw.startsWith('https://')) {
    return raw
  }
  if (raw.startsWith('http://localhost') || raw.startsWith('http://127.0.0.1')) {
    return raw
  }
  return null
}
