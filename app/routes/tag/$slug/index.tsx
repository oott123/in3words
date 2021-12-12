import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getTagPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import { tagPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'

type Params = {
  page: string
  slug: string
}

type Data = Awaited<ReturnType<typeof getTagPosts>> & { page: number }

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

  const { posts, total, totalPages, tag } = await getTagPosts(params.slug, page)
  return { posts, total, totalPages, page, tag }
}

export const meta: BlogMetaFunction<
  Data,
  Record<string, any>,
  keyof Params
> = ({
  parentsData: { root },
  data: {
    tag: { name },
    page,
  },
}) => {
  return {
    title: blogTitle((page > 1 ? `第 ${page} 页 - ` : '') + `${name}`, root),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page, tag } = useLoaderData<Data>()

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
