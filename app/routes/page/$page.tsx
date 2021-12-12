import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import BlogListPage, { BlogListPageData } from '~/pages/BlogListPage'

type PageParams = {
  page: string
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

export const meta: BlogMetaFunction<keyof PageParams> = ({
  parentsData: { root },
  params: { page },
}) => {
  return {
    title: blogTitle(`第 ${page!} 页`, root),
  }
}

export default BlogListPage
