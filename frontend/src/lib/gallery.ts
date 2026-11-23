/** One drawing as returned by GET /api/gallery and /api/admin/gallery. */
export type GalleryItem = {
  id: number
  title: string
  caption: string
  artist: string
  ageGroup: string
  width: number
  height: number
  thumbWidth: number
  thumbHeight: number
  imageUrl: string
  thumbUrl: string
  alt: string
  featured: boolean
  hidden?: boolean
  when: string
}

/** "by Mia · Sunbeams" style credit line, or '' when nothing was entered. */
export function galleryCredit(item: Pick<GalleryItem, 'artist' | 'ageGroup'>): string {
  const parts: string[] = []
  if (item.artist) parts.push(`by ${item.artist}`)
  if (item.ageGroup) parts.push(item.ageGroup)
  return parts.join(' · ')
}

/** Suggestions for the age / class field (leaders can type anything). */
export const AGE_GROUP_SUGGESTIONS = [
  'Nursery',
  'Sunbeams',
  'CTR 4',
  'CTR 5',
  'CTR 6',
  'CTR 7',
  'Valiant 8',
  'Valiant 9',
  'Valiant 10',
  'Valiant 11',
  'Primary',
  'Youth',
  'Young Women',
  'Aaronic Priesthood',
  'Relief Society',
  'Elders Quorum',
  'Ward member',
]
