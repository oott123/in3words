import { useEffect } from 'react'
import type { MetaFunction, LoaderFunction } from 'remix'
import { useLoaderData, json, Link } from 'remix'
import BlogPost from '~/components/BlogPost'
import { getPost, getPosts, Post } from '~/data/posts'

type PostData = {
  post: Post
}

// https://remix.run/api/conventions#loader
export const loader: LoaderFunction = async ({ params }) => {
  const post = await getPost(parseInt(params.id!))
  return json({ post: post })
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
  const { post } = useLoaderData<PostData>()

  return (
    <div className="BlogPage">
      <BlogPost post={post} />
    </div>
  )
}
