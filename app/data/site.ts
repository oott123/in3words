import { get } from './base'

export interface SiteMeta {
  name: string
  description: string
  url: string
}

export async function getSite() {
  const data = await get('/')
  return {
    name: data.name,
    description: data.description,
    url: data.url,
  } as SiteMeta
}

export interface CategoryMeta {
  id: number
  name: string
  slug: string
  count: number
}

export async function getCategories(): Promise<CategoryMeta[]> {
  const data = await get('/wp/v2/categories')
  return data.map(({ id, name, slug, count }: CategoryMeta) => ({
    id,
    name,
    slug,
    count,
  }))
}
