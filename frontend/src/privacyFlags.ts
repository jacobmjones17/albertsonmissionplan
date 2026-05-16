/**
 * Mirrors backend/internal/privacyhint/hint.go heuristics.
 * Keep UTF-16 string indices in sync with the DOM textarea slice API.
 */

export type PrivacyFlagKind = 'brother_sister' | 'two_cap' | 'lone_conj'

export type PrivacyFlag = {
  kind: PrivacyFlagKind
  /** start index inclusive (textarea / JavaScript slice) */
  start: number
  /** end index exclusive */
  end: number
  text: string
  message: string
}

/** Enough room for several two-word names plus a lone capitalized name. */
const MAX_FLAGS = 6

// Must stay aligned with backend allowedPairs.
const ALLOWED_TWO_WORD = new Set<string>([
  'book mormon',
  'jesus christ',
  'heavenly father',
  'the lord',
  'holy ghost',
  'latter day',
  'new testament',
  'old testament',
  'our young',
  'young women',
  'young men',
  'elders quorum',
  'relief society',
  'sunday school',
  'ward council',
  'stake conference',
  'general conference',
  'sacrament meeting',
  'family history',
  'temple recommend',
  'temple trip',
  'family home',
  'home evening',
  'full time',
  'part time',
  'ministering brothers',
  'ministering sisters',
  'mission leaders',
  'ward mission',
  'primary program',
  'primary children',
  'aaronic priesthood',
  'melchizedek priesthood',
  'first presidency',
  'quorum twelve',
  'salt lake',
  'north america',
  'south america',
  'united states',
  'new year',
  'new members',
  'new convert',
  'great blessing',
  'great experience',
  'other ward',
  'other stake',
  'my family',
  'our family',
  'our ward',
  'our stake',
  'our missionaries',
  'our neighbors',
  'our friends',
  'bishopric youth',
  'stake president',
  'mission president',
  'area authority',
  'relief presidency',
  'elders president',
  'elder quorum',
  'high council',
  'patriarch blessing',
  'eternal families',
  'eternal family',
  'president nelson',
  'president eyring',
  'president cook',
  'temple square',
])

/** Words that look like grammar after "and"/"or", not people (lowercase match). */
const SKIP_AFTER_CONJ = new Set([
  'then',
  'the',
  'also',
  'finally',
  'now',
  'again',
  'yet',
  'so',
  'when',
  'why',
  'how',
  'every',
  'each',
  'all',
  'some',
  'most',
  'such',
  'own',
  'same',
  'other',
  'more',
  'less',
])

/** Short hint for the compact UI (full logic still drives highlights). */
export function privacyFlagHint(kind: PrivacyFlagKind): string {
  switch (kind) {
    case 'brother_sister':
      return 'Title + name'
    case 'two_cap':
      return 'Two-word name'
    case 'lone_conj':
      return 'After and/or'
    default:
      return ''
  }
}

/** Stable key while the same flagged substring stays in the draft. */
export function flagDismissKey(flag: PrivacyFlag): string {
  return `${flag.kind}\u241f${flag.text}`
}

/** Name-like snippets for leaders editing before publishing. */
export function computePrivacyFlags(raw: string): PrivacyFlag[] {
  if (raw.trim().length === 0) {
    return []
  }

  const out: PrivacyFlag[] = []

  const reBrother = new RegExp('\\b(?:brother|sister)\\s+[A-Z][a-z]{2,}\\b', 'i')
  const bro = reBrother.exec(raw)
  if (bro) {
    const full = bro[0]
    out.push({
      kind: 'brother_sister',
      start: bro.index,
      end: bro.index + full.length,
      text: full,
      message: 'Brother/Sister + name',
    })
  }

  if (out.length >= MAX_FLAGS) {
    return out.slice(0, MAX_FLAGS)
  }

  // First word ≥3 letters (e.g. "Bre"), second word ≥3 (e.g. "Jones") — aligns with Go reTwoCap.
  const reTwo = /\b([A-Z][a-z]{2,})\s+([A-Z][a-z]{2,})\b/g
  const seenPair = new Set<string>()

  for (const m of raw.matchAll(reTwo)) {
    const full = m[0]
    const i = m.index ?? 0
    const word1 = m[1]
    const word2 = m[2]
    if (word1 == null || word2 == null) {
      continue
    }
    const pair = `${word1} ${word2}`.toLowerCase()
    if (ALLOWED_TWO_WORD.has(pair)) {
      continue
    }
    if (seenPair.has(full)) {
      continue
    }
    seenPair.add(full)
    out.push({
      kind: 'two_cap',
      start: i,
      end: i + full.length,
      text: full,
      message: 'Two-word name',
    })

    if (out.length >= MAX_FLAGS) {
      return out.slice(0, MAX_FLAGS)
    }
  }

  const reLone = /\b(?:and|or)\s+([A-Z][a-z]{2,})\b/g
  const seenLone = new Set<string>()

  for (const m of raw.matchAll(reLone)) {
    const name = m[1]
    if (name == null) {
      continue
    }
    if (SKIP_AFTER_CONJ.has(name.toLowerCase())) {
      continue
    }
    const full = m[0]
    const idx = m.index ?? 0
    const start = idx + full.length - name.length
    const end = start + name.length
    const key = `${start}\u241f${name}`
    if (seenLone.has(key)) {
      continue
    }
    seenLone.add(key)
    out.push({
      kind: 'lone_conj',
      start,
      end,
      text: name,
      message: 'After and/or',
    })

    if (out.length >= MAX_FLAGS) {
      break
    }
  }

  return out
}
