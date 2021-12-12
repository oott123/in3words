import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getCategoryPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import { categoryPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'

type Params = {
  page: string
  slug: string
}

type Data = Awaited<ReturnType<typeof getCategoryPosts>> & { page: number }

export const loader: LoaderFunction<Data, keyof Params> = async ({
  params,
}) => {
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

export const meta: BlogMetaFunction<
  Data,
  Record<string, any>,
  keyof Params
> = ({
  parentsData: { root },
  data: {
    category: { name },
    page,
  },
}) => {
  return {
    title: blogTitle((page > 1 ? `第 ${page} 页 - ` : '') + `${name}`, root),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page, category } = useLoaderData<Data>()

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
