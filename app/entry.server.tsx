import 'dotenv/config'
import { PassThrough } from 'node:stream'

import type { EntryContext } from 'react-router'
import { createReadableStreamFromReadable } from '@react-router/node'
import { ServerRouter } from 'react-router'
import { isbot } from 'isbot'
import type { RenderToPipeableStreamOptions } from 'react-dom/server'
import { renderToPipeableStream } from 'react-dom/server'

export const streamTimeout = 5_000

function setDocumentHeaders(request: Request, responseHeaders: Headers) {
  responseHeaders.set('Content-Type', 'text/html')
  if (process.env.CDN_URL) {
    responseHeaders.append('Link', `<${process.env.CDN_URL}>; rel=preconnect`)
  }
  if (process.env.PUBLIC_URL) {
    responseHeaders.append(
      'Link',
      `<${process.env.PUBLIC_URL}>; rel=alternate; type=application/rss+xml`,
    )
  }
  if (request.method === 'GET') {
    responseHeaders.append('Cache-Control', 'public, max-age=60, s-maxage=3600')
  }
}

export default function handleRequest(
  request: Request,
  responseStatusCode: number,
  responseHeaders: Headers,
  routerContext: EntryContext,
) {
  // https://httpwg.org/specs/rfc9110.html#HEAD
  if (request.method.toUpperCase() === 'HEAD') {
    setDocumentHeaders(request, responseHeaders)
    return new Response(null, {
      status: responseStatusCode,
      headers: responseHeaders,
    })
  }

  return new Promise<Response>((resolve, reject) => {
    let shellRendered = false
    const userAgent = request.headers.get('user-agent')

    // 爬虫需要等待全部内容渲染完成后再返回
    const readyOption: keyof RenderToPipeableStreamOptions =
      (userAgent && isbot(userAgent)) || routerContext.isSpaMode
        ? 'onAllReady'
        : 'onShellReady'

    let timeoutId: ReturnType<typeof setTimeout> | undefined = setTimeout(
      () => abort(),
      streamTimeout + 1000,
    )

    const { pipe, abort } = renderToPipeableStream(
      <ServerRouter context={routerContext} url={request.url} />,
      {
        [readyOption]() {
          shellRendered = true
          const body = new PassThrough({
            final(callback) {
              clearTimeout(timeoutId)
              timeoutId = undefined
              callback()
            },
          })
          const stream = createReadableStreamFromReadable(body)

          setDocumentHeaders(request, responseHeaders)

          pipe(body)

          resolve(
            new Response(stream, {
              headers: responseHeaders,
              status: responseStatusCode,
            }),
          )
        },
        onShellError(error: unknown) {
          // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
          reject(error)
        },
        onError(error: unknown) {
          responseStatusCode = 500
          if (shellRendered) {
            console.error(error)
          }
        },
      },
    )
  })
}
