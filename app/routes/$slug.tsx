import type { MetaFunction, LoaderFunction } from '~/types'
import { useLoaderData } from 'remix'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPage, Post } from '~/data/posts'
import { pagePath, postPath } from '~/path'
import { blogTitle } from '~/utils/meta'
import { getComments } from '~/data/comments'
import BlogComment from '~/components/BlogComment'
import { meta as postMeta } from './$id[.moe]'

type PostData = {
  post: Post
  comments: Awaited<ReturnType<typeof getComments>>
  commentPage: number
  slug: string
}

type PageParam = {
  slug: string
}

export const loader: LoaderFunction<PostData, keyof PageParam> = async ({
  params,
  request,
}) => {
  if (!params.slug) {
    throw new Error('页面 ID 不合法')
  }
  const post = await getPage(params.slug)
  const commentPage = Number(
    new URL(request.url).searchParams.get('comments_page') || '1',
  )
  const comments = await getComments(post.id, commentPage)
  return { post, comments, commentPage, slug: params.slug }
}

export const meta: MetaFunction<PostData> = (arg) => {
  return postMeta(arg)
}

export default function Page() {
  const { post, commentPage, comments, slug } = useLoaderData<PostData>()
  const path = pagePath({ slug })

  return (
    <BlogPage>
      <BlogPost post={post} postPath={path} />
      <BlogComment
        {...comments}
        page={commentPage}
        postId={post.id}
        newCommentsAllowed={post.newCommentsAllowed}
      />
    </BlogPage>
  )
}
