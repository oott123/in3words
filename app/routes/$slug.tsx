import type { Route } from './+types/$slug'
import BlogPage from '~/components/BlogPage'
import BlogPost from '~/components/BlogPost'
import { getPage } from '~/data/posts'
import { pagePath } from '~/path'
import { postMeta } from '~/utils/meta'
import { getComments } from '~/data/comments'
import BlogComment from '~/components/BlogComment'

export async function loader({ params, url }: Route.LoaderArgs) {
  if (!params.slug) {
    throw new Error('页面 ID 不合法')
  }
  const post = await getPage(params.slug)
  const commentPage = Number(url.searchParams.get('comments_page') || '1')
  const comments = await getComments(post.id, commentPage)
  return { post, comments, commentPage, slug: params.slug }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  return postMeta(matches, loaderData?.post)
}

export default function Page({ loaderData }: Route.ComponentProps) {
  const { post, commentPage, comments, slug } = loaderData
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
