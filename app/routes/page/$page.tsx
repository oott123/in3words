import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getPosts, SummarizedPost } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { indexPath } from '~/path'
import ListPage from '~/components/ListPage'

type PageParams = {
  page: string
}

type BlogListPageData = {
  page: number
  posts: SummarizedPost[]
  total: number
  totalPages: number
}

export const loader: LoaderFunction<
  BlogListPageData,
  keyof PageParams
> = async ({ params }) => {
  const page = Number(params.page)
  if (isNaN(page)) {
    throw new Error('非法页号')
  }

  const { posts, total, totalPages } = await getPosts(page)
  return { posts, total, totalPages, page }
}

export const meta: BlogMetaFunction<
  undefined,
  Record<string, any>,
  keyof PageParams
> = ({ parentsData: { root }, params: { page } }) => {
  return {
    title: blogTitle(`第 ${page!} 页`, root),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page } = useLoaderData<BlogListPageData>()

  return (
    <ListPage>
      <PostList posts={posts} />
      <PageNavigation page={page} totalPages={totalPages} path={indexPath} />
    </ListPage>
  )
}
