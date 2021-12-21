import { feed, forwardRequest } from '~/data/base'
import { LoaderFunction } from '~/types'

export const loader: LoaderFunction = async ({ request }) => {
  return feed()
}
