import type { JSX, TextareaHTMLAttributes } from 'react'
import { useId, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { computePrivacyFlags, flagDismissKey, privacyFlagHint, type PrivacyFlag } from '../privacyFlags'

function highlightSegments(body: string, spans: PrivacyFlag[]): JSX.Element[] {
  if (!spans.length) {
    return [<span key="full">{body}</span>]
  }

  const ordered = [...spans].sort((a, b) => a.start - b.start || a.end - b.end)
  const parts: JSX.Element[] = []
  let cursor = 0

  ordered.forEach((f, ix) => {
    if (f.start < cursor) {
      return
    }
    if (cursor < f.start) {
      parts.push(<span key={`gap-${cursor}-${ix}`}>{body.slice(cursor, f.start)}</span>)
    }
    parts.push(
      <mark key={`hit-${ix}-${f.start}-${f.end}`} className="privacy-flag-mark">
        {body.slice(f.start, f.end)}
      </mark>,
    )
    cursor = f.end
  })

  if (cursor < body.length) {
    parts.push(<span key={`trail-${cursor}`}>{body.slice(cursor)}</span>)
  }

  return parts
}

type Props = {
  /** Visible label; associated with the textarea for accessibility. */
  label: string
  value: string
  onChange: (next: string) => void
  maxLength?: number
  textareaProps?: Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'value' | 'className' | 'id'>
}

/**
 * Optional checks first when present, then label, then layered highlight field (yellow underlay +
 * transparent typing layer). Surfaces use opaque tokens only.
 */
export function PrivacyHighlightedBodyField({ label, value, onChange, maxLength, textareaProps }: Props) {
  const fieldId = useId()
  const taRef = useRef<HTMLTextAreaElement>(null)
  const [scroll, setScroll] = useState({ x: 0, y: 0 })
  /** Match textarea layout viewport so mirror wraps identically (scrollbar / gutter otherwise desync lines). */
  const [taBox, setTaBox] = useState<{ h?: number; w?: number }>({})
  const [dismissed, setDismissed] = useState<Set<string>>(() => new Set())

  const computed = useMemo(() => computePrivacyFlags(value), [value])

  const active = useMemo(
    () => computed.filter((f) => !dismissed.has(flagDismissKey(f))),
    [computed, dismissed],
  )

  const activeKeys = useMemo(() => new Set(active.map(flagDismissKey)), [active])

  function syncTaBox(el: HTMLTextAreaElement | null) {
    if (!el) {
      return
    }
    const h = el.clientHeight
    const w = el.clientWidth
    setTaBox((prev) => (prev.h === h && prev.w === w ? prev : { h, w }))
  }

  /** Outer box: user resize, window. Inner width also changes when a vertical scrollbar appears — that often does not resize the border box, so we sync on value/focus too. */
  useLayoutEffect(() => {
    const el = taRef.current
    if (!el || typeof ResizeObserver === 'undefined') {
      syncTaBox(taRef.current)
      return
    }
    const ro = new ResizeObserver(() => {
      syncTaBox(el)
    })
    ro.observe(el)
    syncTaBox(el)
    return () => ro.disconnect()
  }, [])

  useLayoutEffect(() => {
    syncTaBox(taRef.current)
  }, [value])

  const segments = useMemo(() => highlightSegments(value, active), [value, active])

  function dismiss(k: string) {
    setDismissed((prev) => {
      const next = new Set(prev)
      next.add(k)
      return next
    })
  }

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
          <div
            className="privacy-highlight-mask"
            style={{
              minHeight: '12rem',
              ...(taBox.h != null ? { height: taBox.h } : {}),
              ...(taBox.w != null ? { width: taBox.w } : {}),
            }}
            aria-hidden
          >
            <div
              className="privacy-highlight-shift"
              style={{ transform: `translate(${-scroll.x}px, ${-scroll.y}px)` }}
            >
              <div className="privacy-highlight-plain">{segments}</div>
            </div>
          </div>
          <textarea
            ref={taRef}
            className="privacy-highlight-input moderate-approve-form__body"
            {...textareaProps}
            id={fieldId}
            maxLength={maxLength}
            value={value}
            onChange={(e) => {
              textareaProps?.onChange?.(e)
              onChange(e.currentTarget.value)
            }}
            onFocus={(e) => {
              textareaProps?.onFocus?.(e)
              syncTaBox(e.currentTarget)
            }}
            onScroll={(e) => {
              textareaProps?.onScroll?.(e)
              const t = e.currentTarget
              setScroll({ x: t.scrollLeft, y: t.scrollTop })
            }}
          />
        </div>
      </div>
    </div>
  )
}
