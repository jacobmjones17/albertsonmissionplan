import { useEffect } from 'react'

const DEFAULT_MESSAGE = 'You have unsaved changes. Leave this page and lose them?'

/**
 * Warn before leaving with unsaved edits: the browser's own prompt on reload/close/external
 * navigation (beforeunload), and a confirm() before in-app link clicks. The app uses
 * <BrowserRouter> (not a data router), so React Router's useBlocker isn't available; catching
 * link clicks in the capture phase covers the leader nav, header, and footer links.
 */
export function useUnsavedChangesWarning(dirty: boolean, message = DEFAULT_MESSAGE) {
  useEffect(() => {
    if (!dirty) return

    function onBeforeUnload(e: BeforeUnloadEvent) {
      e.preventDefault()
      // Required by some browsers to show the prompt.
      e.returnValue = ''
    }

    function onClickCapture(e: MouseEvent) {
      if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
        return
      }
      const target = e.target
      if (!(target instanceof Element)) return
      const a = target.closest('a[href]')
      if (!(a instanceof HTMLAnchorElement)) return
      if (a.target && a.target !== '_self') return
      const url = new URL(a.href, window.location.href)
      if (url.origin !== window.location.origin) return // beforeunload covers external links
      if (url.pathname === window.location.pathname && url.search === window.location.search) return
      if (!window.confirm(message)) {
        e.preventDefault()
        e.stopPropagation()
      }
    }

    window.addEventListener('beforeunload', onBeforeUnload)
    document.addEventListener('click', onClickCapture, true)
    return () => {
      window.removeEventListener('beforeunload', onBeforeUnload)
      document.removeEventListener('click', onClickCapture, true)
    }
  }, [dirty, message])
}
