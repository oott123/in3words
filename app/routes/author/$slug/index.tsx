import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getAuthorPosts, getCategoryPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import { authorPath, categoryPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'

type Params = {
  page: string
  slug: string
}

type Data = Awaited<ReturnType<typeof getAuthorPosts>> & { page: number }

export const loader: LoaderFunction<Data, keyof Params> = async ({
  params,
}) => {
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

export const meta: BlogMetaFunction<
  Data,
  Record<string, any>,
  keyof Params
> = ({
  parentsData: { root },
  data: {
    author: { name },
    page,
  },
}) => {
  return {
    title: blogTitle(
      (page > 1 ? `第 ${page} 页 - ` : '') + `作者：${name}`,
      root,
    ),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page, author } = useLoaderData<Data>()

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
