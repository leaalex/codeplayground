const GO_TEMPLATE = `package main

import "fmt"

func main() {
	fmt.Println("Hello!")
}
`

const PYTHON_TEMPLATE = `print("Hello!")
`

const MARKDOWN_TEMPLATE = `# Assignment

Describe the task here.

## Requirements

- Item 1
- Item 2
`

export function isMarkdownFile(filename) {
  return filename?.toLowerCase().endsWith('.md')
}

export function isCodeFile(filename) {
  if (!filename) return false
  const n = filename.toLowerCase()
  return n.endsWith('.go') || n.endsWith('.py')
}

export function detectLanguage(filename) {
  if (isMarkdownFile(filename)) return 'markdown'
  const n = filename?.toLowerCase() || ''
  if (n.endsWith('.py')) return 'python'
  if (n.endsWith('.go')) return 'go'
  return 'go'
}

export function defaultTemplate(lang) {
  if (lang === 'python') return PYTHON_TEMPLATE
  if (lang === 'markdown') return MARKDOWN_TEMPLATE
  return GO_TEMPLATE
}

export function defaultFileName(lang) {
  if (lang === 'python') return 'untitled.py'
  if (lang === 'markdown') return 'untitled.md'
  return 'untitled.go'
}

export function fileExtension(lang) {
  if (lang === 'python') return '.py'
  if (lang === 'markdown') return '.md'
  return '.go'
}

export function ensureExtension(name, lang) {
  const ext = fileExtension(lang)
  if (name.endsWith(ext)) return name
  const lower = name.toLowerCase()
  for (const other of ['.go', '.py', '.md']) {
    if (other !== ext && lower.endsWith(other)) {
      return name.slice(0, -other.length) + ext
    }
  }
  return name + ext
}

export function preserveExtension(name, currentFilename) {
  const lang = detectLanguage(currentFilename)
  return ensureExtension(name.trim() || defaultFileName(lang), lang)
}

export function monacoLanguage(lang) {
  if (lang === 'python') return 'python'
  if (lang === 'markdown') return 'markdown'
  return 'go'
}

export function exportMimeType(lang) {
  if (lang === 'python') return 'text/x-python'
  if (lang === 'markdown') return 'text/markdown'
  return 'text/x-go'
}

export function languageLabel(lang) {
  if (lang === 'python') return 'Python'
  if (lang === 'markdown') return 'Markdown'
  return 'Go'
}
