import { renderToString } from 'react-dom/server'
import { RemixServer } from 'remix'
import type { EntryContext } from 'remix'
import jsesc from 'jsesc'

// eslint-disable-next-line @typescript-eslint/no-var-requires
require('dotenv').config()

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  remixContext: EntryContext,
) {
  let markup = renderToString(
    <RemixServer context={remixContext} url={request.url} />,
  )

  if (remixContext.serverHandoffString) {
    const str = `<script>window.__remixContext = ${remixContext.serverHandoffString};</script>`
    const find = markup.indexOf(str)
    const replace = `<script type="application/json" id="in3wordsContext">${jsesc(
      eval(`(${remixContext.serverHandoffString})`),
      {
        isScriptContext: true,
        json: true,
        minimal: true,
      },
    )}</script><script>window.__remixContext = JSON.parse(document.getElementById('in3wordsContext').textContent);document.body.removeChild(document.currentScript);</script>`
    markup =
      markup.substring(0, find) + replace + markup.substring(find + str.length)
  }

  responseHeaders.set('Content-Type', 'text/html')

  return new Response('<!DOCTYPE html>' + markup, {
    status: responseStatusCode,
    headers: responseHeaders,
  })
}
