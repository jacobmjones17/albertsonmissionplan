import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { WardGoalLine } from '../components/richtext/WardGoalLine'
import { WARD_PLAN_PRINT_HINT, WARD_PLAN_STYLES } from '../lib/wardPlanDocumentStyles'

const SHEET_STYLE_ID = 'ward-plan-sheet-styles'

export function PrintWardPlan() {
  const [wardGoals, setWardGoals] = useState<string[] | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    if (document.getElementById(SHEET_STYLE_ID)) return
    const el = document.createElement('style')
    el.id = SHEET_STYLE_ID
    el.textContent = WARD_PLAN_STYLES
    document.head.appendChild(el)
    return () => {
      el.remove()
    }
  }, [])

  useEffect(() => {
    apiJson<{ wardGoals: string[] }>('/api/ward-plan')
      .then((d) => setWardGoals(d.wardGoals ?? []))
      .catch(() => setErr('Could not load the ward goals. Please go back and try again.'))
  }, [])

  useEffect(() => {
    if (!wardGoals || wardGoals.length === 0) return
    const id = window.setTimeout(() => {
      window.print()
    }, 400)
    return () => window.clearTimeout(id)
  }, [wardGoals])

  const ready = Boolean(wardGoals && wardGoals.length > 0)

  return (
    <div className="print-doc print-doc--ward-plan">
      <div className="print-doc__toolbar no-print">
        <Link to="/ward-plan" className="print-doc__back">
          ← Back to Ward Plan
        </Link>
        <p className="print-doc__print-note">
          Use <strong>Print / Save as PDF</strong> to open your device&apos;s print sheet. Choose{' '}
          <strong>Letter</strong> paper and turn on <strong>Background graphics</strong> if you want
          the cream page and gold accents.
        </p>
        <div className="print-doc__actions">
          <button
            type="button"
            className="btn btn-primary"
            disabled={!ready}
            onClick={() => window.print()}
          >
            Print / Save as PDF
          </button>
        </div>
      </div>

      <p className="print-doc__small-screen-note no-print" role="note">
        Tap <strong>Print / Save as PDF</strong> to open your phone&apos;s print sheet, where you can
        print or save a PDF. This page is a preview of what will print.
      </p>

      <article className="print-doc__body print-doc__body--ward-plan">
        <p className="sheet-print-hint" dangerouslySetInnerHTML={{ __html: WARD_PLAN_PRINT_HINT }} />
        {err ? (
          <p className="flash" role="alert">
            {err}
          </p>
        ) : !wardGoals ? (
          <p className="loading-note" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading the ward goals…
          </p>
        ) : (
          <div className="ward-goals-sheet">
            <header className="ward-goals-sheet__header">
              <p className="ward-goals-sheet__eyebrow">Albertson Ward</p>
              <h1 className="ward-goals-sheet__title">Ward Goals</h1>
              <p className="ward-goals-sheet__lead">
                <strong>We will strengthen our ward by:</strong>
              </p>
            </header>
            <ul className="ward-goals-sheet__list">
              {wardGoals.map((g, i) => (
                <li key={i}>
                  <WardGoalLine content={g} />
                </li>
              ))}
            </ul>
            <blockquote className="ward-goals-sheet__quote">
              Matthew 28:19 — &ldquo;Go ye therefore, and teach all nations, baptizing them in the
              name of the Father, and of the Son, and of the Holy Ghost.&rdquo;
            </blockquote>
          </div>
        )}
      </article>
    </div>
  )
}
