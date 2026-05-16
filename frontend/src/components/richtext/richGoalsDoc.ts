function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function escapeAttr(s: string): string {
  return s.replace(/"/g, '&quot;')
}

/** Legacy plain lines: **bold** and [text](url) → HTML (safe). */
export function legacyMarkdownToHtml(s: string): string {
  let t = escapeHtml(s)
  t = t.replace(/\*\*([\s\S]*?)\*\*/g, '<strong>$1</strong>')
  t = t.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label: string, href: string) => {
    const h = href.trim()
    if (
      h.startsWith('/') ||
      h.startsWith('https://') ||
      h.startsWith('http://localhost') ||
      h.startsWith('http://127.0.0.1')
    ) {
      return `<a href="${escapeAttr(h)}">${label}</a>`
    }
    return `[${label}](${href})`
  })
  return t
}

export function lineToListItemHtml(line: string): string {
  const t = line.trim()
  if (!t) return ''
  if (/<[a-z][\s\S]*>/i.test(t)) {
    return `<li>${t}</li>`
  }
  return `<li><p>${legacyMarkdownToHtml(t)}</p></li>`
}

/** Build TipTap / ProseMirror HTML for ward goal list. */
export function linesToBulletListHtml(lines: string[]): string {
  const items = lines.map((l) => lineToListItemHtml(l)).filter(Boolean)
  if (items.length === 0) {
    return '<ul><li><p></p></li></ul>'
  }
  return `<ul>${items.join('')}</ul>`
}

/** Split stored newline bullets from editor HTML. */
export function bulletListHtmlToLines(html: string): string[] {
  const trimmed = html.trim()
  if (!trimmed) return []
  const doc = new DOMParser().parseFromString(trimmed, 'text/html')
  const ul = doc.querySelector('ul')
  if (!ul) {
    const inner = doc.body.innerHTML.trim()
    if (!inner) return []
    return [inner]
  }
  return Array.from(ul.querySelectorAll(':scope > li'))
    .map((li) => li.innerHTML.trim())
    .filter(Boolean)
}

/** Newline-separated storage ↔ editor. */
export function storageToHtml(stored: string): string {
  const lines = stored.replace(/\r\n/g, '\n').split('\n').map((l) => l.trim()).filter(Boolean)
  return linesToBulletListHtml(lines)
}

export function htmlToStorage(html: string): string {
  return bulletListHtmlToLines(html).join('\n')
}
