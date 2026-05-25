/** Trigger a browser download for a text/HTML blob. */
export function downloadTextFile(content: string, filename: string, mime = 'text/html;charset=utf-8') {
  const blob = new Blob([content], { type: mime })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.rel = 'noopener'
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}

export function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

const PRINT_DOC_STYLES = `
  :root { color-scheme: light; }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    padding: 2rem 1.25rem 3rem;
    font-family: "Source Sans 3", system-ui, sans-serif;
    font-size: 11pt;
    line-height: 1.5;
    color: #1b1812;
    background: #fff;
  }
  h1, h2, h3 { font-family: "Cormorant Garamond", Georgia, serif; font-weight: 600; line-height: 1.2; }
  h1 { font-size: 1.85rem; margin: 0 0 0.35rem; }
  h2 { font-size: 1.35rem; margin: 1.75rem 0 0.65rem; page-break-after: avoid; }
  h3 { font-size: 1.05rem; margin: 0 0 0.35rem; page-break-after: avoid; }
  p { margin: 0 0 0.75rem; }
  .doc-eyebrow {
    margin: 0 0 0.35rem;
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #5b554a;
  }
  .doc-lede { color: #5b554a; max-width: 42rem; }
  .doc-meta { margin-top: 0.5rem; font-size: 0.85rem; color: #8c8676; }
  ul { margin: 0.25rem 0 0.75rem; padding-left: 1.25rem; }
  li { margin: 0.35rem 0; }
  li::marker { color: #b08745; }
  .doc-section { page-break-inside: avoid; }
  .doc-month { page-break-after: always; }
  .doc-month:last-child { page-break-after: auto; }
  .doc-org { margin-top: 1rem; page-break-inside: avoid; }
  .doc-quote {
    margin: 1.5rem 0 0;
    padding: 0.75rem 0 0.75rem 1rem;
    border-left: 3px solid #b08745;
    font-family: "Cormorant Garamond", Georgia, serif;
    font-size: 1.05rem;
    font-style: italic;
    color: #5b554a;
  }
  a { color: #1f3c63; }
  @media print {
    body { padding: 0; }
    .doc-month { page-break-after: always; }
  }
`

export function wrapPrintHtml(title: string, body: string, extraStyles = '', offline = false): string {
  const fonts = offline
    ? ''
    : `<link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:wght@600;700&family=Source+Sans+3:wght@400;600&display=swap" rel="stylesheet" />`
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>${escapeHtml(title)}</title>
  ${fonts}
  <style>${extraStyles || PRINT_DOC_STYLES}</style>
</head>
<body>
${body}
</body>
</html>`
}
