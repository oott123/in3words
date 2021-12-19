import { json } from 'remix'

if (!globalThis.blogApiGetCache) {
  globalThis.blogApiGetCache = new Map()
}

const BASE_URL = process.env.BASE_URL || 'https://best33.com'

export async function forwardRequest(request: Request) {
  const hopByHopHeaders = [
    'Keep-Alive',
    'Transfer-Encoding',
    'TE',
    'Connection',
    'Trailer',
    'Upgrade',
    'Proxy-Authorization',
    'Proxy-Authenticate',
  ]
  const oldUrl = new URL(request.url)
  const newUrl = new URL(request.url.substring(oldUrl.origin.length), BASE_URL)
  const newHeaders = new Headers(request.headers)

  newHeaders.delete('Host')
  newHeaders.delete('Cookie')

  newHeaders.set('Host', newUrl.host)
  for (const h of hopByHopHeaders) {
    newHeaders.delete(h)
  }

  const resp = await fetch(newUrl.toString(), {
    headers: newHeaders,
    body: request.body,
    method: request.method,
  })

  resp.headers.delete('Set-Cookie')

  // if (resp.status === 404) {
  //   throw createErrorResponse('找不到该页面', 'unknown_route', 404, request.url)
  // }

  return resp
}

export async function post<T = any>(
  path: string,
  body: any,
  prefix = '/wp-json/wp/v2',
): Promise<T> {
  const url = `${BASE_URL}${prefix}${path}`

  const resp = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })
  const result = await response(resp)

  return result
}

export async function get<T = any>(
  path: string,
  query?: Record<string, any>,
  prefix = '/wp-json/wp/v2',
): Promise<T> {
  const url = buildUrl(query, path, prefix)

  if (globalThis.blogApiGetCache!.has(url)) {
    // await sleep(2000)
    return globalThis.blogApiGetCache!.get(url)
  }

  const resp = await fetch(url)
  const result = await response(resp)
  globalThis.blogApiGetCache!.set(url, result)
  return result
}

export async function getList<T = any[]>(
  path: string,
  query?: Record<string, any>,
  prefix = '/wp-json/wp/v2',
): Promise<{ items: T; total: number; totalPages: number }> {
  const url = buildUrl(query, path, prefix)
  const cacheKey = `list_${url}`

  if (globalThis.blogApiGetCache!.has(cacheKey)) {
    return globalThis.blogApiGetCache!.get(cacheKey)
  }

  const resp = await fetch(url)
  const items = await response(resp)
  const total = Number(resp.headers.get('X-WP-Total'))
  const totalPages = Number(resp.headers.get('X-WP-TotalPages'))
  const data = { items, total, totalPages }

  globalThis.blogApiGetCache!.set(cacheKey, data)
  return data
}

async function response(resp: Response) {
  const data = await resp.json()
  if (!resp.ok && data) {
    throw createErrorResponse(data.message, data.code, resp.status, resp.url)
  }
  if (!resp.ok) {
    throw new CmsError('Unknown error', 'unknown', resp.status, resp.url)
  }

  return data
}

export type CmsErrorResponse = Response & {
  _isCmsError: true
  message: string
  code: string
  status: number
  url?: string
}

export function createErrorResponse(
  message: string,
  code: string,
  status = 500,
  url?: string,
) {
  const resp = json(
    { message, code, url, _isCmsError: true },
    status,
  ) as CmsErrorResponse
  resp._isCmsError = true
  resp.message = message
  resp.code = code
  resp.status = status
  if (url) resp.url = url

  return resp
}

export function isErrorResponse(obj: any): obj is CmsErrorResponse {
  return obj._isCmsError
}

export class CmsError extends Error {
  public constructor(
    message: string,
    public readonly code: string,
    public readonly status?: number,
    public readonly url?: string,
  ) {
    super(message)
    this.name = 'CmsError'
  }
}

function buildUrl(
  query: Record<string, any> | undefined,
  path: string,
  prefix: string,
) {
  const queryString = query
    ? `?${Object.entries(query)
        .map((x) => x.map(encodeURIComponent).join('='))
        .sort()
        .join('&')}`
    : ''
  const url = `${BASE_URL}${prefix}${path}${queryString}`
  return url
}

function sleep(time: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, time))
}

export function parseGmt(date: string): string {
  return new Date(`${date}Z`).toISOString()
}

export const voidTags = [
  'area',
  'base',
  'br',
  'col',
  'embed',
  'hr',
  'img',
  'input',
  'keygen',
  'link',
  'meta',
  'param',
  'source',
  'track',
  'wbr',
]

export function encodeHtmlText(str: string) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

export function encodeHtmlAttr(str: string) {
  return encodeHtmlText(str).replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}

export function replaceMediaUrl(url: string) {
  const cdnUrl =
    process.env.CDN_URL ||
    process.env.BASE_URL ||
    'https://cdn-best33-com.oott123.com/'

  return url
    .replace(/^https?:\/\/(www\.)?best33\.com\//, `${cdnUrl}/`)
    .replace(/^https?:\/\/best33\.b0\.upaiyun\.com\//, `${cdnUrl}/`)
    .replace(
      /^https?:\/\/(\w+)\.gravatar\.com\/avatar\//,
      'https://dn-qiniu-avatar.qbox.me/avatar/',
    )
}
