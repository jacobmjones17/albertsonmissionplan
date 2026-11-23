import { useCallback, useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import { RichGoalsEditor } from '../components/richtext/RichGoalsEditor'
import { WardGoalLine } from '../components/richtext/WardGoalLine'
import { useUnsavedChangesWarning } from '../lib/useUnsavedChangesWarning'

type Org = { slug: string; title: string; bullets: string[] }

/** `ward-wide` or an organization slug from the API. */
type WardPlanTab = 'ward-wide' | string

type SaveStatus =
  | { kind: 'idle' }
  | { kind: 'saving' }
  | { kind: 'saved'; at: Date }
  | { kind: 'error'; message: string }

function storedLines(text: string): string[] {
  return text
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)
}

/** Compare ignoring blank lines / trailing spaces so a no-op edit doesn't count as a change. */
function sameContent(a: string, b: string): boolean {
  return storedLines(a).join('\n') === storedLines(b).join('\n')
}

function formatTime(d: Date): string {
  return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
}

export function AdminWardPlan() {
  const { refresh } = useBootstrap()

  const [wardGoals, setWardGoals] = useState('')
  const [orgs, setOrgs] = useState<Org[]>([])
  const [orgText, setOrgText] = useState<Record<string, string>>({})
  /** Last content loaded from / saved to the server, per tab — drives the unsaved dots. */
  const [saved, setSaved] = useState<Record<WardPlanTab, string>>({})
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState<SaveStatus>({ kind: 'idle' })
  /** Increment after loading from API so TipTap remounts with server content. */
  const [editorResetKey, setEditorResetKey] = useState(0)
  const [activeTab, setActiveTab] = useState<WardPlanTab>('ward-wide')

  const load = useCallback(async () => {
    setErr(null)
    setLoading(true)
    try {
      const d = await apiJson<{ wardGoals: string[]; orgs: Org[] }>('/api/admin/ward-plan')
      const ward = d.wardGoals.join('\n')
      setWardGoals(ward)
      setOrgs(d.orgs)
      const m: Record<string, string> = {}
      for (const o of d.orgs) {
        m[o.slug] = o.bullets.join('\n')
      }
      setOrgText(m)
      setSaved({ 'ward-wide': ward, ...m })
      setEditorResetKey((k) => k + 1)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load the ward plan.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
    void refresh()
  }, [load, refresh])

  useEffect(() => {
    if (!orgs.some((o) => o.slug === activeTab) && activeTab !== 'ward-wide') {
      setActiveTab('ward-wide')
    }
  }, [orgs, activeTab])

  const textFor = useCallback(
    (tab: WardPlanTab) => (tab === 'ward-wide' ? wardGoals : (orgText[tab] ?? '')),
    [wardGoals, orgText],
  )

  const tabs = useMemo(
    () => [{ slug: 'ward-wide', title: 'Ward-wide' }, ...orgs.map((o) => ({ slug: o.slug, title: o.title }))],
    [orgs],
  )
  const dirtyTabs = useMemo(
    () => tabs.filter((t) => saved[t.slug] !== undefined && !sameContent(textFor(t.slug), saved[t.slug])),
    [tabs, saved, textFor],
  )
  const dirty = dirtyTabs.length > 0
  const dirtySet = useMemo(() => new Set(dirtyTabs.map((t) => t.slug)), [dirtyTabs])

  useUnsavedChangesWarning(dirty)

  async function onSave(e: FormEvent) {
    e.preventDefault()
    setToast(null)
    setStatus({ kind: 'saving' })
    try {
      await apiJson<{ ok: boolean; message?: string }>('/api/admin/ward-plan', {
        method: 'POST',
        json: { wardGoals, orgs: orgText },
      })
      setSaved({ 'ward-wide': wardGoals, ...orgText })
      const at = new Date()
      setStatus({ kind: 'saved', at })
      setToast({ message: 'Saved. The public Ward Plan page now shows these goals.', variant: 'success' })
    } catch (ex) {
      const message = ex instanceof Error ? ex.message : 'Could not save.'
      setStatus({ kind: 'error', message })
      setToast({ message: `Not saved: ${message}`, variant: 'error' })
    }
  }

  function discardChanges() {
    if (!window.confirm('Discard all unsaved changes on every tab?')) return
    setWardGoals(saved['ward-wide'] ?? '')
    const m: Record<string, string> = {}
    for (const o of orgs) m[o.slug] = saved[o.slug] ?? ''
    setOrgText(m)
    setEditorResetKey((k) => k + 1)
    setStatus({ kind: 'idle' })
  }

  const activeTitle = tabs.find((t) => t.slug === activeTab)?.title ?? activeTab
  const editorValue = textFor(activeTab)
  const onEditorChange =
    activeTab === 'ward-wide'
      ? setWardGoals
      : (v: string) => setOrgText((prev) => ({ ...prev, [activeTab]: v }))
  const previewLines = storedLines(editorValue)

  const tabsId = 'ward-plan-tabs'
  const panelId = 'ward-plan-editor-panel'

  return (
    <main id="main">
      <PageHero
        titleId="ward-edit-title"
        eyebrow="Leader Tools"
        title="Edit Ward Plan"
        lede="Pick a group, edit its goals, and check the preview. Save publishes every group at once."
        compact
      />
      <div className="wrap-wide ward-plan-page">
        <Toast toast={toast} onDismiss={() => setToast(null)} durationMs={6000} />
        {err ? (
          <p className="panel-sites panel-sites--warn" role="alert">
            {err}
          </p>
        ) : null}
        {loading ? (
          <p className="loading-note" role="status">
            <span className="loading-note__spinner" aria-hidden="true" />
            Loading the ward plan…
          </p>
        ) : null}

        <form className="form-stack ward-plan-editor" onSubmit={(e) => void onSave(e)} hidden={loading}>
          <div className="ward-plan-tabs-wrap">
            <label className="ward-plan-tabs-select">
              <span className="ward-plan-tabs-label">Goals for</span>
              <select value={activeTab} onChange={(e) => setActiveTab(e.target.value)}>
                {tabs.map((t) => (
                  <option key={t.slug} value={t.slug}>
                    {t.title}
                    {dirtySet.has(t.slug) ? ' • unsaved' : ''}
                  </option>
                ))}
              </select>
            </label>
            <div className="ward-plan-tabs-desktop">
              <div className="ward-plan-tabs-label" id={`${tabsId}-label`}>
                Goals for
              </div>
              <div className="ward-plan-tabs-shell">
                <div
                  id={tabsId}
                  className="ward-plan-tabs"
                  role="tablist"
                  aria-labelledby={`${tabsId}-label`}
                >
                  {tabs.map((t) => {
                    const isActive = activeTab === t.slug
                    const isDirty = dirtySet.has(t.slug)
                    return (
                      <button
                        key={t.slug}
                        type="button"
                        role="tab"
                        id={`${tabsId}-tab-${t.slug}`}
                        aria-selected={isActive}
                        aria-controls={panelId}
                        className={`ward-plan-tabs__tab${isActive ? ' is-active' : ''}`}
                        onClick={() => setActiveTab(t.slug)}
                      >
                        {t.title}
                        {isDirty ? (
                          <>
                            <span className="ward-plan-tabs__dot" aria-hidden="true" />
                            <span className="sr-only"> (unsaved changes)</span>
                          </>
                        ) : null}
                      </button>
                    )
                  })}
                </div>
              </div>
            </div>
          </div>

          <div
            id={panelId}
            role="tabpanel"
            aria-labelledby={`${tabsId}-tab-${activeTab}`}
            className="ward-plan-editor-panel"
          >
            <RichGoalsEditor
              key={`${activeTab}-${editorResetKey}`}
              label={`${activeTitle} goals`}
              value={editorValue}
              onChange={onEditorChange}
              resetKey={editorResetKey}
            />

            <section className="ward-plan-preview" aria-labelledby="ward-plan-preview-title">
              <h3 id="ward-plan-preview-title" className="ward-plan-preview__title">
                Preview <span>how {activeTitle} looks on the public Ward Plan page</span>
              </h3>
              {previewLines.length === 0 ? (
                <p className="ward-plan-preview__empty">No goals yet for this group.</p>
              ) : activeTab === 'ward-wide' ? (
                <ul className="ward-goals-spotlight__list">
                  {previewLines.map((g, i) => (
                    <li key={i}>
                      <WardGoalLine content={g} />
                    </li>
                  ))}
                </ul>
              ) : (
                <div className="org-body">
                  <ul className="clean">
                    {previewLines.map((b, i) => (
                      <li key={i}>
                        <WardGoalLine content={b} />
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </section>
          </div>

          <div className="ward-plan-form-actions">
            <p className="ward-plan-form-actions__hint" role="status" aria-live="polite">
              {status.kind === 'saving' ? (
                'Saving…'
              ) : status.kind === 'error' ? (
                <span className="ward-plan-status ward-plan-status--error">
                  Not saved: {status.message}
                </span>
              ) : dirty ? (
                <span className="ward-plan-status ward-plan-status--dirty">
                  Unsaved changes in {dirtyTabs.map((t) => t.title).join(', ')}.
                </span>
              ) : status.kind === 'saved' ? (
                <span className="ward-plan-status ward-plan-status--saved">
                  ✓ All changes saved at {formatTime(status.at)}.
                </span>
              ) : (
                'No unsaved changes. Saving publishes every group at once.'
              )}
            </p>
            <div className="ward-plan-form-actions__buttons">
              {dirty ? (
                <button type="button" className="btn-secondary" onClick={discardChanges}>
                  Discard Changes
                </button>
              ) : null}
              <button
                type="submit"
                className="btn-primary ward-plan-form-actions__submit"
                disabled={status.kind === 'saving' || !dirty}
              >
                {status.kind === 'saving' ? 'Saving…' : 'Save and Publish'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </main>
  )
}
