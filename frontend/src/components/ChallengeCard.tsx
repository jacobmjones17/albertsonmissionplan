import { renderChallengeBullet } from '../lib/challengeBullets'
import type { MonthlyChallenge } from '../lib/monthlyChallenges'
import {
  MONTHLY_CHALLENGE_IMAGE_HEIGHT,
  MONTHLY_CHALLENGE_IMAGE_SIZES,
  MONTHLY_CHALLENGE_IMAGE_WIDTH,
} from '../monthlyChallengeImages'
import { MONTHS_LONG } from '../monthlyData'

/**
 * The featured challenge card on the Monthly Challenges page. Also used as the live preview in
 * the leader editor, so what leaders see is exactly what members see.
 */
export function ChallengeCard({
  challenge,
  isCurrent,
  titleId = 'challenge-card-featured-title',
  labelledBy,
  imageSrcOverride,
}: {
  challenge: MonthlyChallenge
  isCurrent: boolean
  titleId?: string
  /** Extra element id for aria-labelledby (the page's month heading). */
  labelledBy?: string
  /** A local preview (object URL) of a picture that hasn't been uploaded yet. */
  imageSrcOverride?: string
}) {
  const monthName = MONTHS_LONG[challenge.month - 1] ?? ''
  return (
    <article
      className={`challenge-card challenge-card--featured${isCurrent ? ' challenge-card--current-accent' : ''}`}
      aria-labelledby={labelledBy ? `${titleId} ${labelledBy}` : titleId}
    >
      <div className="challenge-card__art">
        {imageSrcOverride ? (
          <img
            key={imageSrcOverride}
            src={imageSrcOverride}
            width={MONTHLY_CHALLENGE_IMAGE_WIDTH}
            height={MONTHLY_CHALLENGE_IMAGE_HEIGHT}
            alt=""
            className="challenge-card__icon"
          />
        ) : (
          <img
            key={challenge.imageSrc}
            src={challenge.imageSrc}
            srcSet={challenge.imageSrcSet}
            sizes={MONTHLY_CHALLENGE_IMAGE_SIZES}
            width={MONTHLY_CHALLENGE_IMAGE_WIDTH}
            height={MONTHLY_CHALLENGE_IMAGE_HEIGHT}
            alt=""
            className="challenge-card__icon"
            loading="eager"
            decoding="async"
            fetchPriority="high"
          />
        )}
        <div className="challenge-card__art-front">
          <p className="challenge-card__month">{monthName}</p>
        </div>
        <span className="challenge-card__art-corner challenge-card__art-corner--tl" aria-hidden="true" />
        <span className="challenge-card__art-corner challenge-card__art-corner--br" aria-hidden="true" />
      </div>

      <h3 id={titleId} className="sr-only">
        {challenge.heading}
      </h3>

      <div className="challenge-card__body">
        <p className="challenge-card__goal">
          <strong>Goal:</strong> {challenge.goal}
        </p>
        <ul className="challenge-card__list">
          {challenge.bullets.map((b, i) => renderChallengeBullet(b.who, b.text, i))}
        </ul>
      </div>
    </article>
  )
}
