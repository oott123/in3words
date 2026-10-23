import type { Route } from './+types/index'
import { getTagPosts } from '~/data/posts'
import { blogTitle, getRootData, mergeRootMeta } from '~/utils/meta'
import { tagPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'

export async function loader({ params }: Route.LoaderArgs) {
  let page = Number(params.page)
  if (isNaN(page)) {
    page = 1
  }
  if (!params.slug) {
    throw new Error('非法分类')
  }

  const { posts, total, totalPages, tag } = await getTagPosts(params.slug, page)
  return { posts, total, totalPages, page, tag }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  if (!loaderData) {
    return mergeRootMeta(matches, {})
  }
  const {
    tag: { name },
    page,
  } = loaderData
  return mergeRootMeta(matches, {
    title: blogTitle(
      (page > 1 ? `第 ${page} 页 - ` : '') + `${name}`,
      getRootData(matches),
    ),
  })
}

export default function BlogListPage({ loaderData }: Route.ComponentProps) {
  const { posts, totalPages, page, tag } = loaderData

  return (
    <ListPage>
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => tagPath(tag, p)}
      >
        <div>
          {tag.name} ({tag.count})
        </div>
      </PageNavigation>
      <PostList posts={posts} />
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => tagPath(tag, p)}
      />
    </ListPage>
  )
}
