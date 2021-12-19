import { LoaderFunction } from '~/types'

export const loader: LoaderFunction = async ({ request }) => {
  const url = new URL(request.url)
  const newUrl = `https://seccdn.libravatar.org/${url.pathname}${url.search}`

  const resp = await fetch(newUrl)
  resp.headers.delete('Access-Control-Allow-Origin')

  return resp
}
