import { createPortal } from 'react-dom'
import { useCallback, useEffect, useRef } from 'react'

export type ToastPayload = {
  message: string
  variant?: 'success' | 'error'
}

type ToastProps = {
  toast: ToastPayload | null
  onDismiss: () => void
  /** Time before auto-hide (ms). */
  durationMs?: number
}

/**
 * Fixed toast for save / admin feedback. Portals to `document.body`; auto-dismisses.
 */
export function Toast({ toast, onDismiss, durationMs = 4200 }: ToastProps) {
  const onDismissRef = useRef(onDismiss)
  onDismissRef.current = onDismiss

  const stableDismiss = useCallback(() => {
    onDismissRef.current()
  }, [])

  useEffect(() => {
    if (!toast) return
    const id = window.setTimeout(stableDismiss, durationMs)
    return () => window.clearTimeout(id)
  }, [toast, durationMs, stableDismiss])

  useEffect(() => {
    if (!toast) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') stableDismiss()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [toast, stableDismiss])

  if (!toast || typeof document === 'undefined') {
    return null
  }

  const variant = toast.variant ?? 'success'

  return createPortal(
    <div className="toast-anchored" aria-live="polite">
      <div
        role="status"
        className={`toast toast--${variant}`}
      >
        {variant === 'success' ? (
          <span className="toast__icon toast__icon--ok" aria-hidden>
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2.25" strokeLinecap="round" strokeLinejoin="round">
              <path d="M20 6L9 17l-5-5" />
            </svg>
          </span>
        ) : (
          <span className="toast__icon toast__icon--err" aria-hidden>
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2.25" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="9" />
              <path d="M12 8v4.5" />
              <circle cx="12" cy="16.5" r="0.9" fill="currentColor" stroke="none" />
            </svg>
          </span>
        )}
        <span className="toast__msg">{toast.message}</span>
        <button
          type="button"
          className="toast__close"
          aria-label="Dismiss notification"
          onClick={stableDismiss}
        >
          ×
        </button>
      </div>
    </div>,
    document.body,
  )
}
