import { useLoaderData } from 'remix'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { SummarizedPost } from '~/data/posts'
import { indexPath } from '~/path'

export type BlogListPageData = {
  page: number
  posts: SummarizedPost[]
  total: number
  totalPages: number
}

export default function BlogListPage() {
  const { posts, totalPages, page } = useLoaderData<BlogListPageData>()

  return (
    <div className="BlogIndex">
      <PostList posts={posts} />
      <PageNavigation page={page} totalPages={totalPages} path={indexPath} />
    </div>
  )
}
