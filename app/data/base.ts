if (!globalThis.blogApiGetCache) {
  globalThis.blogApiGetCache = new Map()
}

export async function get<T = any>(
  path: string,
  query?: Record<string, any>,
): Promise<T> {
  const baseUrl = 'https://best33.com/wp-json'

  const queryString = query
    ? `?${Object.entries(query)
        .map((x) => x.map(encodeURIComponent).join('='))
        .sort()
        .join('&')}`
    : ''
  const url = `${baseUrl}${path}${queryString}`

  if (globalThis.blogApiGetCache!.has(url)) {
    return globalThis.blogApiGetCache!.get(url)
  }

  const resp = await fetch(url)
  const result = await response(resp)
  globalThis.blogApiGetCache!.set(url, result)
  return result
}

async function response(resp: Response) {
  const data = await resp.json()
  if (!resp.ok && data) {
    throw new CmsError(data.message, data.code, resp.status)
  }
  if (!resp.ok) {
    throw new CmsError('Unknown error', 'unknown', resp.status)
  }

  return data
}

export class CmsError extends Error {
  public constructor(
    message: string,
    public readonly code: string,
    public readonly status?: number,
  ) {
    super(message)
    this.name = 'CmsError'
  }
}
