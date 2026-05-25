import { useState } from 'react'
import { downloadTextFile } from '../lib/downloadDocument'

type DocumentDownloadIconProps = {
  /** Accessible label, e.g. "Download monthly challenges". */
  label: string
  filename: string
  /** Returns the full HTML document to save. */
  buildDocument: () => Promise<string>
  disabled?: boolean
}

function DownloadIcon() {
  return (
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" aria-hidden="true">
      <path
        d="M12 3v10m0 0 4-4m-4 4-4-4M5 17v2a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-2"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

/** Small download control — saves a file in place without navigating away. */
export function DocumentDownloadIcon({
  label,
  filename,
  buildDocument,
  disabled = false,
}: DocumentDownloadIconProps) {
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  async function handleDownload() {
    setBusy(true)
    setErr(null)
    try {
      const html = await buildDocument()
      downloadTextFile(html, filename)
    } catch {
      setErr('Download failed. Please try again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="document-download">
      <button
        type="button"
        className="document-download__btn"
        aria-label={label}
        title={label}
        disabled={busy || disabled}
        onClick={() => void handleDownload()}
      >
        <DownloadIcon />
      </button>
      {err ? (
        <p className="document-download__err" role="alert">
          {err}
        </p>
      ) : null}
    </div>
  )
}
