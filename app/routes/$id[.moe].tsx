import type { MetaFunction, LoaderFunction } from '~/types'
import { useLoaderData } from 'remix'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPost, Post } from '~/data/posts'
import { postPath } from '~/path'
import { blogTitle } from '~/utils/meta'

type PostData = {
  post: Post
}

export const loader: LoaderFunction<PostData> = async ({ params }) => {
  const post = await getPost(parseInt(params.id!))
  return { post }
}

export const meta: MetaFunction<PostData> = ({
  parentsData: { root },
  data: { post },
}) => {
  return {
    title: blogTitle(post.title, root),
  }
}

export default function Index() {
  const { post } = useLoaderData<PostData>()
  const path = postPath(post)

  return (
    <BlogPage>
      <BlogPost post={post} postPath={path} />
    </BlogPage>
  )
}
