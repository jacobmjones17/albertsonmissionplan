import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { apiJson, apiUpload } from '../api'
import { ChallengeCard } from '../components/ChallengeCard'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import {
  fromApiItem,
  invalidateMonthlyChallenges,
  shortTheme,
  type ChallengeApiItem,
  type ChallengeBullet,
} from '../lib/monthlyChallenges'
import { useUnsavedChangesWarning } from '../lib/useUnsavedChangesWarning'
import { MONTHS_LONG, MONTHS_SHORT } from '../monthlyData'

type AdminChallenge = ChallengeApiItem & {
  hidden: boolean
  isOriginalText: boolean
  isOriginalImage: boolean
  updatedAt?: string
  updatedBy?: string
  original: { heading: string; goal: string; bullets: ChallengeBullet[] }
}

type IdeaRow = { key: string; who: string; text: string }

type Draft = { heading: string; goal: string; ideas: IdeaRow[]; hidden: boolean }

type PendingImage = { file: File; url: string }

type SaveStatus =
  | { kind: 'idle' }
  | { kind: 'saving'; label: string }
  | { kind: 'saved'; at: Date }
  | { kind: 'error'; message: string }

/** Limits match the server (backend/internal/handlers/challenges.go). */
const LIMITS = { heading: 60, goal: 200, ideas: 8, who: 30, text: 400 }
const MAX_BYTES = 15 * 1024 * 1024
const WHO_SUGGESTIONS = ['Individuals', 'Couples', 'Families', 'Youth', 'Children', 'Ward']

let rowSeq = 0
const newKey = () => `idea-${++rowSeq}`

function toDraft(c: AdminChallenge): Draft {
  return {
    heading: c.heading,
    goal: c.goal,
    hidden: c.hidden,
    ideas: (c.bullets ?? []).map((b) => ({ key: newKey(), who: b.who, text: b.text })),
  }
}

const tidy = (s: string) => s.replace(/\s+/g, ' ').trim()

function cleanIdeas(ideas: IdeaRow[]): ChallengeBullet[] {
  return ideas
    .map((r) => ({ who: tidy(r.who), text: tidy(r.text) }))
    .filter((b) => b.who !== '' || b.text !== '')
}

function sameDraft(a: Draft, b: Draft): boolean {
  return (
    tidy(a.heading) === tidy(b.heading) &&
    tidy(a.goal) === tidy(b.goal) &&
    a.hidden === b.hidden &&
    JSON.stringify(cleanIdeas(a.ideas)) === JSON.stringify(cleanIdeas(b.ideas))
  )
}

/** Friendly problems with a draft, or [] when it can be saved. */
function problemsFor(d: Draft): string[] {
  const out: string[] = []
  const heading = tidy(d.heading)
  const goal = tidy(d.goal)
  if (!heading) out.push('Give the challenge a name, like “Gratitude Challenge”.')
  else if (heading.length > LIMITS.heading) out.push(`The challenge name can be up to ${LIMITS.heading} characters.`)
  if (!goal) out.push('Add the goal for this month.')
  else if (goal.length > LIMITS.goal) out.push(`The goal can be up to ${LIMITS.goal} characters.`)
  const ideas = cleanIdeas(d.ideas)
  if (ideas.length === 0) out.push('Add at least one idea (for example, one for Individuals).')
  if (ideas.length > LIMITS.ideas) out.push(`A challenge can have up to ${LIMITS.ideas} ideas.`)
  for (const b of ideas) {
    if (!b.who) out.push('Each idea needs a “who” label, like Individuals, Families, or Ward.')
    else if (!b.text) out.push(`The idea for ${b.who} is empty. Fill it in or remove that row.`)
    if (b.who.length > LIMITS.who) out.push(`“Who” labels can be up to ${LIMITS.who} characters.`)
    if (b.text.length > LIMITS.text) out.push(`Each idea can be up to ${LIMITS.text} characters.`)
  }
  return [...new Set(out)]
}

function precheckImage(f: File): string | undefined {
  if (/image\/hei[cf]/i.test(f.type) || /\.(heic|heif)$/i.test(f.name)) {
    return 'HEIC photos can’t be read yet. On iPhone choose the photo from the Photos picker (it converts to JPEG) or set Camera → Formats → Most Compatible.'
  }
  if (f.type && !/^image\/(jpeg|png|webp)$/i.test(f.type)) return 'Please use a JPEG, PNG, or WebP picture.'
  if (f.size > MAX_BYTES) return 'This picture is larger than 15 MB. Try a smaller photo.'
  return undefined
}

function formatTime(d: Date): string {
  return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
}

const monthName = (m: number) => MONTHS_LONG[m - 1] ?? `Month ${m}`

export function AdminChallenges() {
  const [items, setItems] = useState<AdminChallenge[] | null>(null)
  const [drafts, setDrafts] = useState<Record<number, Draft>>({})
  const [images, setImages] = useState<Record<number, PendingImage>>({})
  const [selected, setSelected] = useState(() => new Date().getMonth() + 1)
  const [err, setErr] = useState<string | null>(null)
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [status, setStatus] = useState<SaveStatus>({ kind: 'idle' })
  const [busy, setBusy] = useState(false)
  const [showProblems, setShowProblems] = useState(false)
  const [linkFor, setLinkFor] = useState<string | null>(null)
  const [linkText, setLinkText] = useState('')
  const [linkUrl, setLinkUrl] = useState('')
  const [moveTo, setMoveTo] = useState('')
  const fileInput = useRef<HTMLInputElement>(null)
  const imagesRef = useRef(images)
  imagesRef.current = images

  /** Apply a fresh server list; keep unsaved drafts for months listed in `keep`. */
  const applyItems = useCallback((list: AdminChallenge[], keep: Set<number> = new Set()) => {
    setItems(list)
    setDrafts((prev) => {
      const next: Record<number, Draft> = {}
      for (const c of list) next[c.month] = keep.has(c.month) && prev[c.month] ? prev[c.month]! : toDraft(c)
      return next
    })
    // Public pages re-fetch so they show the change right away.
    invalidateMonthlyChallenges()
  }, [])

  const load = useCallback(async () => {
    try {
      const d = await apiJson<{ items: AdminChallenge[] | null }>('/api/admin/challenges')
      applyItems(d.items ?? [])
      setErr(null)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load the monthly challenges.')
      setItems((prev) => prev ?? [])
    }
  }, [applyItems])

  useEffect(() => {
    void load()
  }, [load])

  // Free local picture previews when leaving the page.
  useEffect(
    () => () => {
      for (const p of Object.values(imagesRef.current)) URL.revokeObjectURL(p.url)
    },
    [],
  )

  const byMonth = useMemo(() => new Map((items ?? []).map((c) => [c.month, c])), [items])
  const isDirty = useCallback(
    (m: number) => {
      const c = byMonth.get(m)
      const d = drafts[m]
      return Boolean(images[m]) || Boolean(c && d && !sameDraft(d, toDraft(c)))
    },
    [byMonth, drafts, images],
  )
  const dirtyMonths = useMemo(() => (items ?? []).map((c) => c.month).filter(isDirty), [items, isDirty])
  const dirty = dirtyMonths.length > 0
  useUnsavedChangesWarning(dirty)

  const current = byMonth.get(selected)
  const draft = drafts[selected]
  const pendingImage = images[selected]

  function updateDraft(patch: Partial<Draft>) {
    setDrafts((prev) => ({ ...prev, [selected]: { ...prev[selected]!, ...patch } }))
    if (status.kind === 'saved' || status.kind === 'error') setStatus({ kind: 'idle' })
  }

  function updateIdea(key: string, patch: Partial<IdeaRow>) {
    if (!draft) return
    updateDraft({ ideas: draft.ideas.map((r) => (r.key === key ? { ...r, ...patch } : r)) })
  }

  function moveIdea(key: string, delta: -1 | 1) {
    if (!draft) return
    const i = draft.ideas.findIndex((r) => r.key === key)
    const j = i + delta
    if (i < 0 || j < 0 || j >= draft.ideas.length) return
    const ideas = [...draft.ideas]
    ;[ideas[i], ideas[j]] = [ideas[j]!, ideas[i]!]
    updateDraft({ ideas })
  }

  function removeIdea(key: string) {
    if (!draft) return
    updateDraft({ ideas: draft.ideas.filter((r) => r.key !== key) })
  }

  function addIdea() {
    if (!draft) return
    const used = new Set(draft.ideas.map((r) => r.who))
    const who = ['Individuals', 'Families', 'Ward'].find((w) => !used.has(w)) ?? ''
    updateDraft({ ideas: [...draft.ideas, { key: newKey(), who, text: '' }] })
  }

  function openLinkForm(key: string) {
    setLinkFor(key)
    setLinkText('')
    setLinkUrl('')
  }

  function insertLink(key: string) {
    const label = tidy(linkText)
    let url = linkUrl.trim()
    if (!label || !url) return
    if (!/^https?:\/\//i.test(url)) url = `https://${url}`
    url = url.replace(/^http:\/\//i, 'https://').replace(/[()\s]/g, (ch) => `%${ch.charCodeAt(0).toString(16).toUpperCase().padStart(2, '0')}`)
    const row = draft?.ideas.find((r) => r.key === key)
    if (!row) return
    const sep = row.text && !row.text.endsWith(' ') ? ' ' : ''
    updateIdea(key, { text: `${row.text}${sep}[${label.replace(/[[\]]/g, '')}](${url})` })
    setLinkFor(null)
  }

  function chooseImage(f: File | undefined) {
    if (!f) return
    const problem = precheckImage(f)
    if (problem) {
      setToast({ message: problem, variant: 'error' })
      return
    }
    setImages((prev) => {
      if (prev[selected]) URL.revokeObjectURL(prev[selected]!.url)
      return { ...prev, [selected]: { file: f, url: URL.createObjectURL(f) } }
    })
    if (status.kind !== 'saving') setStatus({ kind: 'idle' })
  }

  function clearPendingImage(m: number) {
    setImages((prev) => {
      if (!prev[m]) return prev
      URL.revokeObjectURL(prev[m]!.url)
      const next = { ...prev }
      delete next[m]
      return next
    })
  }

  function discardMonth(m: number) {
    const c = byMonth.get(m)
    if (!c) return
    setDrafts((prev) => ({ ...prev, [m]: toDraft(c) }))
    clearPendingImage(m)
    setShowProblems(false)
  }

  async function onSave(e?: FormEvent) {
    e?.preventDefault()
    setToast(null)
    const blocking = dirtyMonths.filter((m) => drafts[m] && problemsFor(drafts[m]!).length > 0)
    if (blocking.length > 0) {
      setShowProblems(true)
      if (!blocking.includes(selected)) setSelected(blocking[0]!)
      setStatus({ kind: 'error', message: `Fix the highlighted problems in ${blocking.map(monthName).join(', ')} first.` })
      return
    }
    const saving = [...dirtyMonths]
    const remaining = new Set(saving)
    let latest: AdminChallenge[] | null = null
    try {
      for (const m of saving) {
        setStatus({ kind: 'saving', label: monthName(m) })
        const img = images[m]
        if (img) {
          const form = new FormData()
          form.append('image', img.file)
          const r = await apiUpload<{ items: AdminChallenge[] }>(`/api/admin/challenges/${m}/image`, form)
          latest = r.items
          clearPendingImage(m)
        }
        const d = drafts[m]!
        const c = byMonth.get(m)
        if (c && !sameDraft(d, toDraft(c))) {
          const r = await apiJson<{ items: AdminChallenge[] }>(`/api/admin/challenges/${m}`, {
            method: 'POST',
            json: { heading: tidy(d.heading), goal: tidy(d.goal), hidden: d.hidden, bullets: cleanIdeas(d.ideas) },
          })
          latest = r.items
        }
        remaining.delete(m)
      }
      if (latest) applyItems(latest)
      setShowProblems(false)
      setStatus({ kind: 'saved', at: new Date() })
      setToast({ message: 'Saved. The Monthly Challenges page and printable sheet now show your changes.', variant: 'success' })
    } catch (ex) {
      const message = ex instanceof Error ? ex.message : 'Could not save.'
      if (latest) applyItems(latest, remaining)
      setStatus({ kind: 'error', message })
      setToast({ message: `Not saved: ${message}`, variant: 'error' })
    }
  }

  async function resetMonth(part: 'all' | 'image') {
    if (!current) return
    const name = monthName(current.month)
    const question =
      part === 'all'
        ? `Reset ${name} to the original challenge?\n\nThis puts back the original name, goal, ideas, and picture, and shows ${name} on the site. Your edits for ${name} will be lost.`
        : `Put back the original picture for ${name}?`
    if (!window.confirm(question)) return
    setBusy(true)
    try {
      const r = await apiJson<{ items: AdminChallenge[] }>(`/api/admin/challenges/${current.month}/reset`, {
        method: 'POST',
        json: { part },
      })
      clearPendingImage(current.month)
      const keep = new Set(dirtyMonths.filter((m) => m !== current.month))
      if (part === 'image' && drafts[current.month]) keep.add(current.month) // keep unsaved text edits
      applyItems(r.items, keep)
      setToast({ message: part === 'all' ? `${name} is back to the original challenge.` : `${name} has its original picture again.`, variant: 'success' })
    } catch (ex) {
      setToast({ message: ex instanceof Error ? ex.message : 'Could not reset.', variant: 'error' })
    } finally {
      setBusy(false)
    }
  }

  async function moveChallenge() {
    const to = Number(moveTo)
    if (!current || !to || to === current.month) return
    if (isDirty(current.month) || isDirty(to)) {
      setToast({ message: `Save or discard your changes to ${monthName(current.month)} and ${monthName(to)} before moving.`, variant: 'error' })
      return
    }
    const other = byMonth.get(to)
    if (!window.confirm(`Move “${current.heading}” to ${monthName(to)}?\n\n“${other?.heading ?? 'That month’s challenge'}” will move to ${monthName(current.month)}.`)) return
    setBusy(true)
    try {
      const r = await apiJson<{ items: AdminChallenge[] }>(`/api/admin/challenges/${current.month}/move`, {
        method: 'POST',
        json: { to },
      })
      applyItems(r.items, new Set(dirtyMonths))
      setMoveTo('')
      setSelected(to)
      setToast({ message: `Swapped ${monthName(current.month)} and ${monthName(to)}.`, variant: 'success' })
    } catch (ex) {
      setToast({ message: ex instanceof Error ? ex.message : 'Could not move the challenge.', variant: 'error' })
    } finally {
      setBusy(false)
    }
  }

  const problems = draft ? problemsFor(draft) : []
  const preview =
    current && draft
      ? {
          ...fromApiItem(current),
          heading: tidy(draft.heading) || current.heading,
          goal: tidy(draft.goal),
          bullets: cleanIdeas(draft.ideas),
        }
      : null
  const thisMonth = new Date().getMonth() + 1
  const saving = status.kind === 'saving'

  return (
    <main id="main">
      <PageHero
        titleId="challenges-edit-title"
        eyebrow="Leader Tools"
        title="Edit Monthly Challenges"
        lede="Pick a month, change its challenge, ideas, or picture, and check the preview. Save publishes your changes."
        compact
      />
      <div className="wrap-wide admin-challenges-page">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={6000} />
        {err ? (
          <p className="panel-sites panel-sites--warn" role="alert">
            {err}
          </p>
        ) : null}
        {items === null ? (
          <p className="loading-note" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading the monthly challenges…
          </p>
        ) : null}

        {items && items.length > 0 ? (
          <form className="admin-challenges" onSubmit={(e) => void onSave(e)} noValidate>
            <nav className="admin-challenges__months" aria-label="Months">
              <label className="admin-challenges__month-select">
                <span className="ward-plan-tabs-label">Month</span>
                <select value={selected} onChange={(e) => setSelected(Number(e.target.value))}>
                  {items.map((c) => (
                    <option key={c.month} value={c.month}>
                      {monthName(c.month)} · {shortTheme(drafts[c.month]?.heading || c.heading)}
                      {isDirty(c.month) ? ' • unsaved' : drafts[c.month]?.hidden ? ' (hidden)' : ''}
                    </option>
                  ))}
                </select>
              </label>
              <p className="ward-plan-tabs-label admin-challenges__months-label">Months</p>
              <ul className="admin-challenges__month-list" role="list">
                {items.map((c) => {
                  const d = drafts[c.month]
                  const hidden = d?.hidden ?? c.hidden
                  const edited = !c.isOriginalText || !c.isOriginalImage
                  return (
                    <li key={c.month}>
                      <button
                        type="button"
                        className={`admin-challenges__month${selected === c.month ? ' is-active' : ''}${hidden ? ' is-hidden' : ''}`}
                        aria-current={selected === c.month ? 'true' : undefined}
                        onClick={() => setSelected(c.month)}
                      >
                        <span className="admin-challenges__month-abbr">{MONTHS_SHORT[c.month - 1]}</span>
                        <span className="admin-challenges__month-name">{shortTheme(d?.heading || c.heading)}</span>
                        <span className="admin-challenges__month-tags">
                          {c.month === thisMonth ? <span className="challenge-tag challenge-tag--now">This month</span> : null}
                          {hidden ? <span className="challenge-tag challenge-tag--hidden">Hidden</span> : null}
                          {edited ? <span className="challenge-tag">Edited</span> : null}
                        </span>
                        {isDirty(c.month) ? (
                          <>
                            <span className="ward-plan-tabs__dot" aria-hidden="true" />
                            <span className="sr-only"> (unsaved changes)</span>
                          </>
                        ) : null}
                      </button>
                    </li>
                  )
                })}
              </ul>
            </nav>

            {current && draft && preview ? (
              <div className="admin-challenges__editor">
                <div className="admin-challenges__fields form-stack">
                  <div className="admin-challenges__editor-head">
                    <h2 className="admin-challenges__editor-title">{monthName(current.month)}</h2>
                    {current.updatedAt ? (
                      <p className="admin-challenges__meta">
                        Last changed {current.updatedAt}
                        {current.updatedBy ? ` by ${current.updatedBy}` : ''}
                      </p>
                    ) : (
                      <p className="admin-challenges__meta">Original challenge</p>
                    )}
                  </div>

                  {showProblems && problems.length > 0 ? (
                    <div className="panel-sites panel-sites--warn admin-challenges__problems" role="alert">
                      <p>Please fix before saving:</p>
                      <ul>
                        {problems.map((p) => (
                          <li key={p}>{p}</li>
                        ))}
                      </ul>
                    </div>
                  ) : null}

                  <label>
                    Challenge name
                    <input
                      type="text"
                      value={draft.heading}
                      maxLength={LIMITS.heading + 20}
                      placeholder="Gratitude Challenge"
                      onChange={(e) => updateDraft({ heading: e.target.value })}
                      aria-describedby="ch-heading-hint"
                    />
                    <span id="ch-heading-hint" className="admin-challenges__hint">
                      The month button shows “{shortTheme(tidy(draft.heading)) || '…'}”. Ending the name with
                      “Challenge” is optional.
                    </span>
                  </label>

                  <label>
                    Goal
                    <textarea
                      rows={2}
                      value={draft.goal}
                      maxLength={LIMITS.goal + 20}
                      placeholder="Cultivate an attitude of gratitude"
                      onChange={(e) => updateDraft({ goal: e.target.value })}
                    />
                  </label>

                  <fieldset className="admin-challenges__ideas">
                    <legend>Ideas</legend>
                    <p className="admin-challenges__hint">
                      One idea for each group, like Individuals, Families, and Ward. Use the arrows to change
                      the order.
                    </p>
                    <datalist id="challenge-who-options">
                      {WHO_SUGGESTIONS.map((w) => (
                        <option key={w} value={w} />
                      ))}
                    </datalist>
                    <ol className="admin-challenges__idea-list">
                      {draft.ideas.map((row, i) => (
                        <li key={row.key} className="admin-challenges__idea">
                          <div className="admin-challenges__idea-fields">
                            <label className="admin-challenges__who">
                              Who
                              <input
                                type="text"
                                list="challenge-who-options"
                                value={row.who}
                                maxLength={LIMITS.who + 10}
                                placeholder="Individuals"
                                onChange={(e) => updateIdea(row.key, { who: e.target.value })}
                              />
                            </label>
                            <label className="admin-challenges__what">
                              Idea
                              <textarea
                                rows={3}
                                value={row.text}
                                maxLength={LIMITS.text + 40}
                                placeholder="Keep a daily gratitude journal."
                                onChange={(e) => updateIdea(row.key, { text: e.target.value })}
                              />
                            </label>
                          </div>
                          {linkFor === row.key ? (
                            <div className="admin-challenges__link-form" role="group" aria-label="Add a link">
                              <label>
                                Words to show
                                <input
                                  type="text"
                                  value={linkText}
                                  placeholder="FamilySearch.org"
                                  onChange={(e) => setLinkText(e.target.value)}
                                  autoFocus
                                />
                              </label>
                              <label>
                                Web address
                                <input
                                  type="text"
                                  inputMode="url"
                                  value={linkUrl}
                                  placeholder="https://www.familysearch.org"
                                  onChange={(e) => setLinkUrl(e.target.value)}
                                  onKeyDown={(e) => {
                                    if (e.key === 'Enter') {
                                      e.preventDefault()
                                      insertLink(row.key)
                                    }
                                  }}
                                />
                              </label>
                              <div className="admin-challenges__link-actions">
                                <button
                                  type="button"
                                  className="btn-primary"
                                  disabled={!tidy(linkText) || !linkUrl.trim()}
                                  onClick={() => insertLink(row.key)}
                                >
                                  Add Link
                                </button>
                                <button type="button" className="btn-secondary" onClick={() => setLinkFor(null)}>
                                  Cancel
                                </button>
                              </div>
                            </div>
                          ) : null}
                          <div className="admin-challenges__idea-actions" role="group" aria-label={`Idea ${i + 1} options`}>
                            <button type="button" className="admin-challenges__icon-btn" onClick={() => moveIdea(row.key, -1)} disabled={i === 0} aria-label="Move up" title="Move up">
                              ↑
                            </button>
                            <button
                              type="button"
                              className="admin-challenges__icon-btn"
                              onClick={() => moveIdea(row.key, 1)}
                              disabled={i === draft.ideas.length - 1}
                              aria-label="Move down"
                              title="Move down"
                            >
                              ↓
                            </button>
                            <button type="button" className="admin-challenges__text-btn" onClick={() => openLinkForm(row.key)}>
                              Add a Link
                            </button>
                            <button type="button" className="admin-challenges__text-btn admin-challenges__text-btn--danger" onClick={() => removeIdea(row.key)}>
                              Remove
                            </button>
                          </div>
                        </li>
                      ))}
                    </ol>
                    {draft.ideas.length < LIMITS.ideas ? (
                      <button type="button" className="btn-secondary admin-challenges__add" onClick={addIdea}>
                        + Add Another Idea
                      </button>
                    ) : null}
                    <p className="admin-challenges__hint">
                      Tip: links show up as underlined words. To make words bold, put two stars on each side,
                      like **this**.
                    </p>
                  </fieldset>

                  <fieldset className="admin-challenges__picture">
                    <legend>Picture</legend>
                    <div className="admin-challenges__picture-row">
                      <img
                        src={pendingImage?.url ?? current.thumbUrl}
                        alt=""
                        className="admin-challenges__picture-thumb"
                        width={240}
                        height={160}
                      />
                      <div className="admin-challenges__picture-body">
                        <p className="admin-challenges__hint">
                          {pendingImage
                            ? `New picture chosen: ${pendingImage.file.name}. It uploads when you save.`
                            : current.isOriginalImage
                              ? 'This is the original picture.'
                              : 'A leader replaced this picture.'}{' '}
                          Pictures are trimmed to a wide 3:2 shape, so keep the important part in the middle.
                        </p>
                        <div className="admin-challenges__picture-actions">
                          <button type="button" className="btn-secondary" onClick={() => fileInput.current?.click()}>
                            {pendingImage ? 'Choose a Different Picture' : 'Replace Picture'}
                          </button>
                          {pendingImage ? (
                            <button type="button" className="admin-challenges__text-btn" onClick={() => clearPendingImage(selected)}>
                              Keep the Current Picture
                            </button>
                          ) : !current.isOriginalImage ? (
                            <button type="button" className="admin-challenges__text-btn" disabled={busy} onClick={() => void resetMonth('image')}>
                              Use the Original Picture
                            </button>
                          ) : null}
                        </div>
                        <input
                          ref={fileInput}
                          type="file"
                          accept="image/jpeg,image/png,image/webp,image/*"
                          hidden
                          onChange={(e) => {
                            chooseImage(e.target.files?.[0])
                            e.target.value = ''
                          }}
                        />
                      </div>
                    </div>
                  </fieldset>

                  <label className="admin-challenges__toggle">
                    <input
                      type="checkbox"
                      checked={!draft.hidden}
                      onChange={(e) => updateDraft({ hidden: !e.target.checked })}
                    />
                    <span>
                      Show {monthName(current.month)} on the public site
                      <span className="admin-challenges__hint">
                        Unchecked months are hidden from the Monthly Challenges page and the printable sheet.
                      </span>
                    </span>
                  </label>

                  <details className="admin-challenges__more">
                    <summary>More options</summary>
                    <div className="admin-challenges__more-body">
                      <div className="admin-challenges__more-item">
                        <label>
                          Move this challenge to another month
                          <select value={moveTo} onChange={(e) => setMoveTo(e.target.value)}>
                            <option value="">Choose a month…</option>
                            {items
                              .filter((c) => c.month !== current.month)
                              .map((c) => (
                                <option key={c.month} value={c.month}>
                                  {monthName(c.month)} (swaps with {shortTheme(c.heading)})
                                </option>
                              ))}
                          </select>
                        </label>
                        <button type="button" className="btn-secondary" disabled={!moveTo || busy} onClick={() => void moveChallenge()}>
                          Move
                        </button>
                      </div>
                      <div className="admin-challenges__more-item">
                        <p className="admin-challenges__hint">
                          Undo every change to {monthName(current.month)} and go back to the original “
                          {current.original.heading}” challenge and picture.
                        </p>
                        <button
                          type="button"
                          className="btn-danger"
                          disabled={busy || (current.isOriginalText && current.isOriginalImage && !current.hidden && !isDirty(current.month))}
                          onClick={() => void resetMonth('all')}
                        >
                          Reset to Original
                        </button>
                      </div>
                    </div>
                  </details>
                </div>

                <section className="admin-challenges__preview" aria-labelledby="challenge-preview-title">
                  <h3 id="challenge-preview-title" className="ward-plan-preview__title">
                    Preview <span>how {monthName(current.month)} looks on the public page</span>
                  </h3>
                  {draft.hidden ? (
                    <p className="admin-challenges__hidden-note">Hidden: members won’t see this month.</p>
                  ) : null}
                  <div className="admin-challenges__preview-card">
                    <ChallengeCard
                      challenge={preview}
                      isCurrent={current.month === thisMonth}
                      titleId="challenge-preview-card-title"
                      imageSrcOverride={pendingImage?.url}
                    />
                  </div>
                  <p className="challenge-month-tile admin-challenges__tile-preview" aria-hidden="true">
                    <span className="challenge-month-tile__abbr">{MONTHS_SHORT[current.month - 1]}</span>
                    <span className="challenge-month-tile__theme">{shortTheme(preview.heading)}</span>
                  </p>
                </section>
              </div>
            ) : null}

            <div className="ward-plan-form-actions admin-challenges__actions">
              <p className="ward-plan-form-actions__hint" role="status" aria-live="polite">
                {saving ? (
                  `Saving ${status.label}…`
                ) : status.kind === 'error' ? (
                  <span className="ward-plan-status ward-plan-status--error">Not saved: {status.message}</span>
                ) : dirty ? (
                  <span className="ward-plan-status ward-plan-status--dirty">
                    Unsaved changes in {dirtyMonths.map(monthName).join(', ')}.
                  </span>
                ) : status.kind === 'saved' ? (
                  <span className="ward-plan-status ward-plan-status--saved">✓ All changes saved at {formatTime(status.at)}.</span>
                ) : (
                  'No unsaved changes.'
                )}
              </p>
              <div className="ward-plan-form-actions__buttons">
                {isDirty(selected) ? (
                  <button
                    type="button"
                    className="btn-secondary"
                    disabled={saving}
                    onClick={() => {
                      if (window.confirm(`Discard your unsaved changes to ${monthName(selected)}?`)) discardMonth(selected)
                    }}
                  >
                    Discard {MONTHS_SHORT[selected - 1]} Changes
                  </button>
                ) : null}
                <button type="submit" className="btn-primary ward-plan-form-actions__submit" disabled={saving || !dirty}>
                  {saving ? 'Saving…' : 'Save and Publish'}
                </button>
              </div>
            </div>
          </form>
        ) : null}
      </div>
    </main>
  )
}
