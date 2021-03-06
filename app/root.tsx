import {
  json,
  Links,
  LiveReload,
  LoaderFunction,
  Meta,
  MetaFunction,
  Outlet,
  Scripts,
  ScrollRestoration,
  useCatch,
  useLoaderData,
} from 'remix'
import type { LinksFunction } from 'remix'
import globalStylesUrl from '~/styles/index.css'
import BlogTitle from './components/BlogTitle'
import SideLinks, { SideLinkItem } from './components/SideLinks'
import SideHeader from './components/SideHeader'
import CreativeCommons from './components/CreativeCommons'
import { BlogIconsDef } from './components/BlogIcon'
import { getCategories, getSite, getTagCloud } from './data/site'
import TagCloud from './components/TagCloud'
import BlogSidebar from './components/BlogSidebar'
import { categoryPath, pagePath } from './path'
import BlogFooter from './components/BlogFooter'
import BlogCard from './components/BlogCard'
import BlogPage from './components/BlogPage'

const loadData = async () => {
  const [site, categories, tagCloud] = await Promise.all([
    getSite(),
    getCategories(),
    getTagCloud(),
  ] as const)
  return { site, categories, tagCloud }
}

export type RootData = Awaited<ReturnType<typeof loadData>>

export const loader: LoaderFunction = async () => {
  return json(await loadData())
}

export const meta: MetaFunction = ({ data }: { data: RootData }) => {
  if (data) {
    return {
      title: data.site.name,
      'og:type': 'website',
      'og:locale': 'zh_CN',
      'og:site_name': data.site.name,
      'og:title': data.site.name,
      'og:description': data.site.description,
    } as Record<string, string>
  } else {
    return {
      title: 'Error',
    }
  }
}

// https://remix.run/api/app#links
export const links: LinksFunction = () => {
  return [
    { rel: 'stylesheet', href: globalStylesUrl },
    {
      rel: 'alternate',
      type: 'application/rss+xml',
      href: `/feed/`,
    },
  ]
}

// https://remix.run/api/conventions#default-export
// https://remix.run/api/conventions#route-filenames
export default function App() {
  return (
    <Document>
      <Layout>
        <Outlet />
      </Layout>
    </Document>
  )
}

// https://remix.run/docs/en/v1/api/conventions#errorboundary
export function ErrorBoundary({ error }: { error: Error }) {
  console.error(error)
  return (
    <Document title="Error!">
      <Layout>
        <BlogCard>
          <h2>???????????????</h2>
          <p>{error.message}</p>
          <p>
            ???????????????<a href="/">??????</a>????????????
          </p>
        </BlogCard>
      </Layout>
    </Document>
  )
}

// https://remix.run/docs/en/v1/api/conventions#catchboundary
export function CatchBoundary() {
  const caught = useCatch()

  let message
  switch (caught.status) {
    case 404:
      message = (
        <div>
          <p>
            ??????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
          </p>
          <p>
            ??????????????????????????????????????????????????????????????????
            <a href="/about">?????????</a>
            ????????????????????????????????????????????????????????????
          </p>
          <p>
            ????????????????????????<a href="/">??????</a>?????????????????????????????????
          </p>
        </div>
      )
      break

    case 400:
      message = (
        <div>
          <p>????????????????????????????????????????????????</p>
        </div>
      )
      break

    case 500:
      message = (
        <div>
          <p>?????????????????????????????????????????????????????????????????????</p>
        </div>
      )
      break

    default:
      message = (
        <div>
          <p>????????????????????????????????????????????????????????????</p>
        </div>
      )
  }

  return (
    <Document title={`${caught.status} ${caught.statusText}`}>
      <Layout>
        <BlogPage>
          <BlogCard>
            <h1>
              {caught.status} {caught.statusText}
            </h1>
            {message}
            <div>
              {caught.data?.code}: {caught.data?.message}
            </div>
          </BlogCard>
        </BlogPage>
      </Layout>
    </Document>
  )
}

function Document({
  children,
  title,
}: {
  children: React.ReactNode
  title?: string
}) {
  return (
    <html lang="zh-Hans">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        {title ? <title>{title}</title> : <Meta />}
        <meta name="theme-color" content="#fef2f2" />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <BlogIconsDef />
        <Scripts />
        {process.env.NODE_ENV === 'development' && <LiveReload />}
      </body>
    </html>
  )
}

function Layout({ children }: { children: React.ReactNode }) {
  const data = useLoaderData<RootData>()
  const { site, categories, tagCloud } = data || {}

  return (
    <div className="BlogLayout">
      <main className="BlogLayout_Main BlogMain">
        {children}
        <BlogFooter />
      </main>
      <aside className="BlogLayout_Sidebar">
        {data && (
          <BlogSidebar>
            <BlogTitle title={site.name} headline={site.description} />
            <SideLinks>
              <SideLinkItem to="/" title="??????" />
              <SideLinkItem to={pagePath({ slug: 'guestbook' })} title="??????" />
              <SideLinkItem to={pagePath({ slug: 'links' })} title="??????" />
              <SideLinkItem to={pagePath({ slug: 'about' })} title="??????" />
              <li className="SideLinks_Item BlogLayout_DetailsTrigger">
                <label htmlFor="BlogLayout_DetailsCheckbox">??????</label>
              </li>
            </SideLinks>
            <input type="checkbox" id="BlogLayout_DetailsCheckbox"></input>
            <div className="BlogLayout_SidebarDetails">
              <SideHeader>??????</SideHeader>
              <SideLinks>
                {categories
                  .filter((c) => c.count > 0)
                  .map((c) => (
                    <SideLinkItem
                      key={c.id}
                      to={categoryPath(c)}
                      title={`${c.name} (${c.count})`}
                    />
                  ))}
              </SideLinks>
              <SideHeader>??????</SideHeader>
              <TagCloud tags={tagCloud} />
              <SideHeader>????????????</SideHeader>
              <CreativeCommons />
            </div>
          </BlogSidebar>
        )}
      </aside>
      <aside className="BlogLayout_Ainou"></aside>
    </div>
  )
}
