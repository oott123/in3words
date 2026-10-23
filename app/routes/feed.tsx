import { feed } from '~/data/base'

export async function loader() {
  return feed()
}
