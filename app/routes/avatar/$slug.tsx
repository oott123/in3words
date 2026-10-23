import { proxyResponse } from '~/data/base'
import type { Route } from './+types/$slug'

export async function loader({ url }: Route.LoaderArgs) {
  const newUrl = `https://seccdn.libravatar.org/${url.pathname}${url.search}`

  const resp = await fetch(newUrl)

  return proxyResponse(resp, ['Access-Control-Allow-Origin'])
}
