import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { apiJson } from '../api'
import { shuffled, type GalleryItem } from './gallery'

/**
 * Shared, once-per-load fetch of the public gallery list. The Ward Gallery page and every
 * rotating photo slot use the same request. Leaders' changes call invalidateGalleryItems().
 */
let cached: Promise<GalleryItem[]> | null = null

export function loadGalleryItems(): Promise<GalleryItem[]> {
  if (!cached) {
    cached = apiJson<{ items: GalleryItem[] | null }>('/api/gallery')
      .then((d) => (Array.isArray(d.items) ? d.items : []))
      .catch((e) => {
        cached = null // let a later page try again
        throw e
      })
  }
  return cached
}

export function invalidateGalleryItems() {
  cached = null
}

type Claimer = {
  ready: boolean
  /** Returns this slot's drawing for the current page, or null to keep the static photo. */
  claim: (slot: string) => GalleryItem | null
}

const Ctx = createContext<Claimer | null>(null)

/**
 * Hands out gallery drawings to the in-page photo slots (never heroes). On every page view the
 * list is reshuffled and each slot gets a different drawing; once the drawings run out the
 * remaining slots keep their original photo.
 */
export function GalleryDrawingsProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<GalleryItem[] | null>(null)
  const loc = useLocation()
  const page = useRef<{ key: string; order: GalleryItem[]; claims: Map<string, GalleryItem | null> }>({
    key: '',
    order: [],
    claims: new Map(),
  })

  useEffect(() => {
    let alive = true
    loadGalleryItems()
      .then((list) => alive && setItems(list))
      .catch(() => alive && setItems([]))
    return () => {
      alive = false
    }
  }, [])

  const claimer: Claimer = {
    ready: items !== null,
    claim(slot) {
      if (items === null) return null
      const p = page.current
      // New page view (location.key changes on every navigation): fresh shuffle, no claims.
      if (p.key !== loc.key) {
        p.key = loc.key
        p.order = shuffled(items)
        p.claims = new Map()
      }
      if (p.claims.has(slot)) return p.claims.get(slot) ?? null
      const used = new Set([...p.claims.values()].filter(Boolean).map((g) => (g as GalleryItem).id))
      const next = p.order.find((g) => !used.has(g.id)) ?? null
      p.claims.set(slot, next)
      return next
    },
  }

  return <Ctx.Provider value={claimer}>{children}</Ctx.Provider>
}

/**
 * The drawing for one photo slot. `undefined` while the list is loading, `null` when the slot
 * should show its original photo (no drawings left, gallery empty, or outside the provider).
 */
export function useSlotDrawing(slot: string | undefined): GalleryItem | null | undefined {
  const ctx = useContext(Ctx)
  if (!slot || !ctx) return null
  if (!ctx.ready) return undefined
  return ctx.claim(slot)
}
