if (!globalThis.blogApiGetCache) {
  globalThis.blogApiGetCache = new Map()
}

const BASE_URL = process.env.BASE_URL || 'https://best33.com/wp-json'

export async function get<T = any>(
  path: string,
  query?: Record<string, any>,
  prefix = '/wp/v2',
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
  prefix = '/wp/v2',
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
    throw new CmsError(data.message, data.code, resp.status, resp.url)
  }
  if (!resp.ok) {
    throw new CmsError('Unknown error', 'unknown', resp.status, resp.url)
  }

  return data
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
