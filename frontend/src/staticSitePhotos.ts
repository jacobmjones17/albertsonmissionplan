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
 */
export const STATIC_SITE_PHOTOS = {
  header: '/site/header.jpg',
  pageHome: '',
  pageWardPlan: '',
  pageMonthlyChallenges: '',
  pageExperiences: '',
  pageLoveShareInvite: '',
  about: '/site/home-about.jpg',
  wardPlanFeature: '/site/jesus-carries-child.jpg',
  experiencesFeature: '/site/jesus-among-people.jpg',
  loveShareFeature: '/site/missionary-badge-heart.jpg',
} as const
