import { forwardRequest } from '~/data/base'
import type { Route } from './+types/$'

export async function loader({ request, url }: Route.LoaderArgs) {
  return forwardRequest(request, url)
}
