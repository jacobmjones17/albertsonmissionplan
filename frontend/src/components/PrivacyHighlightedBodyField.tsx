import type { TextareaHTMLAttributes } from 'react'
import { useId, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { computePrivacyFlags, flagDismissKey, privacyFlagHint, type PrivacyFlag } from '../privacyFlags'

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function buildHighlightHtml(body: string, spans: PrivacyFlag[]): string {
  if (!body) {
    return ''
  }
  if (!spans.length) {
    return escapeHtml(body).replace(/\n/g, '<br>')
  }

  const ordered = [...spans].sort((a, b) => a.start - b.start || a.end - b.end)
  let html = ''
  let cursor = 0

  for (const f of ordered) {
    if (f.start < cursor) {
      continue
    }
    html += escapeHtml(body.slice(cursor, f.start))
    html += `<mark class="privacy-flag-mark">${escapeHtml(body.slice(f.start, f.end))}</mark>`
    cursor = f.end
  }
  html += escapeHtml(body.slice(cursor))
  return html.replace(/\n/g, '<br>')
}

function readPlainText(root: HTMLElement): string {
  return root.innerText.replace(/\r\n/g, '\n')
}

function getCaretOffset(root: HTMLElement): number | null {
  const sel = window.getSelection()
  if (!sel || sel.rangeCount === 0) {
    return null
  }
  const range = sel.getRangeAt(0)
  if (!root.contains(range.startContainer)) {
    return null
  }
  const pre = document.createRange()
  pre.selectNodeContents(root)
  pre.setEnd(range.startContainer, range.startOffset)
  return pre.toString().length
}

function setCaretOffset(root: HTMLElement, offset: number) {
  const sel = window.getSelection()
  if (!sel) {
    return
  }
  let remaining = Math.max(0, offset)
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  let node = walker.nextNode()
  while (node) {
    const len = node.textContent?.length ?? 0
    if (remaining <= len) {
      const range = document.createRange()
      range.setStart(node, remaining)
      range.collapse(true)
      sel.removeAllRanges()
      sel.addRange(range)
      return
    }
    remaining -= len
    node = walker.nextNode()
  }
  const end = document.createRange()
  end.selectNodeContents(root)
  end.collapse(false)
  sel.removeAllRanges()
  sel.addRange(end)
}

type Props = {
  label: string
  value: string
  onChange: (next: string) => void
  maxLength?: number
  textareaProps?: Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'value' | 'className' | 'id'>
}

/**
 * Privacy checks + single contenteditable field so yellow marks and the caret share one layout
 * (overlay textarea + HTML mirror misaligns cursor/selection when marks change wrapping).
 */
export function PrivacyHighlightedBodyField({ label, value, onChange, maxLength, textareaProps }: Props) {
  const fieldId = useId()
  const editRef = useRef<HTMLDivElement>(null)
  const lastHtmlRef = useRef('')
  const [dismissed, setDismissed] = useState<Set<string>>(() => new Set())

  const computed = useMemo(() => computePrivacyFlags(value), [value])

  const active = useMemo(
    () => computed.filter((f) => !dismissed.has(flagDismissKey(f))),
    [computed, dismissed],
  )

  const activeKeys = useMemo(() => new Set(active.map(flagDismissKey)), [active])

  const highlightHtml = useMemo(() => buildHighlightHtml(value, active), [value, active])

  useLayoutEffect(() => {
    const el = editRef.current
    if (!el || highlightHtml === lastHtmlRef.current) {
      return
    }
    const focused = document.activeElement === el
    const caret = focused ? getCaretOffset(el) : null

    lastHtmlRef.current = highlightHtml
    el.innerHTML = highlightHtml || '<br>'

    if (caret != null) {
      setCaretOffset(el, Math.min(caret, value.length))
    }
  }, [highlightHtml, value.length])

  function dismiss(k: string) {
    setDismissed((prev) => {
      const next = new Set(prev)
      next.add(k)
      return next
    })
  }

  function emitFromEditor() {
    const el = editRef.current
    if (!el) {
      return
    }
    let next = readPlainText(el)
    if (maxLength != null && next.length > maxLength) {
      next = next.slice(0, maxLength)
      lastHtmlRef.current = ''
    }
    onChange(next)
  }

  const { name, required, rows } = textareaProps ?? {}

  return (
    <div className="privacy-check-stack">
      {computed.length > 0 ? (
        <div className="privacy-flag-panel" role="region" aria-label="Automatic privacy checks">
          <p className="privacy-flag-panel-intro">
            <strong>Quick scan</strong> — same text is marked in yellow below. Heuristic only.
          </p>
          {!active.length ? (
            <p className="privacy-flag-panel-note">All flags dismissed for this draft.</p>
          ) : null}
          <ul className="privacy-flag-list">
            {computed.map((f) => {
              const dk = flagDismissKey(f)
              const isShowing = activeKeys.has(dk)
              return (
                <li key={dk} className={`privacy-flag-item${isShowing ? ' is-active' : ''}`}>
                  <div className="privacy-flag-line">
                    <span className="privacy-flag-phrase">&quot;{f.text}&quot;</span>
                    <span className="privacy-flag-kind">{privacyFlagHint(f.kind)}</span>
                    {!dismissed.has(dk) ? (
                      <button type="button" className="privacy-flag-ignore" onClick={() => dismiss(dk)}>
                        Ignore
                      </button>
                    ) : (
                      <span className="privacy-flag-muted">Ignored</span>
                    )}
                  </div>
                </li>
              )
            })}
          </ul>
        </div>
      ) : null}

      <div className="privacy-editor-block">
        <label htmlFor={fieldId} className="privacy-body-field-label">
          {label}
        </label>
        <div className="privacy-highlight-shell">
          {name ? (
            <textarea
              className="sr-only"
              tabIndex={-1}
              aria-hidden
              name={name}
              value={value}
              required={required}
              readOnly
            />
          ) : null}
          <div
            ref={editRef}
            id={fieldId}
            role="textbox"
            aria-multiline="true"
            aria-required={required || undefined}
            contentEditable
            suppressContentEditableWarning
            className="privacy-highlight-editable moderate-approve-form__body"
            data-rows={rows ?? 10}
            onInput={() => emitFromEditor()}
            onPaste={(e) => {
              e.preventDefault()
              const text = e.clipboardData.getData('text/plain')
              document.execCommand('insertText', false, text)
            }}
          />
        </div>
      </div>
    </div>
  )
}
