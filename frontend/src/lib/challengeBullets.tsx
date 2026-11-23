import type { ReactNode } from 'react'
import { escapeHtml } from './downloadDocument'

/**
 * Challenge idea text is plain text with two optional bits of markup that leaders can add from
 * the editor: **bold** and [label](https://link). Anything else shows exactly as typed.
 */
const MARKUP_RE = /\*\*([\s\S]*?)\*\*|\[([^\]]+)\]\(([^)\s]+)\)/g

function safeHref(raw: string): string | null {
  const href = raw.trim()
  if (/^https:\/\/[^\s"'<>]+$/i.test(href)) return href
  if (/^mailto:[^\s"'<>]+$/i.test(href)) return href
  if (/^\/[^\s"'<>]*$/.test(href) && !href.startsWith('//')) return href
  return null
}

export function renderChallengeText(text: string): ReactNode {
  const out: ReactNode[] = []
  const re = new RegExp(MARKUP_RE.source, 'g')
  let last = 0
  let k = 0
  for (let m = re.exec(text); m; m = re.exec(text)) {
    if (m.index > last) out.push(text.slice(last, m.index))
    if (m[1] !== undefined) {
      out.push(<strong key={k++}>{m[1]}</strong>)
    } else {
      const href = safeHref(m[3]!)
      out.push(
        href ? (
          <a key={k++} href={href} rel="noopener noreferrer">
            {m[2]}
          </a>
        ) : (
          m[0]
        ),
      )
    }
    last = re.lastIndex
  }
  if (last < text.length) out.push(text.slice(last))
  return out.length === 1 ? out[0] : <>{out}</>
}

export function challengeTextHtml(text: string): string {
  const re = new RegExp(MARKUP_RE.source, 'g')
  let html = ''
  let last = 0
  for (let m = re.exec(text); m; m = re.exec(text)) {
    html += escapeHtml(text.slice(last, m.index))
    if (m[1] !== undefined) {
      html += `<strong>${escapeHtml(m[1])}</strong>`
    } else {
      const href = safeHref(m[3]!)
      html += href
        ? `<a href="${escapeHtml(href)}" rel="noopener noreferrer">${escapeHtml(m[2]!)}</a>`
        : escapeHtml(m[0])
    }
    last = re.lastIndex
  }
  return html + escapeHtml(text.slice(last))
}

export function challengeBulletHtml(who: string, text: string): string {
  return `<li><span class="challenge-list-label">${escapeHtml(who)}:</span><span class="challenge-list-text">${challengeTextHtml(text)}</span></li>`
}

export function renderChallengeBullet(who: string, text: string, key: number) {
  return (
    <li key={key}>
      <span className="challenge-list-label">{who}:</span>
      <span className="challenge-list-text">{renderChallengeText(text)}</span>
    </li>
  )
}
