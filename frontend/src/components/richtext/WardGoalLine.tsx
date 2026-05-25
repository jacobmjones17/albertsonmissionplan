import { useMemo } from 'react'
import DOMPurify, { type Config as DomPurifyConfig } from 'dompurify'
import { renderInlineMarkup } from '../../lib/inlineMarkup'

const PURIFY_OPTS: DomPurifyConfig = {
  ALLOWED_TAGS: [
    'strong',
    'b',
    'em',
    'i',
    'u',
    's',
    'strike',
    'del',
    'a',
    'span',
    'br',
    'p',
    'h2',
    'h3',
    'h4',
    'ul',
    'ol',
    'li',
    'sub',
    'sup',
    'mark',
    'code',
    'blockquote',
  ],
  ALLOWED_ATTR: ['href', 'title', 'rel', 'target', 'class', 'style'],
  ADD_ATTR: ['target'],
}

/** One ward-plan bullet: HTML from the rich editor, or legacy plain / **markdown**. */
export function WardGoalLine({ content }: { content: string }) {
  const html = useMemo(() => {
    if (!/<[a-z][\s\S]*>/i.test(content)) {
      return null
    }
    return DOMPurify.sanitize(content, PURIFY_OPTS)
  }, [content])

  if (html) {
    return (
      <div
        className="ward-rich-line"
        // Block wrapper: editor output may include <p>, lists, etc. — invalid inside <span>.
        dangerouslySetInnerHTML={{ __html: html }}
      />
    )
  }
  return <>{renderInlineMarkup(content)}</>
}
