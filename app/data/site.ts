import { createErrorResponse, get } from './base'

export interface SiteMeta {
  name: string
  description: string
  url: string
}

export async function getSite(): Promise<SiteMeta> {
  // return {
  //   name: '三言三语',
  //   description: 'best33.com | 希望和你交朋友！',
  //   url: 'https://best33.com',
  // }
  const data = await get(
    '/',
    {
      _fields: ['name', 'description', 'url'],
    },
    '',
  )
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
  const data = await get('/categories')
  return data.map(({ id, name, slug, count }: CategoryMeta) => ({
    id,
    name,
    slug,
    count,
  }))
}

export async function getCategory(slugInput: string) {
  const resp = await get('/categories', { slug: slugInput })
  if (!resp[0]) {
    throw createErrorResponse('请求的分类不存在', 'unknown_category', 404)
  }
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
  const data = await get('/tags', {
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
  const resp = await get('/tags', { slug: slugInput })
  if (!resp[0]) {
    throw createErrorResponse('请求的标签不存在', 'unknown_tag', 404)
  }
  const { id, name, count, slug } = resp[0]

  return { id, name, slug, count } as TagMeta
}

export type UserMeta = {
  name: string
  slug: string
  id: number
}

export async function getUser(slugInput: string) {
  const resp = await get('/users', { slug: slugInput })
  if (!resp[0]) {
    throw createErrorResponse('请求的用户不存在', 'unknown_user', 404)
  }
  const { name, slug, id } = resp[0]

  return { name, slug, id } as UserMeta
}
