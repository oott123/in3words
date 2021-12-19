import type { LoaderFunction, BlogMetaFunction } from '~/types'
import { getCategoryPosts, getPosts } from '~/data/posts'
import { blogTitle } from '~/utils/meta'
import { useLoaderData } from 'remix'
import { archivePath, categoryPath } from '~/path'
import PageNavigation from '~/components/PageNavigation'
import PostList from '~/components/PostList'
import ListPage from '~/components/ListPage'
import * as yup from 'yup'
import { addMonths, addDays, addYears } from 'date-fns'

type Params = {
  page?: string
  year: string
  month?: string
  day?: string
}

type Data = Awaited<ReturnType<typeof getPosts>> & {
  page: number
  year: number
  month?: number
  day?: number
  archiveTitle: string
}

function getDateSpan({
  year,
  month,
  day,
}: {
  year: number
  month?: number | undefined
  day?: number | undefined
}) {
  const startDate = new Date(year, (month || 1) - 1, day || 1)
  let endDate: Date
  if (month == null && day == null) {
    endDate = addYears(startDate, 1)
  } else if (day == null) {
    endDate = addMonths(startDate, 1)
  } else {
    endDate = addDays(startDate, 1)
  }

  return [startDate.toISOString(), endDate.toISOString()] as const
}

export const loader: LoaderFunction<Data, keyof Params> = async ({
  params,
}) => {
  const p = await yup
    .object()
    .shape({
      page: yup.number().positive().integer().default(1),
      year: yup.number().defined().positive().integer().min(1900).max(2100),
      month: yup.number().positive().integer().min(1).max(12).notRequired(),
      day: yup.number().positive().integer().min(1).max(31).notRequired(),
    })
    .validate(params)
  const archiveTitle =
    p.day == null || p.month == null
      ? p.month == null
        ? `按年归档：${p.year} 年`
        : `按月归档：${p.year} 年 ${p.month} 月`
      : `按日归档：${p.year} 年 ${p.month} 月 ${p.day} 日`

  const [after, before] = getDateSpan(p)
  const { posts, total, totalPages } = await getPosts(p.page, { after, before })
  return { posts, total, totalPages, ...p, archiveTitle }
}

export const meta: BlogMetaFunction<
  Data,
  Record<string, any>,
  keyof Params
> = ({ parentsData: { root }, data: { page, archiveTitle } }) => {
  return {
    title: blogTitle([page > 1 && `第 ${page} 页`, archiveTitle], root),
  }
}

export default function BlogListPage() {
  const { posts, totalPages, page, archiveTitle, year, month, day } =
    useLoaderData<Data>()

  return (
    <ListPage>
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => archivePath({ year, month, day }, p)}
      >
        <div>{archiveTitle}</div>
      </PageNavigation>
      <PostList posts={posts} />
      <PageNavigation
        page={page}
        totalPages={totalPages}
        path={(p: number) => archivePath({ year, month, day }, p)}
      />
    </ListPage>
  )
}
