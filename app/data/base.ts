import * as cache from './cache'

if (!globalThis.blogApiGetCache) {
  globalThis.blogApiGetCache = new Map()
}

const BASE_URL = process.env.BASE_URL || 'https://best33.com'
const BASE_HOST = process.env.BASE_HOST

export async function feed() {
  const newUrl = new URL('/feed/', BASE_URL)
  try {
    const resp = await fetch(newUrl.toString(), {
      headers: {
        Host: BASE_HOST || newUrl.host,
      },
    })

    return proxyResponse(resp, ['Set-Cookie'])
  } catch (e) {
    console.error(`Request to ${newUrl.toString()} failed`, e)
    throw e
  }
}

export async function forwardRequest(
  request: Request,
  url = new URL(request.url),
  baseUrl?: string,
  baseHost?: string,
) {
  const removeHeaders = [
    // hop by hop headers
    'Keep-Alive',
    'Transfer-Encoding',
    'TE',
    'Connection',
    'Trailer',
    'Upgrade',
    'Proxy-Authorization',
    'Proxy-Authenticate',
  ]
  const newUrl = new URL(`${url.pathname}${url.search}`, baseUrl || BASE_URL)
  const newHeaders = new Headers(request.headers)

  newHeaders.delete('Host')
  newHeaders.delete('Cookie')

  newHeaders.set('Host', baseHost || BASE_HOST || newUrl.host)
  for (const h of removeHeaders) {
    newHeaders.delete(h)
  }

  // console.log(
  //   `Forwarding from ${url.toString()} to ${newUrl.toString()}`,
  //   newHeaders,
  // )
  try {
    const resp = await fetch(newUrl.toString(), {
      headers: newHeaders,
      body: request.body,
      method: request.method,
      redirect: 'manual',
    })

    // if (resp.status === 404) {
    //   throw createErrorResponse('找不到该页面', 'unknown_route', 404, request.url)
    // }

    return proxyResponse(resp, ['Set-Cookie'])
  } catch (e) {
    console.error(`Request to ${newUrl.toString()} failed`, e)
    throw e
  }
}

/**
 * fetch 返回的 Response 的 headers 是只读的，需要复制一份才能修改。
 * fetch 会自动解压响应体，所以同时去掉 Content-Encoding 和 Content-Length。
 */
export function proxyResponse(resp: Response, removeHeaders: string[] = []) {
  const headers = new Headers(resp.headers)
  for (const h of [...removeHeaders, 'Content-Encoding', 'Content-Length']) {
    headers.delete(h)
  }

  return new Response(resp.body, {
    status: resp.status,
    statusText: resp.statusText,
    headers,
  })
}

export type RequestOptions = {
  prefix?: string
  cacheGroup?: string
  cacheTtl?: number
  realIp?: string
  userAgent?: string
}

export async function post<T = any>(
  path: string,
  body: any,
  options?: RequestOptions,
): Promise<T> {
  const prefix = options?.prefix ?? '/wp-json/wp/v2'
  const url = `${BASE_URL}${prefix}${path}`

  const resp = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Commented-With': 'remix',
      ...(options?.realIp ? { 'X-Real-IP': options.realIp } : {}),
      ...(options?.userAgent ? { 'User-Agent': options.userAgent } : {}),
    },
    body: JSON.stringify(body),
  })
  const result = await response<T>(resp)

  return result.data
}

async function cachedGet<T = any>(
  path: string,
  query?: Record<string, any>,
  options?: RequestOptions,
): Promise<CmsData<T>> {
  const prefix = options?.prefix ?? '/wp-json/wp/v2'
  const url = buildUrl(query, path, prefix)
  const cacheKey = options?.cacheGroup ? `${options.cacheGroup}:${url}` : url

  const cached = cache.get(cacheKey)

  const resp = fetch(url)
  const result = response<T>(resp)

  const got = await new Promise<CmsData<T>>((resolve, reject) => {
    cached.then((data) => {
      if (data !== null) {
        resolve(data)
      }
    }, console.error)

    result.then((data) => {
      cache.set(cacheKey, data).catch(console.error)
      resolve(data)
    }, reject)
  })

  return got
}

export async function get<T = any>(
  path: string,
  query?: Record<string, any>,
  options?: RequestOptions,
): Promise<T> {
  return (await cachedGet(path, query, options)).data
}

export async function getList<T = any[]>(
  path: string,
  query?: Record<string, any>,
  options?: RequestOptions,
): Promise<{ items: T; total: number; totalPages: number }> {
  const data = await cachedGet<T>(path, query, options)
  return {
    items: data.data,
    total: data.total,
    totalPages: data.totalPages,
  }
}

type CmsData<T> = {
  data: T
  total: number
  totalPages: number
}

async function response<T = any>(
  respInput: Response | Promise<Response>,
): Promise<CmsData<T>> {
  const resp = await respInput
  const data = await resp.json()
  if (!resp.ok && data) {
    throw createErrorResponse(data.message, data.code, resp.status, resp.url)
  }
  if (!resp.ok) {
    throw createErrorResponse('Unknown error', 'unknown', resp.status, resp.url)
  }

  const total = Number(resp.headers.get('X-WP-Total') || 0)
  const totalPages = Number(resp.headers.get('X-WP-TotalPages') || 0)

  return { data, total, totalPages }
}

export type CmsErrorResponse = Response & {
  _isCmsError: true
  message: string
  code: string
}

export function createErrorResponse(
  message: string,
  code: string,
  status = 500,
  url?: string,
) {
  const resp = Response.json(
    { message, code, url, _isCmsError: true },
    { status },
  ) as CmsErrorResponse
  // Response 的 status 和 url 是只读的，已经通过构造参数和响应体传递
  resp._isCmsError = true
  resp.message = message
  resp.code = code

  return resp
}

export function isErrorResponse(obj: any): obj is CmsErrorResponse {
  return !!obj?._isCmsError
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
    process.env.CDN_URL || process.env.BASE_URL || 'https://best33.com/'

  // const gravatarUrl = https://dn-qiniu-avatar.qbox.me/avatar/
  // const gravatarUrl = 'https://gravatar.loli.net/avatar/'
  // const gravatarUrl = 'https://cravatar.cn/avatar/'
  // const gravatarUrl = '/avatar/'
  const gravatarUrl = 'https://www.libravatar.org/avatar/'

  return url
    .replace(/(^|\b)(https?:)?\/\/(www\.)?best33\.com\//g, `${cdnUrl}/`)
    .replace(/(^|\b)(https?:)?\/\/best33\.b0\.upaiyun\.com\//g, `${cdnUrl}/`)
    .replace(/(^|\b)(https?:)?\/\/(\w+)\.gravatar\.com\/avatar\//g, gravatarUrl)
}
