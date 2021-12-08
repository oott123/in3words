import type { MetaFunction, LoaderFunction } from 'remix'
import { useLoaderData, json } from 'remix'
import PostList from '~/components/PostList'
import { getPosts, Post } from '~/data/posts'

type IndexData = {
  posts: Post[]
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
  const { posts } = useLoaderData<IndexData>()

  return (
    <div className="BlogIndex">
      <PostList posts={posts} />
    </div>
  )
}
