export async function get<T = any>(
  path: string,
  query?: Record<string, any>,
): Promise<T> {
  const baseUrl = 'https://best33.com/wp-json'

  const queryString = query
    ? `?${Object.entries(query)
        .map((x) => x.map(encodeURIComponent).join('='))
        .join('&')}`
    : ''
  const response = await fetch(`${baseUrl}${path}${queryString}`)
  const data = await response.json()
  if (!response.ok && data) {
    throw new CmsError(data.message, data.code, response.status)
  }
  if (!response.ok) {
    throw new CmsError('Unknown error', 'unknown', response.status)
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
