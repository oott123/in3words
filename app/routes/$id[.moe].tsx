import type { MetaFunction, LoaderFunction } from '~/types'
import { LinksFunction, useLoaderData } from 'remix'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPost, Post } from '~/data/posts'
import { postPath } from '~/path'
import { blogTitle } from '~/utils/meta'
import { getComments } from '~/data/comments'
import BlogComment from '~/components/BlogComment'
import BlogCard from '~/components/BlogCard'

type PostData = {
  post: Post
  comments: Awaited<ReturnType<typeof getComments>>
  commentPage: number
}

export const loader: LoaderFunction<PostData> = async ({ params, request }) => {
  const commentPage = Number(
    new URL(request.url).searchParams.get('comments_page') || '1',
  )
  const postId = parseInt(params.id!)
  const [post, comments] = await Promise.all([
    getPost(postId),
    getComments(postId, commentPage),
  ] as const)
  return { post, comments, commentPage }
}

export const meta: MetaFunction<PostData> = ({
  parentsData: { root },
  data,
}) => {
  return {
    title: blogTitle(data?.post?.title || '', root),
    ...(data?.post
      ? {
          'og:url': data.post.canonicalUrl,
          'og:type': 'article',
          'og:article:published_time': new Date(
            data.post.createdAt,
          ).toISOString(),
          'og:article:modified_time': new Date(
            data.post.updatedAt,
          ).toISOString(),
          'og:article:section':
            data.post.categories && data.post.categories[0].name,
          'og:updated_time': new Date(data.post.updatedAt).toISOString(),
          'og:title': data.post.title,
          'og:description': data.post.summaryText,
        }
      : {}),
  }
}

export default function Post() {
  const { post, comments, commentPage } = useLoaderData<PostData>()
  const path = postPath(post)

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
