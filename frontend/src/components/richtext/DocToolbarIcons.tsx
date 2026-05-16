/** Material-style 24×24 icons (familiar Google Docs–like toolbar glyphs). */

import type { ReactNode } from 'react'

const vb = { viewBox: '0 0 24 24' as const, width: 20 as const, height: 20 as const }

function Ic(props: { children: ReactNode }) {
  const { children } = props
  return (
    <svg {...vb} className="rich-doc-icon" aria-hidden>
      {children}
    </svg>
  )
}

export function IconUndo() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M12.5 8c-2.65 0-5.05.99-6.9 2.6L2 7v9h9l-3.62-3.62c1.39-1.16 3.16-1.88 5.12-1.88 3.54 0 6.55 2.31 7.6 5.5l2.37-.78C21.08 11.03 17.15 8 12.5 8z"
      />
    </Ic>
  )
}

export function IconRedo() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M18.4 10.6C16.55 8.99 14.15 8 11.5 8c-4.65 0-8.58 3.03-9.96 7.22L3.9 16c1.05-2.19 4.05-4.5 7.6-4.5 1.95 0 3.73.72 5.12 1.88L13 16h9V7l-3.6 3.6z"
      />
    </Ic>
  )
}

export function IconBold() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M15.6 10.79c.97-.67 1.65-1.77 1.65-2.79 0-2.26-1.75-4-4-4H7v14h7.04c2.09 0 3.71-1.7 3.71-3.79 0-1.52-.86-2.82-2.19-3.42zM10 6.5h3c.83 0 1.5.67 1.5 1.5s-.67 1.5-1.5 1.5h-3v-3zm3.5 9H10v-3h3.5c.83 0 1.5.67 1.5 1.5s-.67 1.5-1.5 1.5z"
      />
    </Ic>
  )
}

export function IconItalic() {
  return (
    <Ic>
      <path fill="currentColor" d="M10 4v3h2.21l-3.42 8H6v3h8v-3h-2.21l3.42-8H18V4h-8z" />
    </Ic>
  )
}

export function IconUnderline() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M12 17c3.31 0 6-2.69 6-6V3h-2.5v8c0 1.93-1.57 3.5-3.5 3.5S8.5 12.93 8.5 11V3H6v8c0 3.31 2.69 6 6 6zm6.39 3H5.61l-.99 2h15.78l-1-2z"
      />
    </Ic>
  )
}

export function IconStrikethrough() {
  return (
    <Ic>
      <path
        fill="none"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M16 4H9a3 3 0 0 0-2.83 4M14 12a3 3 0 0 1 0 4H6"
      />
      <path fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" d="M4 12h16" />
    </Ic>
  )
}

export function IconSubscript() {
  return (
    <svg viewBox="0 0 24 24" width={20} height={20} className="rich-doc-icon" aria-hidden>
      <text x="3" y="16" fontSize="13" fill="currentColor" fontFamily="Georgia, serif" fontWeight="600">
        x
      </text>
      <text x="13" y="18" fontSize="9" fill="currentColor" fontFamily="Georgia, serif">
        2
      </text>
    </svg>
  )
}

export function IconSuperscript() {
  return (
    <svg viewBox="0 0 24 24" width={20} height={20} className="rich-doc-icon" aria-hidden>
      <text x="3" y="17" fontSize="13" fill="currentColor" fontFamily="Georgia, serif" fontWeight="600">
        x
      </text>
      <text x="13" y="10" fontSize="9" fill="currentColor" fontFamily="Georgia, serif">
        2
      </text>
    </svg>
  )
}

export function IconBulletList() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M4 10.5c-.83 0-1.5.67-1.5 1.5s.67 1.5 1.5 1.5 1.5-.67 1.5-1.5-.67-1.5-1.5-1.5zm0-6c-.83 0-1.5.67-1.5 1.5S3.17 7.5 4 7.5 5.5 6.83 5.5 6 4.83 4.5 4 4.5zm0 12c-.82 0-1.5.68-1.5 1.5s.68 1.5 1.5 1.5 1.5-.67 1.5-1.5-.67-1.5-1.5-1.5zM7 19h14v-2H7v2zm0-6h14v-2H7v2zm0-8v2h14V5H7z"
      />
    </Ic>
  )
}

export function IconNumberedList() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M2 17h2v.5H3v1h1v-.5h1v1H2v-2zm0-3h3v1H2v-1zm1-6.5c-.28 0-.5-.22-.5-.5s.22-.5.5-.5.5.22.5.5-.22.5-.5.5zM5 8h2.5V7H3v1.25C3.97 8.26 5 9.03 5 10v1h1V9.45c-.73-.26-1.05-.75-1.2-1.45zM7 14h5v-2H7v2zm5-4h-5V8h5v2zm-7 13h14v-2H5v2zm0-6h14v-2H5v2zM5 5v2h14V5H5z"
      />
    </Ic>
  )
}

export function IconAlignLeft() {
  return (
    <Ic>
      <path fill="currentColor" d="M15 15H3v2h12v-2zm0-8H3v2h12V7zM3 13h18v-2H3v2zm0 8h18v-2H3v2zM3 3v2h18V3H3z" />
    </Ic>
  )
}

export function IconAlignCenter() {
  return (
    <Ic>
      <path fill="currentColor" d="M7 15v2h10v-2H7zm-4 6h18v-2H3v2zM7 11v2h10v-2H7zM3 5v2h18V5H3zm4 6h10V9H7v2zM7 3v2h10V3H7z" />
    </Ic>
  )
}

export function IconAlignRight() {
  return (
    <Ic>
      <path fill="currentColor" d="M3 21h18v-2H3v2zM9 17h12v-2H9v2zM3 13h18v-2H3v2zM9 9h12V7H9v2zM3 3v2h18V3H3z" />
    </Ic>
  )
}

/** “A” with colored underline bar (Google Docs–style text color). */
export function IconTextColor({ underlineHex }: { underlineHex: string }) {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M10.62 3.79L5.29 17h2.38l1.12-3.12h5.38l1.12 3.12h2.38L11.38 3.79h-.76zm-.67 8.65l1.97-5.48 1.97 5.48h-3.94z"
      />
      <rect x="3.5" y="18.75" width="17" height="2.75" rx="0.35" fill={underlineHex} />
    </Ic>
  )
}

export function IconHighlight({ fillHex }: { fillHex: string }) {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M16.56 8.94L7.62 0 6.21 1.41l2.38 2.38-5.15 5.15c-.59.59-.59 1.54 0 2.12l5.5 5.5c.29.29.68.44 1.06.44s.77-.15 1.06-.44l5.5-5.5c.59-.58.59-1.53 0-2.12zM5.21 10L10 5.21 14.79 10H5.21z"
      />
      <path fill={fillHex} d="M17 16.5c0-1.33-2-3.5-2-3.5s-2 2.17-2 3.5c0 1.1.9 2 2 2s2-.9 2-2z" />
    </Ic>
  )
}

export function IconXMark() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"
      />
    </Ic>
  )
}

export function IconClearFormat() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M3 5.5v3h2.59l4.5 11.09L4.79 14.5H3v3h4.79l1.29-3.09L12 5.5H3zM18.5 7.5L15 4l-1.09 1.09L17.31 8.5l-3.4 3.4 1.09 1.09 4.5-4.5L18.5 7.5z"
      />
    </Ic>
  )
}

export function IconLink() {
  return (
    <Ic>
      <path
        fill="currentColor"
        d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"
      />
    </Ic>
  )
}

export function IconParagraph() {
  return (
    <Ic>
      <path fill="currentColor" d="M3 18h12v-2H3v2zM3 6v2h18V6H3zm0 7h18v-2H3v2z" />
    </Ic>
  )
}

export function IconHeading({ level }: { level: 2 | 3 | 4 }) {
  const d =
    level === 2
      ? 'M7 5v14H5V5h2zm4 0h7v2h-2.5v10H13V7H11V5z'
      : level === 3
        ? 'M7 5v14H5V5h2zm6 0h6v2h-2v10h-2V7h-2V5z'
        : 'M7 5v14H5V5h2zm5 0h5v2h-1.5v10H14V7h-1.5V5z'
  return (
    <Ic>
      <path fill="currentColor" d={d} />
    </Ic>
  )
}
