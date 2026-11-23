import { useCallback, useEffect, useRef, useState, type DragEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson, apiUpload } from '../api'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import { type GalleryItem } from '../lib/gallery'
import { useUnsavedChangesWarning } from '../lib/useUnsavedChangesWarning'

const MAX_BYTES = 15 * 1024 * 1024
const ACCEPT = 'image/jpeg,image/png,image/webp,image/*'

type Staged = {
  key: string
  file: File
  previewUrl: string
  status: 'ready' | 'uploading' | 'done' | 'error'
  progress: number
  error?: string
}

function isHeic(f: File): boolean {
  return /image\/hei[cf]/i.test(f.type) || /\.(heic|heif)$/i.test(f.name)
}

/** Client-side check before uploading; the server checks again. */
function precheck(f: File): string | undefined {
  if (isHeic(f)) {
    return 'HEIC photos can’t be read yet. On iPhone choose the photo from the Photos picker (it converts to JPEG) or set Camera → Formats → Most Compatible.'
  }
  if (f.type && !/^image\/(jpeg|png|webp)$/i.test(f.type)) return 'Please use a JPEG, PNG, or WebP image.'
  if (f.size > MAX_BYTES) return 'This file is larger than 15 MB. Scan at 300 dpi or take a photo instead.'
  return undefined
}

export function AdminGallery() {
  const [items, setItems] = useState<GalleryItem[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [staged, setStaged] = useState<Staged[]>([])
  const [uploading, setUploading] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const [busyId, setBusyId] = useState<number | null>(null)
  const fileInput = useRef<HTMLInputElement>(null)
  const cameraInput = useRef<HTMLInputElement>(null)

  const load = useCallback(async () => {
    try {
      const d = await apiJson<{ items: GalleryItem[] | null }>('/api/admin/gallery')
      setItems(Array.isArray(d.items) ? d.items : [])
      setErr(null)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load the gallery.')
      setItems((prev) => prev ?? [])
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  // Free preview object URLs when the page unmounts.
  const stagedRef = useRef(staged)
  stagedRef.current = staged
  useEffect(() => () => stagedRef.current.forEach((s) => URL.revokeObjectURL(s.previewUrl)), [])

  const pending = staged.filter((s) => s.status !== 'done')
  useUnsavedChangesWarning(pending.length > 0, 'You have drawings that haven’t been uploaded yet. Leave anyway?')

  function addFiles(list: FileList | File[] | null) {
    if (!list) return
    const files = Array.from(list)
    if (files.length === 0) return
    setStaged((prev) => [
      ...prev,
      ...files.map((file, i) => {
        const error = precheck(file)
        return {
          key: `${Date.now()}-${i}-${file.name}`,
          file,
          previewUrl: URL.createObjectURL(file),
          status: error ? ('error' as const) : ('ready' as const),
          progress: 0,
          error,
        }
      }),
    ])
  }

  function updateStaged(key: string, patch: Partial<Staged>) {
    setStaged((prev) => prev.map((s) => (s.key === key ? { ...s, ...patch } : s)))
  }

  function removeStaged(key: string) {
    setStaged((prev) => {
      const s = prev.find((x) => x.key === key)
      if (s) URL.revokeObjectURL(s.previewUrl)
      return prev.filter((x) => x.key !== key)
    })
  }

  function clearDone() {
    setStaged((prev) => {
      prev.filter((s) => s.status === 'done').forEach((s) => URL.revokeObjectURL(s.previewUrl))
      return prev.filter((s) => s.status !== 'done')
    })
  }

  async function uploadAll() {
    const queue = stagedRef.current.filter((s) => s.status === 'ready' || (s.status === 'error' && !precheck(s.file)))
    if (queue.length === 0) return
    setUploading(true)
    let ok = 0
    // One at a time: each file gets its own progress and error, and the server stays responsive.
    for (const s of queue) {
      updateStaged(s.key, { status: 'uploading', progress: 0, error: undefined })
      const form = new FormData()
      form.append('image', s.file)
      try {
        const res = await apiUpload<{ item?: GalleryItem }>('/api/admin/gallery', form, (p) =>
          updateStaged(s.key, { progress: p }),
        )
        updateStaged(s.key, { status: 'done', progress: 1 })
        if (res.item) setItems((prev) => [res.item as GalleryItem, ...(prev ?? [])])
        ok++
      } catch (e) {
        updateStaged(s.key, { status: 'error', error: e instanceof Error ? e.message : 'Upload failed.' })
      }
    }
    setUploading(false)
    await load()
    const failed = queue.length - ok
    setToast(
      failed === 0
        ? { message: `Uploaded ${ok} ${ok === 1 ? 'drawing' : 'drawings'}. They’re on the gallery now.`, variant: 'success' }
        : { message: `Uploaded ${ok}, ${failed} didn’t work. See the messages below.`, variant: 'error' },
    )
  }

  function onDrop(e: DragEvent) {
    e.preventDefault()
    setDragOver(false)
    addFiles(e.dataTransfer.files)
  }

  async function setHidden(it: GalleryItem, hidden: boolean) {
    setBusyId(it.id)
    try {
      await apiJson(`/api/admin/gallery/${it.id}/visibility`, { method: 'POST', json: { hidden } })
      await load()
      setToast({ message: hidden ? 'Hidden from the gallery.' : 'Shown on the gallery.', variant: 'success' })
    } catch (e) {
      setToast({ message: e instanceof Error ? e.message : 'Could not save.', variant: 'error' })
    } finally {
      setBusyId(null)
    }
  }

  async function remove(it: GalleryItem) {
    if (!window.confirm(`Delete this drawing permanently? This can’t be undone. (To take it off the site for now, use Hide instead.)`)) return
    setBusyId(it.id)
    try {
      await apiJson(`/api/admin/gallery/${it.id}/delete`, { method: 'POST' })
      setToast({ message: 'Drawing deleted.', variant: 'success' })
      await load()
    } catch (e) {
      setToast({ message: e instanceof Error ? e.message : 'Could not delete.', variant: 'error' })
    } finally {
      setBusyId(null)
    }
  }

  const readyCount = staged.filter((s) => s.status === 'ready').length
  const retryCount = staged.filter((s) => s.status === 'error' && !precheck(s.file)).length
  const doneCount = staged.filter((s) => s.status === 'done').length
  const hiddenCount = (items ?? []).filter((i) => i.hidden).length

  return (
    <main id="main">
      <PageHero
        titleId="admin-gallery-title"
        eyebrow="Leader Tools"
        title="Ward Gallery"
        lede="Scan or photograph drawings and share them with the ward."
        compact
      />
      <div className="wrap-wide admin-gallery-page">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={6000} />
        {err ? (
          <p className="panel-sites panel-sites--warn" role="alert">
            {err}
          </p>
        ) : null}

        <section className="admin-gallery-upload" aria-labelledby="gallery-upload-h">
          <h2 id="gallery-upload-h" className="moderate-section-title">
            Add Drawings
          </h2>
          <div
            className={`gallery-dropzone${dragOver ? ' is-over' : ''}`}
            onDragOver={(e) => {
              e.preventDefault()
              setDragOver(true)
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={onDrop}
          >
            <p className="gallery-dropzone__lead">
              <strong>Drag scans here</strong>, or choose them from your computer or phone. You can pick many at once.
            </p>
            <div className="gallery-dropzone__buttons">
              <button type="button" className="btn-primary" onClick={() => fileInput.current?.click()}>
                Choose Files
              </button>
              <button type="button" className="btn-secondary" onClick={() => cameraInput.current?.click()}>
                Take a Photo
              </button>
            </div>
            <p className="gallery-dropzone__hint">JPEG, PNG, or WebP up to 15 MB each. Big scans are resized automatically and location data is removed.</p>
            <input
              ref={fileInput}
              type="file"
              accept={ACCEPT}
              multiple
              hidden
              onChange={(e) => {
                addFiles(e.target.files)
                e.target.value = ''
              }}
            />
            <input
              ref={cameraInput}
              type="file"
              accept="image/*"
              capture="environment"
              hidden
              onChange={(e) => {
                addFiles(e.target.files)
                e.target.value = ''
              }}
            />
          </div>

          <p id="gallery-privacy-note" className="gallery-privacy-note">
            <strong>Privacy:</strong> the gallery shows only the pictures: no names, titles, or
            captions. Before scanning, cover or crop out any name written on a drawing (many kids sign
            their work).
          </p>

          {staged.length > 0 ? (
            <div className="gallery-staged">
              <div className="gallery-staged__bar">
                <p className="gallery-staged__summary">
                  {staged.length} selected
                  {doneCount ? ` · ${doneCount} uploaded` : ''}
                </p>
              </div>
              <ul className="gallery-staged__list" role="list">
                {staged.map((s) => (
                  <li key={s.key} className={`gallery-staged__item is-${s.status}`}>
                    <img src={s.previewUrl} alt="" className="gallery-staged__thumb" />
                    <div className="gallery-staged__body">
                      <p className="gallery-staged__file">
                        {s.file.size < 100 * 1024 ? `${Math.max(1, Math.round(s.file.size / 1024))} KB` : `${(s.file.size / 1024 / 1024).toFixed(1)} MB`}
                      </p>
                      {s.status === 'uploading' ? (
                        <div className="gallery-progress" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(s.progress * 100)} aria-label={`Uploading ${s.file.name}`}>
                          <span style={{ width: `${Math.max(4, s.progress * 100)}%` }} />
                        </div>
                      ) : null}
                      {s.status === 'done' ? <p className="gallery-staged__ok">✓ Uploaded</p> : null}
                      {s.status === 'error' ? (
                        <p className="gallery-staged__err" role="alert">
                          {s.error}
                        </p>
                      ) : null}
                    </div>
                    {s.status !== 'uploading' && s.status !== 'done' ? (
                      <button type="button" className="gallery-staged__remove" onClick={() => removeStaged(s.key)} aria-label={`Remove ${s.file.name}`}>
                        ×
                      </button>
                    ) : null}
                  </li>
                ))}
              </ul>
              <div className="gallery-staged__actions">
                {doneCount ? (
                  <button type="button" className="btn-secondary" onClick={clearDone} disabled={uploading}>
                    Clear Uploaded
                  </button>
                ) : null}
                <button
                  type="button"
                  className="btn-primary"
                  onClick={() => void uploadAll()}
                  disabled={uploading || readyCount + retryCount === 0}
                >
                  {uploading
                    ? 'Uploading…'
                    : retryCount && !readyCount
                      ? `Try Again (${retryCount})`
                      : `Upload ${readyCount + retryCount} ${readyCount + retryCount === 1 ? 'Drawing' : 'Drawings'}`}
                </button>
              </div>
            </div>
          ) : null}
        </section>

        <section className="admin-gallery-manage" aria-labelledby="gallery-manage-h">
          <h2 id="gallery-manage-h" className="moderate-section-title moderate-section-title--spaced">
            On the Gallery{items && items.length ? ` (${items.length})` : ''}
          </h2>
          <p className="panel-sites__note" style={{ marginTop: 0 }}>
            Newest first. The public page shows every drawing that isn&apos;t hidden, in a new random
            order each time it loads. Hidden drawings stay here but are off the public
            page{hiddenCount ? ` (${hiddenCount} hidden now)` : ''}. <Link to="/gallery">View the public gallery</Link>
          </p>
          {items === null ? (
            <p className="loading-note" role="status">
              <span className="loading-note__spinner" aria-hidden="true" />
              Loading…
            </p>
          ) : items.length === 0 ? (
            <div className="panel-sites">
              <p className="lead" style={{ margin: 0 }}>
                No drawings yet. Add the first ones above.
              </p>
            </div>
          ) : (
            <ul className="admin-gallery-list" role="list">
              {items.map((it) => {
                const busy = busyId === it.id
                return (
                  <li key={it.id} className={`admin-gallery-item${it.hidden ? ' is-hidden' : ''}`}>
                    <a href={it.imageUrl} target="_blank" rel="noopener noreferrer" className="admin-gallery-item__thumb" aria-label="Open full size">
                      <img src={it.thumbUrl} alt={it.alt} width={it.thumbWidth} height={it.thumbHeight} loading="lazy" />
                      {it.hidden ? <span className="gallery-badge gallery-badge--hidden">Hidden</span> : null}
                    </a>
                    <p className="admin-gallery-item__when">Added {it.when}</p>
                    <div className="admin-gallery-item__actions" role="group" aria-label={`Actions for drawing added ${it.when}`}>
                      <button type="button" className="btn-secondary" disabled={busy} onClick={() => void setHidden(it, !it.hidden)}>
                        {it.hidden ? 'Show' : 'Hide'}
                      </button>
                      <button type="button" className="btn-danger" disabled={busy} onClick={() => void remove(it)}>
                        Delete
                      </button>
                    </div>
                  </li>
                )
              })}
            </ul>
          )}
        </section>
      </div>
    </main>
  )
}
