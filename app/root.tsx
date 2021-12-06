import {
  Link,
  Links,
  LiveReload,
  LoaderFunction,
  Meta,
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

export const loader: LoaderFunction = async () => {
  return 1
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
        {title ? <title>{title}</title> : null}
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
        {process.env.NODE_ENV === 'development' && <LiveReload />}
      </body>
    </html>
  )
}

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="BlogLayout">
      <aside className="BlogLayout_Sidebar BlogSidebar">
        <BlogTitle title="三言三语" headline="best33.com | 希望和你做朋友！" />
        <SideLinks>
          <SideLinkItem to="/" title="文章" active />
          <SideLinkItem to="/" title="留言" />
          <SideLinkItem to="/" title="链接" />
          <SideLinkItem to="/" title="关于" />
        </SideLinks>
        <SideHeader>分类</SideHeader>
        <SideLinks>
          <SideLinkItem to="/" title="Linux 笔记 (18)" />
          <SideLinkItem to="/" title="M$ 大法 (3)" />
          <SideLinkItem to="/" title="PHP 手札 (23)" />
          <SideLinkItem to="/" title="Web 万象 (30)" />
          <SideLinkItem to="/" title="仓鼠症 (1)" />
          <SideLinkItem to="/" title="博客记录 (13)" />
          <SideLinkItem to="/" title="原理拾穗 (4)" />
          <SideLinkItem to="/" title="瞎想瞎扯 (1)" />
          <SideLinkItem to="/" title="辣鸡水果 (1)" />
        </SideLinks>
        <SideHeader>标签</SideHeader>
        <SideHeader>授权协议</SideHeader>
        <CreativeCommons />
      </aside>
      <main className="BlogLayout_Main BlogMain">{children}</main>
      <aside className="BlogLayout_Ainou"></aside>
    </div>
  )
}
