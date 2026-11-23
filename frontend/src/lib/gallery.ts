/**
 * One drawing as returned by GET /api/gallery and /api/admin/gallery. By design there are no
 * names, titles, or captions: the gallery shows only the pictures, to protect the children.
 */
export type GalleryItem = {
  id: number
  width: number
  height: number
  thumbWidth: number
  thumbHeight: number
  imageUrl: string
  thumbUrl: string
  alt: string
  hidden?: boolean
  when: string
}

/**
 * Fisher-Yates shuffle (returns a new array). The public gallery shows every visible drawing,
 * in a fresh random order on each page load, so no one's picture is always at the bottom.
 */
export function shuffled<T>(list: readonly T[], random: () => number = Math.random): T[] {
  const out = list.slice()
  for (let i = out.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1))
    ;[out[i], out[j]] = [out[j], out[i]]
  }
  return out
}
