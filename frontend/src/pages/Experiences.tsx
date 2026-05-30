import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { apiJson } from '../api'
import { MISSIONARY_REFERRAL_URL } from '../churchLinks'
import { useBootstrap } from '../BootstrapContext'
import { FeaturePhoto } from '../components/FeaturePhoto'
import { PageHero } from '../components/PageHero'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

type Experience = { body: string; author: string; when?: string }

export function Experiences() {
  const { refresh } = useBootstrap()
  const [list, setList] = useState<Experience[]>([])
  const [flash, setFlash] = useState<string | null>(null)
  const [listErr, setListErr] = useState<string | null>(null)
  const [formErr, setFormErr] = useState<string | null>(null)
  const [author, setAuthor] = useState('')
  const [body, setBody] = useState('')
  const [hp, setHp] = useState('')

  async function load() {
    try {
      const d = await apiJson<{ experiences: Experience[] }>('/api/experiences')
      setList(d.experiences)
      setListErr(null)
    } catch {
      setListErr('Could not load experiences.')
    }
  }

  useEffect(() => {
    void load()
    void refresh()
  }, [refresh])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setFlash(null)
    setFormErr(null)
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
      setFormErr(ex instanceof Error ? ex.message : 'Could not submit.')
    }
  }

  return (
    <>
      <PageHero
        titleId="page-title"
        eyebrow="Stories of faith"
        title="Mission Experiences"
        lede="Read how the Lord has blessed ward members — then share your own story"
      />
      <div className="home-mission-strip home-mission-strip--tight">
        <p>
          Scroll down to read what others have shared. When you&apos;re ready, use the form below to send
          your experience — submissions are reviewed before they appear publicly. Please avoid full names of
          friends, investigators, or ward members—use initials or general descriptions instead.
        </p>
        <p className="home-mission-strip__aside">
          Know someone who wants to meet the missionaries?{' '}
          <a href={MISSIONARY_REFERRAL_URL} target="_blank" rel="noopener noreferrer">
            Submit a Church referral
          </a>{' '}
          (with their permission), or read more on{' '}
          <Link to="/love-share-invite">Love, Share, and Invite</Link>.
        </p>
      </div>
      <main id="main">
        <div className="wrap">
          <section className="page-content-section experiences-published-first" aria-labelledby="published-h">
            <p className="page-inline-eyebrow">From the ward</p>
            <h2 id="published-h">Published experiences</h2>
            <p className="experiences-published-first__intro panel-sites__note">
              Stories appear here after ward leaders review them.
            </p>
            {listErr ? <p className="panel-sites panel-sites--warn">{listErr}</p> : null}
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
              !listErr && (
                <p className="lead">Nothing published yet—be the first after leader review.</p>
              )
            )}
          </section>

          <section className="panel-sites experiences-share-after" aria-labelledby="share-h">
            <p className="page-inline-eyebrow">Share yours</p>
            <h2 id="share-h" className="panel-sites__title">
              Tell us your experience
            </h2>
            <p className="panel-sites__p">
              No account is needed. Leaders review each submission before it appears above. To protect
              privacy, please <strong>do not include full names</strong> of friends, investigators, or ward
              members. Use initials or roles instead.
            </p>
            {flash ? <p className="flash">{flash}</p> : null}
            {formErr ? <p className="panel-sites panel-sites--warn">{formErr}</p> : null}
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

          <FeaturePhoto
            src={STATIC_SITE_PHOTOS.experiencesFeature}
            caption="“The greatest service we can provide is to help others come unto Christ.” — Elder Dieter F. Uchtdorf"
          />
        </div>
      </main>
    </>
  )
}
