import { useCallback, useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'
import { Toast, type ToastPayload } from '../components/Toast'
import { RichGoalsEditor } from '../components/richtext/RichGoalsEditor'

type Org = { slug: string; title: string; bullets: string[] }

/** `ward-wide` or an organization slug from the API. */
type WardPlanTab = 'ward-wide' | string

export function AdminWardPlan() {
  const { refresh } = useBootstrap()
  const [wardGoals, setWardGoals] = useState('')
  const [orgs, setOrgs] = useState<Org[]>([])
  const [orgText, setOrgText] = useState<Record<string, string>>({})
  const [toast, setToast] = useState<ToastPayload | null>(null)
  const [err, setErr] = useState<string | null>(null)
  /** Increment after loading from API so TipTap remounts with server content. */
  const [editorResetKey, setEditorResetKey] = useState(0)
  const [activeTab, setActiveTab] = useState<WardPlanTab>('ward-wide')

  const load = useCallback(async () => {
    setErr(null)
    try {
      const d = await apiJson<{ wardGoals: string[]; orgs: Org[] }>('/api/admin/ward-plan')
      setWardGoals(d.wardGoals.join('\n'))
      setOrgs(d.orgs)
      const m: Record<string, string> = {}
      for (const o of d.orgs) {
        m[o.slug] = o.bullets.join('\n')
      }
      setOrgText(m)
      setEditorResetKey((k) => k + 1)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Could not load.')
    }
  }, [])

  useEffect(() => {
    void load()
    void refresh()
  }, [load, refresh])

  useEffect(() => {
    if (activeTab !== 'ward-wide' && !orgs.some((o) => o.slug === activeTab)) {
      setActiveTab('ward-wide')
    }
  }, [orgs, activeTab])

  async function onSave(e: FormEvent) {
    e.preventDefault()
    setToast(null)
    try {
      await apiJson<{ ok: boolean; message?: string }>('/api/admin/ward-plan', {
        method: 'POST',
        json: { wardGoals, orgs: orgText },
      })
      setToast({ message: 'Changes saved.', variant: 'success' })
    } catch (ex) {
      setToast({
        message: ex instanceof Error ? ex.message : 'Could not save.',
        variant: 'error',
      })
    }
  }

  const activeOrg = orgs.find((o) => o.slug === activeTab)
  const editorLabel = activeTab === 'ward-wide' ? 'Ward-wide goals' : (activeOrg?.title ?? activeTab)
  const editorValue = activeTab === 'ward-wide' ? wardGoals : (orgText[activeTab] ?? '')
  const onEditorChange =
    activeTab === 'ward-wide'
      ? setWardGoals
      : (v: string) => setOrgText((prev) => ({ ...prev, [activeTab]: v }))

  const tabsId = 'ward-plan-tabs'
  const panelId = 'ward-plan-editor-panel'

  return (
    <>
      <PageHero
        titleId="ward-edit-title"
        eyebrow="Editing"
        title="Edit ward plan"
        lede="Format goals like a document: bold, lists, colors, links. Each top-level bullet is one line on the public page. Starter text is only a suggestion—rewrite any line or the whole list to match your ward. Choose a tab below—save publishes every organization at once."
        compact
      />
      <main id="main" className="wrap-wide ward-plan-page">
        <p className="admin-crumb">
          <Link to="/admin">&larr; Back to leader tools</Link>
        </p>
        <Toast toast={toast} onDismiss={() => setToast(null)} />
        {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}

        <form className="form-stack ward-plan-editor" onSubmit={(e) => void onSave(e)}>
          <div className="ward-plan-tabs-wrap">
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
                <button
                  type="button"
                  role="tab"
                  id={`${tabsId}-tab-ward`}
                  aria-selected={activeTab === 'ward-wide'}
                  aria-controls={panelId}
                  className={`ward-plan-tabs__tab${activeTab === 'ward-wide' ? ' is-active' : ''}`}
                  onClick={() => setActiveTab('ward-wide')}
                >
                  Ward-wide
                </button>
                {orgs.map((o) => (
                  <button
                    key={o.slug}
                    type="button"
                    role="tab"
                    id={`${tabsId}-tab-${o.slug}`}
                    aria-selected={activeTab === o.slug}
                    aria-controls={panelId}
                    className={`ward-plan-tabs__tab${activeTab === o.slug ? ' is-active' : ''}`}
                    onClick={() => setActiveTab(o.slug)}
                  >
                    {o.title}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div
            id={panelId}
            role="tabpanel"
            aria-labelledby={
              activeTab === 'ward-wide' ? `${tabsId}-tab-ward` : `${tabsId}-tab-${activeTab}`
            }
          >
            <RichGoalsEditor
              key={`${activeTab}-${editorResetKey}`}
              label={editorLabel}
              value={editorValue}
              onChange={onEditorChange}
              resetKey={editorResetKey}
            />
          </div>

          <div className="ward-plan-form-actions">
            <p className="ward-plan-form-actions__hint">One save updates every tab.</p>
            <button type="submit" className="btn-primary ward-plan-form-actions__submit">
              Save all changes
            </button>
          </div>
        </form>
      </main>
    </>
  )
}
