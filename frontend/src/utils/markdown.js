import MarkdownIt from 'markdown-it'
import createDOMPurify from 'dompurify'

const markdown = new MarkdownIt({
  html: false,
  breaks: true,
  linkify: true
})

const defaultLinkOpenRenderer = markdown.renderer.rules.link_open
let domPurifyInstance = null

markdown.renderer.rules.link_open = (tokens, idx, options, env, self) => {
  const token = tokens[idx]
  token.attrSet('target', '_blank')
  token.attrSet('rel', 'noopener noreferrer')

  if (defaultLinkOpenRenderer) {
    return defaultLinkOpenRenderer(tokens, idx, options, env, self)
  }

  return self.renderToken(tokens, idx, options)
}

const getDOMPurify = () => {
  if (domPurifyInstance) {
    return domPurifyInstance
  }

  const browserWindow = globalThis.window

  if (!browserWindow) {
    return null
  }

  domPurifyInstance = createDOMPurify(browserWindow)
  return domPurifyInstance
}

export const renderMarkdownToSafeHtml = (content) => {
  const source = typeof content === 'string' ? content : ''
  const renderedHtml = markdown.render(source)
  const DOMPurify = getDOMPurify()

  if (!DOMPurify) {
    throw new Error('renderMarkdownToSafeHtml requires a browser window')
  }

  return DOMPurify.sanitize(renderedHtml, {
    ADD_ATTR: ['target', 'rel']
  })
}
