import type { RootData } from '~/root'

export function blogTitle(title: string | string[], rootData: RootData) {
  if (!title) {
    return rootData.site.name
  } else {
    return `${
      Array.isArray(title) ? title.filter((x) => !!x).join(' - ') : title
    } - ${rootData.site.name}`
  }
}
