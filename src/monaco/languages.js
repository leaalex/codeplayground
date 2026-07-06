// Eager-load Monaco basic language grammars so production chunks load reliably.
export function preloadMonacoLanguages() {
  return Promise.all([
    import('monaco-editor/esm/vs/basic-languages/go/go.contribution.js'),
    import('monaco-editor/esm/vs/basic-languages/python/python.contribution.js'),
    import('monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js'),
    import('monaco-editor/esm/vs/basic-languages/go/go.js'),
    import('monaco-editor/esm/vs/basic-languages/python/python.js'),
    import('monaco-editor/esm/vs/basic-languages/markdown/markdown.js'),
  ])
}
