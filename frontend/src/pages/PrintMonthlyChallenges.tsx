import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { MonthlyChallengesSheet } from '../components/MonthlyChallengesSheet'
import { CHALLENGES_PRINT_HINT, CHALLENGES_SHEET_STYLES } from '../lib/challengesSheetStyles'
import { downloadTextFile } from '../lib/downloadDocument'
import { buildOfflineMonthlyChallengesHtml } from '../lib/downloadMonthlyChallenges'

const SHEET_STYLE_ID = 'challenges-sheet-styles'

export function PrintMonthlyChallenges() {
  const year = useMemo(() => new Date().getFullYear(), [])
  const filename = `Albertson-Ward-Monthly-Challenges-${year}.html`
  const [downloading, setDownloading] = useState(false)

  useEffect(() => {
    document.title = `Monthly Challenges — Albertson Ward Plan`
  }, [])

  useEffect(() => {
    if (document.getElementById(SHEET_STYLE_ID)) return
    const el = document.createElement('style')
    el.id = SHEET_STYLE_ID
    el.textContent = CHALLENGES_SHEET_STYLES
    document.head.appendChild(el)
    return () => {
      el.remove()
    }
  }, [])

  async function handleDownload() {
    setDownloading(true)
    try {
      const html = await buildOfflineMonthlyChallengesHtml(year)
      downloadTextFile(html, filename)
    } finally {
      setDownloading(false)
    }
  }

  return (
    <div className="print-doc print-doc--challenges-sheet">
      <div className="print-doc__toolbar no-print">
        <Link to="/monthly-challenges" className="print-doc__back">
          ← Back
        </Link>
        <p className="print-doc__print-note">
          In print settings, choose <strong>Landscape</strong>, paper size <strong>Letter</strong>, and scale{' '}
          <strong>100%</strong>.
        </p>
        <div className="print-doc__actions">
          <button type="button" className="btn btn-secondary" onClick={() => window.print()}>
            Print
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            disabled={downloading}
            onClick={() => void handleDownload()}
          >
            {downloading ? 'Preparing…' : 'Download'}
          </button>
        </div>
      </div>

      <article className="print-doc__body print-doc__body--sheet">
        <p className="sheet-print-hint" dangerouslySetInnerHTML={{ __html: CHALLENGES_PRINT_HINT }} />
        <MonthlyChallengesSheet year={year} />
      </article>
    </div>
  )
}
