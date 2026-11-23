import { useEditor, useEditorState, EditorContent, type Editor } from '@tiptap/react'
import { useId, type ReactNode } from 'react'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import TextAlign from '@tiptap/extension-text-align'
import { TextStyle, BackgroundColor } from '@tiptap/extension-text-style'
import { Color } from '@tiptap/extension-color'
import Subscript from '@tiptap/extension-subscript'
import Superscript from '@tiptap/extension-superscript'
import Placeholder from '@tiptap/extension-placeholder'
import { htmlToStorage, storageToHtml } from './richGoalsDoc'
import {
  IconBold,
  IconBulletList,
  IconClearFormat,
  IconHeading,
  IconItalic,
  IconLink,
  IconNumberedList,
  IconRedo,
  IconUnderline,
  IconUndo,
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

function ToolbarBtn({
  label,
  shortcut,
  onClick,
  isActive,
  disabled,
  showText = false,
  children,
}: {
  /** Accessible name and tooltip, e.g. "Bold". */
  label: string
  shortcut?: string
  onClick: () => void
  isActive?: boolean
  disabled?: boolean
  /** Show the label next to the icon (for less obvious actions). */
  showText?: boolean
  children: ReactNode
}) {
  const tip = shortcut ? `${label} (${shortcut})` : label
  return (
    <button
      type="button"
      className={`rich-toolbar-btn${isActive ? ' is-active' : ''}${showText ? ' rich-toolbar-btn--text' : ''}`}
      title={tip}
      aria-label={showText ? undefined : label}
      aria-pressed={isActive === undefined ? undefined : isActive}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
      {showText ? <span className="rich-toolbar-btn__text">{label}</span> : null}
    </button>
  )
}

function ToolbarSep() {
  return <span className="rich-toolbar__sep" aria-hidden />
}

const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)
const mod = isMac ? '⌘' : 'Ctrl'

/**
 * Short, labeled toolbar: undo/redo, bold/italic/underline, link, bulleted/numbered sub-list,
 * heading, clear formatting. Font family / arbitrary size pickers were removed so goals keep the
 * site's typography; existing colors/alignment in saved content still load and render.
 */
function Toolbar({ editor }: { editor: Editor }) {
  const s = useEditorState({
    editor,
    selector: ({ editor: ed }) => ({
      bold: ed.isActive('bold'),
      italic: ed.isActive('italic'),
      underline: ed.isActive('underline'),
      link: ed.isActive('link'),
      heading: ed.isActive('heading'),
      ordered: ed.isActive('orderedList'),
      canUndo: ed.can().undo(),
      canRedo: ed.can().redo(),
    }),
  })

  const setLink = () => {
    const prev = editor.getAttributes('link').href as string | undefined
    const url = window.prompt(
      'Link address (for example /monthly-challenges or https://…). Leave empty to remove the link.',
      prev ?? 'https://',
    )
    if (url === null) return
    const t = url.trim()
    if (t === '' || t === 'https://') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }
    editor.chain().focus().extendMarkRange('link').setLink({ href: t }).run()
  }

  return (
    <div className="rich-toolbar rich-toolbar--simple" role="toolbar" aria-label="Formatting">
      <ToolbarBtn
        label="Undo"
        shortcut={`${mod}+Z`}
        onClick={() => editor.chain().focus().undo().run()}
        disabled={!s.canUndo}
      >
        <IconUndo />
      </ToolbarBtn>
      <ToolbarBtn
        label="Redo"
        shortcut={isMac ? '⌘+Shift+Z' : 'Ctrl+Y'}
        onClick={() => editor.chain().focus().redo().run()}
        disabled={!s.canRedo}
      >
        <IconRedo />
      </ToolbarBtn>
      <ToolbarSep />
      <ToolbarBtn
        label="Bold"
        shortcut={`${mod}+B`}
        isActive={s.bold}
        onClick={() => editor.chain().focus().toggleBold().run()}
      >
        <IconBold />
      </ToolbarBtn>
      <ToolbarBtn
        label="Italic"
        shortcut={`${mod}+I`}
        isActive={s.italic}
        onClick={() => editor.chain().focus().toggleItalic().run()}
      >
        <IconItalic />
      </ToolbarBtn>
      <ToolbarBtn
        label="Underline"
        shortcut={`${mod}+U`}
        isActive={s.underline}
        onClick={() => editor.chain().focus().toggleUnderline().run()}
      >
        <IconUnderline />
      </ToolbarBtn>
      <ToolbarSep />
      <ToolbarBtn label="Link" isActive={s.link} onClick={setLink} showText>
        <IconLink />
      </ToolbarBtn>
      <ToolbarBtn
        label="Sub-points"
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        showText
      >
        <IconBulletList />
      </ToolbarBtn>
      <ToolbarBtn
        label="Numbered"
        isActive={s.ordered}
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
        showText
      >
        <IconNumberedList />
      </ToolbarBtn>
      <ToolbarBtn
        label="Heading"
        isActive={s.heading}
        onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
        showText
      >
        <IconHeading level={3} />
      </ToolbarBtn>
      <ToolbarSep />
      <ToolbarBtn
        label="Clear formatting"
        onClick={() => editor.chain().focus().unsetAllMarks().setParagraph().run()}
        showText
      >
        <IconClearFormat />
      </ToolbarBtn>
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
          // StarterKit v3 bundles Link + Underline; we configure those separately below.
          link: false,
          underline: false,
        }),
        Underline,
        // Kept (no toolbar buttons) so previously saved formatting still loads and renders.
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
        Color.configure({ types: ['textStyle'] }),
        BackgroundColor.configure({ types: ['textStyle'] }),
        Placeholder.configure({
          placeholder: placeholder ?? 'Type a goal. Press Enter to start the next goal.',
        }),
      ],
      content: storageToHtml(value),
      editorProps: {
        attributes: {
          class: 'rich-goals-prose',
          spellCheck: 'true',
          'aria-labelledby': titleId,
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
      {readOnly || !editor ? null : <Toolbar editor={editor} />}
      <div className="rich-goals-surface">
        <EditorContent editor={editor} />
      </div>
    </section>
  )
}
