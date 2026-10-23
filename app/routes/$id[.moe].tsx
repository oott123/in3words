import type { Route } from './+types/$id[.moe]'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPost } from '~/data/posts'
import { postPath } from '~/path'
import { postMeta } from '~/utils/meta'
import { getComments } from '~/data/comments'
import BlogComment from '~/components/BlogComment'

export async function loader({ params, url }: Route.LoaderArgs) {
  const commentPage = Number(url.searchParams.get('comments_page') || '1')
  const postId = parseInt(params.id)
  const [post, comments] = await Promise.all([
    getPost(postId),
    getComments(postId, commentPage),
  ] as const)
  return { post, comments, commentPage }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  return postMeta(matches, loaderData?.post)
}

export default function Post({ loaderData }: Route.ComponentProps) {
  const { post, comments, commentPage } = loaderData
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
