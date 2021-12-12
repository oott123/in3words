import type { MetaFunction, LoaderFunction } from 'remix'
import { useLoaderData, json } from 'remix'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import { getPosts, SummarizedPost } from '~/data/posts'
import { indexPath } from '~/path'

type IndexData = {
  posts: SummarizedPost[]
  total: number
  totalPages: number
}

// https://remix.run/api/conventions#loader
export const loader: LoaderFunction = async () => {
  const { posts, total, totalPages } = await getPosts()
  return json({ posts, total, totalPages })
}

// https://remix.run/api/conventions#meta
export const meta: MetaFunction = () => {
  return {
    title: 'Remix Starter',
    description: 'Welcome to remix!',
  }
}

// https://remix.run/guides/routing#index-routes
export default function Index() {
  const { posts, totalPages } = useLoaderData<IndexData>()

  return (
    <div className="BlogIndex">
      <PostList posts={posts} />
      <PageNavigation page={1} totalPages={totalPages} path={indexPath} />
    </div>
  )
}
