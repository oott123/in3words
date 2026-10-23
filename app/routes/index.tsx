import type { Route } from './+types/index'
import { getPosts } from '~/data/posts'
import { blogTitle, getRootData, mergeRootMeta } from '~/utils/meta'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { indexPath } from '~/path'
import ListPage from '~/components/ListPage'
import { useCallback } from 'react'

export async function loader({ url, params }: Route.LoaderArgs) {
  const keyword = url.searchParams.get('s')
  const page = isNaN(Number(params.page)) ? 1 : Number(params.page)

  if (!keyword) {
    const { posts, total, totalPages } = await getPosts(page)
    return { posts, total, totalPages, page }
  } else {
    const { posts, total, totalPages } = await getPosts(page, {
      search: keyword,
    })
    return { posts, total, totalPages, page, keyword }
  }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  const keyword = loaderData?.keyword
  const page = loaderData?.page ?? 1
  return mergeRootMeta(matches, {
    title: blogTitle(
      [keyword ? `搜索：${keyword}` : '', page > 1 ? `第 ${page} 页` : ''],
      getRootData(matches),
    ),
    ...(keyword ? { robots: 'noindex, nofollow' } : {}),
  })
}

export default function BlogListPage({ loaderData }: Route.ComponentProps) {
  const { posts, totalPages, page, keyword } = loaderData
  const path = useCallback(
    (page: number) => {
      return indexPath(page, keyword)
    },
    [keyword],
  )

  return (
    <ListPage>
      {keyword && (
        <PageNavigation page={page} totalPages={totalPages} path={path}>
          {keyword ? <div>搜索：{keyword}</div> : null}
        </PageNavigation>
      )}
      <PostList posts={posts} />
      <PageNavigation page={page} totalPages={totalPages} path={path} />
    </ListPage>
  )
}
