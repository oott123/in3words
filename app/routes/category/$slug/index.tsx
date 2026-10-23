import type { Route } from './+types/index'
import { getCategoryPosts } from '~/data/posts'
import { blogTitle, getRootData, mergeRootMeta } from '~/utils/meta'
import { categoryPath } from '~/path'
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

  const { posts, total, totalPages, category } = await getCategoryPosts(
    params.slug,
    page,
  )
  return { posts, total, totalPages, page, category }
}

export function meta({ matches, loaderData }: Route.MetaArgs) {
  if (!loaderData) {
    return mergeRootMeta(matches, {})
  }
  const {
    category: { name },
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
  const { posts, totalPages, page, category } = loaderData

  return (
    <ListPage>
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => categoryPath(category, p)}
      >
        <div>
          {category.name} ({category.count})
        </div>
      </PageNavigation>
      <PostList posts={posts} />
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => categoryPath(category, p)}
      />
    </ListPage>
  )
}
