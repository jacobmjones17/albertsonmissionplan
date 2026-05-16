import { PageHero } from '../components/PageHero'

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
        <div className="wrap">
          <div className="video-wrap video-wrap--sites">
            <iframe
              src="https://www.youtube.com/embed/F_DDHtVZcco"
              title="How can I Love, Share, and Invite as Jesus did?"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
              allowFullScreen
              loading="lazy"
            />
          </div>
          <div className="home-mission-strip home-mission-strip--tight">
            <p>
              Love, Share, and Invite are principles that help us share the gospel in ways that are
              right for each of God&apos;s children. As you think about these ideas, pray for guidance and
              opportunities to love like the Savior.
            </p>
          </div>
          <section className="page-content-section">
            <p className="page-inline-eyebrow">For youth</p>
            <h2>Living it as a young person</h2>
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
          <section className="page-content-section">
            <p className="page-inline-eyebrow">For young adults</p>
            <h2>Living it in college and early career</h2>
            <div className="panel-sites">
              <ul className="clean">
                <li>Build genuine friendships first; invitations flow from trust.</li>
                <li>Share gospel resources when they connect to what someone is already going through.</li>
                <li>
                  Invite friends to experiences (service, gatherings, scripture discussion) where they can
                  feel the Spirit.
                </li>
              </ul>
              <p className="panel-sites__note">Replace or expand these bullets with ward-specific examples.</p>
            </div>
          </section>
          <section className="page-content-section">
            <p className="page-inline-eyebrow">For families</p>
            <h2>Living it together</h2>
            <div className="panel-sites panel-sites--highlight">
              <p className="panel-sites__p">
                Consider adding a short family goal here—one simple invitation per month, one service
                project per quarter, or a recurring &ldquo;who needs us?&rdquo; conversation at dinner.
              </p>
            </div>
          </section>
          <p className="page-footnote">
            For Church resources on sharing the gospel, see{' '}
            <a href="https://www.churchofjesuschrist.org/?lang=eng" rel="noopener noreferrer">
              ChurchofJesusChrist.org
            </a>
            . Align local messages with your bishopric and missionary leaders.
          </p>
        </div>
      </main>
    </>
  )
}
