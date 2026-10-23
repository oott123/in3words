import type { MetaDescriptor } from 'react-router'
import type { Post } from '~/data/posts'
import type { RootData } from '~/root'

type MetaMatch = { id: string; loaderData?: unknown; meta: MetaDescriptor[] }

/**
 * 键为 `title`、`og:*` 或其它 `<meta name>` 的 meta 对象，值为空时忽略。
 */
export type MetaRecord = Record<string, string | null | undefined>

export function getRootData(
  matches: ReadonlyArray<MetaMatch | undefined>,
): RootData | undefined {
  return matches.find((m) => m?.id === 'root')?.loaderData as
    RootData | undefined
}

export function blogTitle(
  title: string | Array<string | boolean | null | undefined | number>,
  rootData: RootData | undefined,
) {
  const siteName = rootData?.site.name
  if (!title) {
    return siteName ?? ''
  } else {
    return [...(Array.isArray(title) ? title : [title]), siteName]
      .filter((x) => !!x)
      .join(' - ')
  }
}

export function toMetaDescriptors(record: MetaRecord): MetaDescriptor[] {
  return Object.entries(record).flatMap(([key, content]): MetaDescriptor[] => {
    if (content == null) {
      return []
    }
    if (key === 'title') {
      return [{ title: content }]
    }
    if (key.startsWith('og:')) {
      return [{ property: key, content }]
    }
    return [{ name: key, content }]
  })
}

function metaKey(descriptor: MetaDescriptor) {
  if ('title' in descriptor) return 'title'
  if ('property' in descriptor) return `property:${descriptor.property}`
  if ('name' in descriptor) return `name:${descriptor.name}`
  return undefined
}

/**
 * 子路由的 meta 会完全替换父路由的 meta，这里把根路由的 meta 合并进来，
 * 同名的标签以子路由为准。
 */
export function mergeRootMeta(
  matches: ReadonlyArray<MetaMatch | undefined>,
  record: MetaRecord,
): MetaDescriptor[] {
  const rootMeta = matches.find((m) => m?.id === 'root')?.meta ?? []
  const ownMeta = toMetaDescriptors(record)
  const ownKeys = new Set(ownMeta.map(metaKey))
  return [...rootMeta.filter((d) => !ownKeys.has(metaKey(d))), ...ownMeta]
}

export function postMeta(
  matches: ReadonlyArray<MetaMatch | undefined>,
  post: Post | undefined,
): MetaDescriptor[] {
  return mergeRootMeta(matches, {
    title: blogTitle(post?.title || '', getRootData(matches)),
    ...(post
      ? {
          'og:url': post.canonicalUrl,
          'og:type': 'article',
          'og:article:published_time': new Date(post.createdAt).toISOString(),
          'og:article:modified_time': new Date(post.updatedAt).toISOString(),
          'og:article:section': post.categories?.[0]?.name,
          'og:updated_time': new Date(post.updatedAt).toISOString(),
          'og:title': post.title,
          'og:description': post.summaryText,
        }
      : {}),
  })
}
