import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getPosts, SummarizedPost } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { indexPath } from '~/path'
import ListPage from '~/components/ListPage'

type BlogListPageData = {
  page: number
  posts: SummarizedPost[]
  total: number
  totalPages: number
}

export const loader: LoaderFunction<BlogListPageData> = async () => {
  const { posts, total, totalPages } = await getPosts()
  return { posts, total, totalPages, page: 1 }
}

export const meta: BlogMetaFunction = ({ parentsData: { root } }) => {
  return {
    title: blogTitle('', root),
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
