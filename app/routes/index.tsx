import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getPosts, SummarizedPost } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { indexPath } from '~/path'
import ListPage from '~/components/ListPage'
import { useCallback } from 'react'

type PageParams = {
  page: string
}

type IndexData = {
  page: number
  posts: SummarizedPost[]
  total: number
  totalPages: number
  keyword?: string
}

export const loader: LoaderFunction<IndexData, keyof PageParams> = async ({
  request,
  params,
}) => {
  const url = new URL(request.url)
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

export const meta: BlogMetaFunction<IndexData> = ({
  parentsData: { root },
  data: { keyword, page },
}) => {
  return {
    title: blogTitle(
      [keyword ? `搜索：${keyword}` : '', page > 1 ? `第 ${page} 页` : ''],
      root,
    ),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page, keyword } = useLoaderData<IndexData>()
  const path = useCallback(
    (page) => {
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
