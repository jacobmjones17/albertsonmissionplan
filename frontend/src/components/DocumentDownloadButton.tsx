import { useState } from 'react'
import { downloadTextFile } from '../lib/downloadDocument'

type DocumentDownloadButtonProps = {
  /** Visible button text, e.g. "Download printable goals". */
  label: string
  filename: string
  /** Returns the full HTML document to save. */
  buildDocument: () => Promise<string>
  disabled?: boolean
}

function DownloadIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true" focusable="false">
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

/** Labeled download button — saves a file in place without navigating away. */
export function DocumentDownloadButton({
  label,
  filename,
  buildDocument,
  disabled = false,
}: DocumentDownloadButtonProps) {
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
        className="btn btn-secondary document-download__btn"
        disabled={busy || disabled}
        aria-busy={busy || undefined}
        onClick={() => void handleDownload()}
      >
        <DownloadIcon />
        <span>{busy ? 'Preparing download…' : label}</span>
      </button>
      {err ? (
        <p className="document-download__err" role="alert">
          {err}
        </p>
      ) : null}
    </div>
  )
}
