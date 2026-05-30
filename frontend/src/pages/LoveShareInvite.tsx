import { MISSIONARY_REFERRAL_INFO_URL, MISSIONARY_REFERRAL_URL } from '../churchLinks'
import { FeaturePhoto } from '../components/FeaturePhoto'
import { PageHero } from '../components/PageHero'
import { STATIC_SITE_PHOTOS } from '../staticSitePhotos'

export function LoveShareInvite() {
  return (
    <>
      <PageHero
        titleId="love-title"
        eyebrow="A Christlike pattern"
        title="Love, Share, and Invite"
        lede="Simple ways to love like the Savior and invite others to come unto Him"
        banner="Simple ways to Love, Share, and Invite"
      />
      <main id="main">
        <div className="wrap-wide">
          <div className="love-share-video-stack">
            <div className="love-share-lead">
              <section className="referral-callout" aria-labelledby="referral-callout-title">
                <h2 id="referral-callout-title" className="referral-callout__title">
                  Refer a friend to the missionaries
                </h2>
                <p className="referral-callout__lede">
                  When someone you know is open to learning more, use the Church&apos;s official referral
                  system. Ask for their permission first, then submit their information so missionaries can
                  coordinate with you on next steps.
                </p>
                <p className="referral-callout__actions">
                  <a
                    className="btn btn-primary"
                    href={MISSIONARY_REFERRAL_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Submit a referral
                  </a>
                </p>
                <p className="referral-callout__note">
                  You can also refer through the Member Tools app.{' '}
                  <a href={MISSIONARY_REFERRAL_INFO_URL} target="_blank" rel="noopener noreferrer">
                    Learn how referrals work
                  </a>
                  .
                </p>
              </section>

              <div className="home-mission-strip home-mission-strip--tight love-share-intro-strip">
                <p>
                  Love, Share, and Invite are principles that help us share the gospel in ways that are
                  right for each of God&apos;s children. The videos below can help you think about how to
                  love like the Savior and invite others to come unto Him.
                </p>
              </div>
            </div>

            <div className="video-wrap video-wrap--sites video-wrap--love-share-hero">
              <iframe
                src="https://www.youtube.com/embed/F_DDHtVZcco"
                title="How can I Love, Share, and Invite as Jesus did?"
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                allowFullScreen
                loading="lazy"
              />
            </div>

            {/* Two columns on desktop; stacked on mobile so each video keeps its bullets underneath. */}
            <div className="love-share-columns">
              <div className="love-share-column">
                <p className="love-share-audience-label">For youth</p>
                <div className="video-wrap video-wrap--sites video-wrap--love-share-split">
                  <iframe
                    src="https://www.youtube.com/embed/FB1tddEMezA?start=6"
                    title="Love, Share, and Invite — for youth"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                    allowFullScreen
                    loading="lazy"
                  />
                </div>
                <section className="love-share-panel">
                  <h2 className="love-share-panel__title">Living it as a young person</h2>
                  <div className="panel-sites">
                    <ul className="clean">
                      <li>Look for someone who needs a friend at school or in your quorum or class.</li>
                      <li>Share what you believe naturally—in conversation, not performance.</li>
                      <li>Invite others to uplifting activities, not only Sunday meetings.</li>
                    </ul>
                    <p className="panel-sites__note">
                      Edit this list with ideas from your ward youth council or full-time missionaries.
                    </p>
                  </div>
                </section>
              </div>

              <div className="love-share-column">
                <p className="love-share-audience-label">For young adults</p>
                <div className="video-wrap video-wrap--sites video-wrap--love-share-split">
                  <iframe
                    src="https://www.youtube.com/embed/jPbuguEBPC4"
                    title="Love, Share, and Invite — for young adults"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                    allowFullScreen
                    loading="lazy"
                  />
                </div>
                <section className="love-share-panel">
                  <h2 className="love-share-panel__title">Living it in college and early career</h2>
                  <div className="panel-sites">
                    <ul className="clean">
                      <li>Build genuine friendships first; invitations flow from trust.</li>
                      <li>
                        Share gospel resources when they connect to what someone is already going through.
                      </li>
                      <li>
                        Invite friends to experiences (service, gatherings, scripture discussion) where they
                        can feel the Spirit.
                      </li>
                    </ul>
                    <p className="panel-sites__note">
                      Replace or expand these bullets with ward-specific examples.
                    </p>
                  </div>
                </section>
              </div>
            </div>
          </div>

          <div className="love-share-outro">
            <FeaturePhoto src={STATIC_SITE_PHOTOS.loveShareFeature} size="narrow" variant="flat" />
            <p className="page-footnote love-share-outro__footnote">
              For Church resources on sharing the gospel, see{' '}
              <a href="https://www.churchofjesuschrist.org/?lang=eng" rel="noopener noreferrer">
                ChurchofJesusChrist.org
              </a>
              . Align local messages with your bishopric and missionary leaders.
            </p>
          </div>
        </div>
      </main>
    </>
  )
}
