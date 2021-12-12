import type { RootData } from '~/root'

export function blogTitle(title: string | string[], rootData: RootData) {
  if (!title) {
    return rootData.site.name
  } else {
    return [...(Array.isArray(title) ? title : [title]), rootData.site.name]
      .filter((x) => !!x)
      .join(' - ')
  }
}
