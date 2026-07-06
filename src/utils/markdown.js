import { Marked } from 'marked'
import { markedHighlight } from 'marked-highlight'
import hljs from 'highlight.js/lib/core'
import go from 'highlight.js/lib/languages/go'
import python from 'highlight.js/lib/languages/python'

hljs.registerLanguage('go', go)
hljs.registerLanguage('python', python)

const LANG_ALIASES = {
  py: 'python',
  golang: 'go',
}

function resolveLanguage(lang) {
  if (!lang) return 'plaintext'
  const key = lang.trim().toLowerCase()
  const normalized = LANG_ALIASES[key] || key
  return hljs.getLanguage(normalized) ? normalized : 'plaintext'
}

export const markdown = new Marked(
  markedHighlight({
    emptyLangClass: 'hljs',
    langPrefix: 'hljs language-',
    highlight(code, lang) {
      const language = resolveLanguage(lang)
      return hljs.highlight(code, { language }).value
    },
  })
)
