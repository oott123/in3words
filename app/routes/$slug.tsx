import type { MetaFunction, LoaderFunction } from '~/types'
import { useLoaderData } from 'remix'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPage, Post } from '~/data/posts'
import { postPath } from '~/path'
import { blogTitle } from '~/utils/meta'

type PostData = {
  post: Post
}

type PageParam = {
  slug: string
}

export const loader: LoaderFunction<PostData, keyof PageParam> = async ({
  params,
}) => {
  if (!params.slug) {
    throw new Error('页面 ID 不合法')
  }
  const post = await getPage(params.slug)
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

export default function Page() {
  const { post } = useLoaderData<PostData>()
  const path = postPath(post)

  return (
    <BlogPage>
      <BlogPost post={post} postPath={path} />
    </BlogPage>
  )
}
