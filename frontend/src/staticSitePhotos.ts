/**
 * Static photo paths under `frontend/public/site/`.
 *
 * - `header` — shared photo backdrop for the transparent navbar + hero on every public page.
 *   Do not change unless you also want to replace the navbar image (and the home hero, which
 *   reuses it as its backdrop).
 * - `pageHome`, `pageWardPlan`, `pageMonthlyChallenges`, `pageExperiences`, `pageLoveShareInvite`
 *   are optional per-page hero overrides; empty = use `header`. Leave empty unless you want a
 *   different hero photo for that page.
 * - `about`, `wardPlanFeature`, `experiencesFeature`, `loveShareFeature` are decorative in-page
 *   photos rendered below the hero / near the footer (`FeaturePhoto`). `about` is used only on
 *   the home page bottom. Setting a path to "" hides that page's feature photo entirely.
 * - `notFound` — painting of the Savior shown on the "page not found" view (mistyped links).
 */
export const STATIC_SITE_PHOTOS = {
  header: '/site/header.jpg',
  pageHome: '',
  pageWardPlan: '',
  pageMonthlyChallenges: '',
  pageExperiences: '',
  pageLoveShareInvite: '',
  about: '/site/missionary-badge-heart.jpg',
  wardPlanFeature: '/site/ward-plan-feature.jpg',
  experiencesFeature: '/site/jesus-among-people.jpg',
  loveShareFeature: '/site/love-share-feature.jpg',
  notFound: '/site/jesus-carries-child.jpg',
} as const
