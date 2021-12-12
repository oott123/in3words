import type { RootData } from '~/root'

export function blogTitle(title: string, rootData: RootData) {
  if (!title) {
    return rootData.site.name
  } else {
    return `${title} - ${rootData.site.name}`
  }
}
