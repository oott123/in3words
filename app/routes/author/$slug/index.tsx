import type { Route } from './+types/index'
import { getAuthorPosts } from '~/data/posts'
import { blogTitle, getRootData, mergeRootMeta } from '~/utils/meta'
import { authorPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'

export async function loader({ params }: Route.LoaderArgs) {
  let page = Number(params.page)
  if (isNaN(page)) {
    page = 1
  }
  if (!params.slug) {
    throw new Error('非法作者')
  }

  const { posts, total, totalPages, author } = await getAuthorPosts(
    params.slug,
    page,
  )
  return { posts, total, totalPages, page, author }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  if (!loaderData) {
    return mergeRootMeta(matches, {})
  }
  const {
    author: { name },
    page,
  } = loaderData
  return mergeRootMeta(matches, {
    title: blogTitle(
      (page > 1 ? `第 ${page} 页 - ` : '') + `作者：${name}`,
      getRootData(matches),
    ),
  })
}

export default function BlogListPage({ loaderData }: Route.ComponentProps) {
  const { posts, totalPages, page, author } = loaderData

  return (
    <ListPage>
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => authorPath(author, p)}
      >
        <div>作者：{author.name}</div>
      </PageNavigation>
      <PostList posts={posts} />
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => authorPath(author, p)}
      />
    </ListPage>
  )
}
