const GO_TEMPLATE = `package main

import "fmt"

func main() {
	fmt.Println("Hello!")
}
`

const PYTHON_TEMPLATE = `print("Hello!")
`

export function detectLanguage(filename) {
  if (filename?.endsWith('.py')) return 'python'
  return 'go'
}

export function defaultTemplate(lang) {
  return lang === 'python' ? PYTHON_TEMPLATE : GO_TEMPLATE
}

export function defaultFileName(lang) {
  return lang === 'python' ? 'untitled.py' : 'untitled.go'
}

export function fileExtension(lang) {
  return lang === 'python' ? '.py' : '.go'
}

export function ensureExtension(name, lang) {
  const ext = fileExtension(lang)
  if (name.endsWith(ext)) return name
  const otherExt = lang === 'python' ? '.go' : '.py'
  if (name.endsWith(otherExt)) return name.slice(0, -otherExt.length) + ext
  return name + ext
}

export function preserveExtension(name, currentFilename) {
  const lang = detectLanguage(currentFilename)
  return ensureExtension(name.trim() || defaultFileName(lang), lang)
}

export function monacoLanguage(lang) {
  return lang === 'python' ? 'python' : 'go'
}

export function exportMimeType(lang) {
  return lang === 'python' ? 'text/x-python' : 'text/x-go'
}

export function languageLabel(lang) {
  return lang === 'python' ? 'Python' : 'Go'
}
