import {
  json,
  Link,
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
import { useEffect } from 'react'
import BlogTitle from './components/BlogTitle'
import SideLinks, { SideLinkItem } from './components/SideLinks'
import SideHeader from './components/SideHeader'
import CreativeCommons from './components/CreativeCommons'
import { BlogIconsDef } from './components/BlogIcon'
import { getCategories, getSite, getTagCloud } from './data/site'
import TagCloud from './components/TagCloud'
import BlogSidebar from './components/BlogSidebar'
import { categoryPath, pagePath } from './path'

const loadData = async () => {
  const [site, categories, tagCloud] = await Promise.all([
    getSite(),
    getCategories(),
    getTagCloud(),
  ] as const)
  return { site, categories, tagCloud }
}

type RootData = Awaited<ReturnType<typeof loadData>>

export const loader: LoaderFunction = async () => {
  return json(await loadData())
}

export const meta: MetaFunction = ({ data }: { data: RootData }) => {
  if (data) {
    return {
      title: data.site.name,
    }
  } else {
    return {
      title: 'Error',
    }
  }
}

// https://remix.run/api/app#links
export const links: LinksFunction = () => {
  return [{ rel: 'stylesheet', href: globalStylesUrl }]
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
        <div>
          <h1>There was an error</h1>
          <p>{error.message}</p>
          <hr />
          <p>
            Hey, developer, you should replace this with what you want your
            users to see.
          </p>
        </div>
      </Layout>
    </Document>
  )
}

// https://remix.run/docs/en/v1/api/conventions#catchboundary
export function CatchBoundary() {
  const caught = useCatch()

  let message
  switch (caught.status) {
    case 401:
      message = (
        <p>
          Oops! Looks like you tried to visit a page that you do not have access
          to.
        </p>
      )
      break
    case 404:
      message = (
        <p>Oops! Looks like you tried to visit a page that does not exist.</p>
      )
      break

    default:
      throw new Error(caught.data || caught.statusText)
  }

  return (
    <Document title={`${caught.status} ${caught.statusText}`}>
      <Layout>
        <h1>
          {caught.status}: {caught.statusText}
        </h1>
        {message}
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
        <Meta />
        {title ? <title>{title}</title> : null}
        <Links />
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
      <main className="BlogLayout_Main BlogMain">{children}</main>
      <aside className="BlogLayout_Sidebar">
        {data && (
          <BlogSidebar>
            <BlogTitle title={site.name} headline={site.description} />
            <SideLinks>
              <SideLinkItem to="/" title="文章" active />
              <SideLinkItem to={pagePath({ slug: 'guestbook' })} title="留言" />
              <SideLinkItem to={pagePath({ slug: 'links' })} title="链接" />
              <SideLinkItem to={pagePath({ slug: 'about' })} title="关于" />
            </SideLinks>
            <SideHeader>分类</SideHeader>
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
            <SideHeader>标签</SideHeader>
            <TagCloud tags={tagCloud} />
            <SideHeader>授权协议</SideHeader>
            <CreativeCommons />
          </BlogSidebar>
        )}
      </aside>
      <aside className="BlogLayout_Ainou"></aside>
    </div>
  )
}
