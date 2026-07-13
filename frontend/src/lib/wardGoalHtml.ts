import DOMPurify, { type Config as DomPurifyConfig } from 'dompurify'
import { escapeHtml } from './downloadDocument'

const PURIFY_OPTS: DomPurifyConfig = {
  ALLOWED_TAGS: [
    'strong',
    'b',
    'em',
    'i',
    'u',
    's',
    'strike',
    'del',
    'a',
    'span',
    'br',
    'p',
    'h2',
    'h3',
    'h4',
    'ul',
    'ol',
    'li',
    'sub',
    'sup',
    'mark',
    'code',
    'blockquote',
  ],
  ALLOWED_ATTR: ['href', 'title', 'rel', 'target', 'class', 'style'],
  ADD_ATTR: ['target'],
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

function inlineMarkupToHtml(text: string): string {
  const re = /\*\*([\s\S]*?)\*\*|\[([^\]]+)\]\(([^)]+)\)/g
  let result = ''
  let last = 0
  for (;;) {
    const match = re.exec(text)
    if (!match) break
    result += escapeHtml(text.slice(last, match.index))
    if (match[1] !== undefined) {
      result += `<strong>${escapeHtml(match[1])}</strong>`
    } else {
      const href = safeHref(match[3].trim())
      if (href) {
        result += `<a href="${escapeHtml(href)}" class="inline-markup-link">${escapeHtml(match[2])}</a>`
      } else {
        result += escapeHtml(match[0])
      }
    }
    last = re.lastIndex
  }
  result += escapeHtml(text.slice(last))
  return result
}

/** Strip a lone <p> wrapper so **markdown** inside editor HTML still bolds on display. */
function normalizeWardGoalContent(content: string): string {
  const trimmed = content.trim()
  const paragraph = trimmed.match(/^<p>([\s\S]*)<\/p>$/i)
  if (!paragraph) {
    return trimmed
  }
  const inner = paragraph[1].trim()
  if (/<(?:strong|b|a|ul|ol|h[2-4])\b/i.test(inner)) {
    return trimmed
  }
  return inner
}

/** Safe HTML for one ward-plan goal line (download + public ward plan page). */
export function wardGoalLineHtml(content: string): string {
  const normalized = normalizeWardGoalContent(content)
  if (/<[a-z][\s\S]*>/i.test(normalized)) {
    return DOMPurify.sanitize(normalized, PURIFY_OPTS)
  }
  return inlineMarkupToHtml(normalized)
}
