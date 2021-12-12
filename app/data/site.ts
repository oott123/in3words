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

export async function getCategory(slugInput: string) {
  const resp = await get('/wp/v2/categories', { slug: slugInput })
  const { id, name, count, slug } = resp[0]

  return { id, name, slug, count } as CategoryMeta
}

export type TagMeta = {
  id: number
  name: string
  slug: string
  count: number
}

export async function getTagCloud(): Promise<TagMeta[]> {
  const data = await get('/wp/v2/tags', {
    per_page: 50,
    orderby: 'count',
    order: 'desc',
    _fields: 'id,name,slug,count',
  })
  return (data as any[])
    .map(
      ({ id, name, slug, count }: TagMeta) =>
        ({
          id,
          name,
          slug,
          count,
        } as TagMeta),
    )
    .filter((c) => c.count > 1)
    .sort((x1, x2) => x1.name.localeCompare(x2.name))
}

export async function getTag(slugInput: string) {
  const resp = await get('/wp/v2/tags', { slug: slugInput })
  const { id, name, count, slug } = resp[0]

  return { id, name, slug, count } as TagMeta
}
