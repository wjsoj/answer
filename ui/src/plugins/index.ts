
const loadCaptchaBasic = () => import('captcha_basic').then(module => module.default);
export const basic_captcha = loadCaptchaBasic
const loadEditorFormula = () => import('editor-formula').then(module => module.default);
export const formula_editor = loadEditorFormula
const loadQuickLinks = () => import('quick-links').then(module => module.default);
export const quick_links = loadQuickLinks
const loadRenderMarkdownCodehighlight = () => import('render-markdown-codehighlight').then(module => module.default);
export const render_markdown_codehighlight = loadRenderMarkdownCodehighlight