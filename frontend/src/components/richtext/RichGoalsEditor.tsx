import { useEditor, EditorContent, type Editor } from '@tiptap/react'
import { useId, type ReactNode } from 'react'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle, FontSize, BackgroundColor } from '@tiptap/extension-text-style'
import { Color } from '@tiptap/extension-color'
import { FontFamily } from '@tiptap/extension-font-family'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import Placeholder from '@tiptap/extension-placeholder'
import { htmlToStorage, storageToHtml } from './richGoalsDoc'
import {
  IconAlignCenter,
  IconAlignLeft,
  IconAlignRight,
  IconBold,
  IconBulletList,
  IconClearFormat,
  IconHeading,
  IconHighlight,
  IconItalic,
  IconLink,
  IconNumberedList,
  IconParagraph,
  IconRedo,
  IconStrikethrough,
  IconSubscript,
  IconSuperscript,
  IconTextColor,
  IconUnderline,
  IconUndo,
  IconXMark,
} from './DocToolbarIcons'

type Props = {
  /** Field label (legend). */
  label: string
  /** Newline-separated bullets (HTML per line or legacy plain). */
  value: string
  onChange: (value: string) => void
  /** Bump after loading from API so the editor picks up server content. */
  resetKey: number
  placeholder?: string
  /** View-only rich text (no toolbar, not editable). */
  readOnly?: boolean
}

type Opt = { label: string; value: string }

/** Google Docs–style font menu (web-safe / system stacks). */
const FONT_CHOICES: Opt[] = [
  { label: 'Default', value: '' },
  { label: 'Cormorant Garamond', value: 'var(--font-serif), Georgia, "Times New Roman", serif' },
  { label: 'Source Sans 3', value: 'var(--font-sans), system-ui, sans-serif' },
  { label: 'Arial', value: 'Arial, Helvetica, sans-serif' },
  { label: 'Calibri', value: 'Calibri, "Segoe UI", "Helvetica Neue", Arial, sans-serif' },
  { label: 'Comic Sans MS', value: '"Comic Sans MS", "Comic Sans", cursive, sans-serif' },
  { label: 'Courier New', value: '"Courier New", Courier, monospace' },
  { label: 'Georgia', value: 'Georgia, serif' },
  { label: 'Garamond', value: 'Garamond, "Times New Roman", Times, serif' },
  { label: 'Impact', value: 'Impact, Haettenschweiler, "Arial Narrow", sans-serif' },
  { label: 'Palatino', value: '"Palatino Linotype", Palatino, "Book Antiqua", serif' },
  { label: 'Segoe UI', value: '"Segoe UI", Tahoma, Geneva, Verdana, sans-serif' },
  { label: 'Times New Roman', value: '"Times New Roman", Times, serif' },
  { label: 'Trebuchet MS', value: '"Trebuchet MS", sans-serif' },
  { label: 'Verdana', value: 'Verdana, Geneva, sans-serif' },
]

/** Point sizes like Google Docs (stored as CSS pt). */
const FONT_SIZE_PT = [8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 26, 28, 36, 48, 72]

function buildFontSizeOptions(current: string | undefined): Opt[] {
  const base: Opt[] = [
    { label: 'Default', value: '' },
    ...FONT_SIZE_PT.map((n) => ({ label: String(n), value: `${n}pt` })),
  ]
  const set = new Set(base.map((o) => o.value))
  if (current && current !== '' && !set.has(current)) {
    return [...base, { label: `${current} (custom)`, value: current }]
  }
  return base
}

function buildFontFamilyOptions(current: string | undefined): Opt[] {
  const set = new Set(FONT_CHOICES.map((o) => o.value))
  if (current && current !== '' && !set.has(current)) {
    return [...FONT_CHOICES, { label: `${current.slice(0, 28)}…`, value: current }]
  }
  return FONT_CHOICES
}

/** For `<input type="color">` when stored style is rgb() or named color. */
function normalizeHex(input: string | null | undefined): string | undefined {
  if (!input) return undefined
  const s = input.trim()
  if (/^#[0-9a-fA-F]{6}$/i.test(s)) return s.toLowerCase()
  if (/^#[0-9a-fA-F]{3}$/i.test(s)) {
    const h = s.slice(1)
    return (`#${h[0]}${h[0]}${h[1]}${h[1]}${h[2]}${h[2]}`).toLowerCase()
  }
  const m = s.match(/^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i)
  if (m) {
    const r = Number(m[1])
    const g = Number(m[2])
    const b = Number(m[3])
    return `#${[r, g, b].map((x) => x.toString(16).padStart(2, '0')).join('')}`
  }
  return undefined
}

function ToolbarBtn({
  title,
  onClick,
  isActive,
  disabled,
  children,
}: {
  title: string
  onClick: () => void
  isActive?: boolean
  disabled?: boolean
  children: ReactNode
}) {
  return (
    <button
      type="button"
      className={`rich-toolbar-btn${isActive ? ' is-active' : ''}`}
      title={title}
      aria-label={title}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  )
}

function ToolbarSep() {
  return <span className="rich-toolbar__sep" aria-hidden />
}

function Toolbar({ editor }: { editor: Editor | null }) {
  if (!editor) return null

  const attrs = editor.getAttributes('textStyle')
  const textHex = normalizeHex(attrs.color) || '#1b1812'
  const hiHex = normalizeHex(attrs.backgroundColor) || '#fff59d'

  const setLink = () => {
    const prev = editor.getAttributes('link').href as string | undefined
    const url = window.prompt('Link URL (e.g. /monthly-challenges or https://…)', prev ?? 'https://')
    if (url === null) return
    const t = url.trim()
    if (t === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }
    editor.chain().focus().extendMarkRange('link').setLink({ href: t }).run()
  }

  const clearFormatting = () => {
    editor.chain().focus().unsetAllMarks().setParagraph().run()
  }

  return (
    <div className="rich-toolbar rich-toolbar--docs" role="toolbar" aria-label="Formatting">
      <div className="rich-toolbar__sheet">
        <div className="rich-toolbar__line">
          <ToolbarBtn
            title="Undo"
            onClick={() => editor.chain().focus().undo().run()}
            disabled={!editor.can().undo()}
          >
            <IconUndo />
          </ToolbarBtn>
          <ToolbarBtn
            title="Redo"
            onClick={() => editor.chain().focus().redo().run()}
            disabled={!editor.can().redo()}
          >
            <IconRedo />
          </ToolbarBtn>

          <ToolbarSep />

          <label className="rich-toolbar-select-shell" title="Font">
            <span className="sr-only">Font</span>
            <select
              className="rich-toolbar-select rich-toolbar-select--font"
              value={attrs.fontFamily || ''}
              onChange={(e) => {
                const v = e.target.value
                if (!v) editor.chain().focus().unsetFontFamily().run()
                else editor.chain().focus().setFontFamily(v).run()
              }}
            >
              {buildFontFamilyOptions(attrs.fontFamily || undefined).map((f) => (
                <option key={f.value || 'default'} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
          </label>

          <label className="rich-toolbar-select-shell" title="Font size">
            <span className="sr-only">Font size</span>
            <select
              className="rich-toolbar-select rich-toolbar-select--size"
              value={attrs.fontSize || ''}
              onChange={(e) => {
                const v = e.target.value
                if (!v) editor.chain().focus().unsetFontSize().run()
                else editor.chain().focus().setFontSize(v).run()
              }}
            >
              {buildFontSizeOptions(attrs.fontSize || undefined).map((f) => (
                <option key={f.label + f.value} value={f.value}>
                  {f.label}
                </option>
              ))}
            </select>
          </label>

          <ToolbarSep />

          <ToolbarBtn title="Bold" isActive={editor.isActive('bold')} onClick={() => editor.chain().focus().toggleBold().run()}>
            <IconBold />
          </ToolbarBtn>
          <ToolbarBtn
            title="Italic"
            isActive={editor.isActive('italic')}
            onClick={() => editor.chain().focus().toggleItalic().run()}
          >
            <IconItalic />
          </ToolbarBtn>
          <ToolbarBtn
            title="Underline"
            isActive={editor.isActive('underline')}
            onClick={() => editor.chain().focus().toggleUnderline().run()}
          >
            <IconUnderline />
          </ToolbarBtn>
          <ToolbarBtn
            title="Strikethrough"
            isActive={editor.isActive('strike')}
            onClick={() => editor.chain().focus().toggleStrike().run()}
          >
            <IconStrikethrough />
          </ToolbarBtn>
          <span className="rich-toolbar__cluster">
            <ToolbarBtn
              title="Subscript"
              isActive={editor.isActive('subscript')}
              onClick={() => editor.chain().focus().toggleSubscript().run()}
            >
              <IconSubscript />
            </ToolbarBtn>
            <ToolbarBtn
              title="Superscript"
              isActive={editor.isActive('superscript')}
              onClick={() => editor.chain().focus().toggleSuperscript().run()}
            >
              <IconSuperscript />
            </ToolbarBtn>
          </span>

          <ToolbarSep />

          <span className="rich-toolbar__cluster">
            <ToolbarBtn
              title="Align left"
              isActive={editor.isActive({ textAlign: 'left' })}
              onClick={() => editor.chain().focus().setTextAlign('left').run()}
            >
              <IconAlignLeft />
            </ToolbarBtn>
            <ToolbarBtn
              title="Align center"
              isActive={editor.isActive({ textAlign: 'center' })}
              onClick={() => editor.chain().focus().setTextAlign('center').run()}
            >
              <IconAlignCenter />
            </ToolbarBtn>
            <ToolbarBtn
              title="Align right"
              isActive={editor.isActive({ textAlign: 'right' })}
              onClick={() => editor.chain().focus().setTextAlign('right').run()}
            >
              <IconAlignRight />
            </ToolbarBtn>
          </span>

          <ToolbarSep />

          <ToolbarBtn
            title="Bulleted list"
            isActive={editor.isActive('bulletList')}
            onClick={() => editor.chain().focus().toggleBulletList().run()}
          >
            <IconBulletList />
          </ToolbarBtn>
          <ToolbarBtn
            title="Numbered list"
            isActive={editor.isActive('orderedList')}
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
          >
            <IconNumberedList />
          </ToolbarBtn>
        </div>

        <div className="rich-toolbar__line">
          <ToolbarBtn
            title="Heading 2"
            isActive={editor.isActive('heading', { level: 2 })}
            onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
          >
            <IconHeading level={2} />
          </ToolbarBtn>
          <ToolbarBtn
            title="Heading 3"
            isActive={editor.isActive('heading', { level: 3 })}
            onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
          >
            <IconHeading level={3} />
          </ToolbarBtn>
          <ToolbarBtn
            title="Heading 4"
            isActive={editor.isActive('heading', { level: 4 })}
            onClick={() => editor.chain().focus().toggleHeading({ level: 4 }).run()}
          >
            <IconHeading level={4} />
          </ToolbarBtn>
          <ToolbarBtn
            title="Normal text"
            isActive={!editor.isActive('heading')}
            onClick={() => editor.chain().focus().setParagraph().run()}
          >
            <IconParagraph />
          </ToolbarBtn>

          <ToolbarSep />

          <label className="rich-toolbar-btn rich-toolbar-btn--color" title="Text color">
            <span className="sr-only">Text color</span>
            <input
              type="color"
              className="rich-toolbar-color-input"
              value={textHex}
              onChange={(e) => editor.chain().focus().setColor(e.target.value).run()}
            />
            <IconTextColor underlineHex={textHex} />
          </label>
          <ToolbarBtn title="Remove text color" onClick={() => editor.chain().focus().unsetColor().run()}>
            <IconXMark />
          </ToolbarBtn>

          <label className="rich-toolbar-btn rich-toolbar-btn--color" title="Highlight">
            <span className="sr-only">Highlight</span>
            <input
              type="color"
              className="rich-toolbar-color-input"
              value={hiHex}
              onChange={(e) => editor.chain().focus().setBackgroundColor(e.target.value).run()}
            />
            <IconHighlight fillHex={hiHex} />
          </label>
          <ToolbarBtn title="Remove highlight" onClick={() => editor.chain().focus().unsetBackgroundColor().run()}>
            <IconXMark />
          </ToolbarBtn>

          <ToolbarSep />

          <span className="rich-toolbar__cluster">
            <ToolbarBtn title="Clear formatting" onClick={clearFormatting}>
              <IconClearFormat />
            </ToolbarBtn>
            <ToolbarBtn title="Insert link" onClick={() => void setLink()}>
              <IconLink />
            </ToolbarBtn>
          </span>
        </div>
      </div>
    </div>
  )
}

export function RichGoalsEditor({
  label,
  value,
  onChange,
  resetKey,
  placeholder,
  readOnly = false,
}: Props) {
  const titleId = useId()

  const editor = useEditor(
    {
      editable: !readOnly,
      extensions: [
        StarterKit.configure({
          heading: { levels: [2, 3, 4] },
        }),
        Underline,
        Subscript,
        Superscript,
        Link.configure({
          openOnClick: false,
          HTMLAttributes: { class: 'inline-markup-link' },
        }),
        TextAlign.configure({
          types: ['heading', 'paragraph'],
        }),
        TextStyle,
        FontSize.configure({ types: ['textStyle'] }),
        Color.configure({ types: ['textStyle'] }),
        BackgroundColor.configure({ types: ['textStyle'] }),
        FontFamily.configure({ types: ['textStyle'] }),
        Placeholder.configure({
          placeholder: placeholder ?? 'One bullet per row — use toolbar for fonts, size (pt), links…',
        }),
      ],
      content: storageToHtml(value),
      editorProps: {
        attributes: {
          class: 'rich-goals-prose',
          spellCheck: 'true',
        },
      },
      onUpdate: ({ editor: ed }) => {
        if (readOnly) return
        onChange(htmlToStorage(ed.getHTML()))
      },
    },
    [resetKey, readOnly],
  )

  return (
    <section
      className={`rich-goals-block${readOnly ? ' rich-goals-block--readonly' : ''}`}
      aria-labelledby={titleId}
    >
      <h3 id={titleId} className="rich-goals-block__title">
        {label}
      </h3>
      {readOnly ? null : <Toolbar editor={editor} />}
      <div className="rich-goals-surface">
        <EditorContent editor={editor} />
      </div>
    </section>
  )
}
