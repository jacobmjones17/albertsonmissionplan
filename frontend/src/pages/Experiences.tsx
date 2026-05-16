import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { apiJson } from '../api'
import { useBootstrap } from '../BootstrapContext'
import { PageHero } from '../components/PageHero'

type Experience = { body: string; author: string; when?: string }

export function Experiences() {
  const { refresh } = useBootstrap()
  const [list, setList] = useState<Experience[]>([])
  const [flash, setFlash] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [author, setAuthor] = useState('')
  const [body, setBody] = useState('')
  const [hp, setHp] = useState('')

  async function load() {
    try {
      const d = await apiJson<{ experiences: Experience[] }>('/api/experiences')
      setList(d.experiences)
    } catch {
      setErr('Could not load experiences.')
    }
  }

  useEffect(() => {
    void load()
    void refresh()
  }, [refresh])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setFlash(null)
    setErr(null)
    try {
      const res = await apiJson<{ ok: boolean; message?: string }>('/api/experiences', {
        method: 'POST',
        json: { body, author, website: hp },
      })
      if (res.ok && res.message) setFlash(res.message)
      setBody('')
      setAuthor('')
      await load()
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : 'Could not submit.')
    }
  }

  return (
    <>
      <PageHero
        titleId="page-title"
        eyebrow="Stories of faith"
        title="Mission Experiences"
        lede="Share how the Lord has blessed you as you live the ward mission plan"
      />
      <div className="home-mission-strip home-mission-strip--tight">
        <p>
          Submissions are reviewed before they appear publicly. Please avoid full names of friends,
          investigators, or ward members—use initials or general descriptions instead.
        </p>
      </div>
      <main id="main">
        <div className="wrap">
          {flash ? <p className="flash">{flash}</p> : null}
          {err ? <p className="panel-sites panel-sites--warn">{err}</p> : null}
          <section className="panel-sites" aria-labelledby="share-h">
            <p className="page-inline-eyebrow">Submit</p>
            <h2 id="share-h" className="panel-sites__title">
              Share an experience
            </h2>
            <p className="panel-sites__p">
              No account is needed. Leaders review each submission before it appears. To protect
              privacy, please <strong>do not include full names</strong> of friends, investigators, or ward
              members. Use initials or roles instead.
            </p>
            <form className="form-stack" onSubmit={(e) => void onSubmit(e)}>
              <label className="hp-field" aria-hidden="true">
                Leave blank
                <input type="text" name="website" tabIndex={-1} autoComplete="off" value={hp} onChange={(e) => setHp(e.target.value)} />
              </label>
              <label>
                How you&apos;d like to be listed (optional)
                <input
                  type="text"
                  name="author"
                  maxLength={120}
                  placeholder='First name or "Anonymous"'
                  value={author}
                  onChange={(e) => setAuthor(e.target.value)}
                />
              </label>
              <label>
                Your experience
                <textarea
                  name="body"
                  rows={8}
                  required
                  maxLength={8000}
                  placeholder="What happened? How were you blessed as you ministered, invited, or lived the mission focus?"
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                />
              </label>
              <button type="submit" className="btn btn-primary">
                Submit for review
              </button>
            </form>
          </section>
          <section className="page-content-section">
            <p className="page-inline-eyebrow">From the ward</p>
            <h2>Published experiences</h2>
            {list.length > 0 ? (
              <div className="testimonial-stack">
                {list.map((ex, i) => (
                  <article className="testimonial-card testimonial-card--sites" key={i}>
                    <p className="testimonial-body">{ex.body}</p>
                    <p className="testimonial-meta">
                      {ex.author}
                      {ex.when ? ` · ${ex.when}` : ''}
                    </p>
                  </article>
                ))}
              </div>
            ) : (
              <p className="lead">Nothing published yet—be the first after leader review.</p>
            )}
          </section>
        </div>
      </main>
    </>
  )
}
