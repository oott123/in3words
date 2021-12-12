import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import BlogListPage, { BlogListPageData } from '~/pages/BlogListPage'

export const loader: LoaderFunction<BlogListPageData> = async () => {
  const { posts, total, totalPages } = await getPosts()
  return { posts, total, totalPages, page: 1 }
}

export const meta: BlogMetaFunction = ({ parentsData: { root } }) => {
  return {
    title: blogTitle('', root),
  }
}

export default BlogListPage
