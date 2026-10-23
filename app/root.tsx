import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useRouteLoaderData,
} from 'react-router'
import type { Route } from './+types/root'
import globalStylesUrl from '../styles/index.scss?url'
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
import { toMetaDescriptors } from './utils/meta'

const loadData = async () => {
  const [site, categories, tagCloud] = await Promise.all([
    getSite(),
    getCategories(),
    getTagCloud(),
  ] as const)
  return { site, categories, tagCloud }
}

export type RootData = Awaited<ReturnType<typeof loadData>>

export async function loader() {
  return loadData()
}

export function meta({ loaderData }: Route.MetaArgs) {
  if (loaderData) {
    return toMetaDescriptors({
      title: loaderData.site.name,
      'og:type': 'website',
      'og:locale': 'zh_CN',
      'og:site_name': loaderData.site.name,
      'og:title': loaderData.site.name,
      'og:description': loaderData.site.description,
    })
  } else {
    return [{ title: 'Error' }]
  }
}

export const links: Route.LinksFunction = () => {
  return [
    { rel: 'stylesheet', href: globalStylesUrl },
    {
      rel: 'alternate',
      type: 'application/rss+xml',
      href: `/feed/`,
    },
  ]
}

export default function App() {
  return (
    <Document>
      <Layout>
        <Outlet />
      </Layout>
    </Document>
  )
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  if (!isRouteErrorResponse(error)) {
    console.error(error)
    return (
      <Document title="Error!">
        <Layout>
          <BlogCard>
            <h2>出错了……</h2>
            <p>{error instanceof Error ? error.message : String(error)}</p>
            <p>
              要不还是去<a href="/">首页</a>看看吧？
            </p>
          </BlogCard>
        </Layout>
      </Document>
    )
  }

  let message
  switch (error.status) {
    case 404:
      message = (
        <div>
          <p>
            对不起！我觉得很可能是由于我的疏于维护，你访问的这个链接所对应的内容已经不再存在了。
          </p>
          <p>
            不过，我应该不是故意的。如果你确实想看，可以
            <a href="/about">联系我</a>
            试试找回来。别不好意思，我会努力帮忙的！
          </p>
          <p>
            现在你还可以前往<a href="/">首页</a>看看有没有想要的内容。
          </p>
        </div>
      )
      break

    case 400:
      message = (
        <div>
          <p>你输入的内容不符合要求，请检查。</p>
        </div>
      )
      break

    case 500:
      message = (
        <div>
          <p>由于服务器内部错误，这里暂时无法显示任何内容。</p>
        </div>
      )
      break

    default:
      message = (
        <div>
          <p>由于某种原因，这里暂时无法显示任何内容。</p>
        </div>
      )
  }

  return (
    <Document title={`${error.status} ${error.statusText}`}>
      <Layout>
        <BlogPage>
          <BlogCard>
            <h1>
              {error.status} {error.statusText}
            </h1>
            {message}
            <div>
              {error.data?.code}: {error.data?.message}
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
      </body>
    </html>
  )
}

function Layout({ children }: { children: React.ReactNode }) {
  const data = useRouteLoaderData<typeof loader>('root')

  return (
    <div className="BlogLayout">
      <main className="BlogLayout_Main BlogMain">
        {children}
        <BlogFooter />
      </main>
      <aside className="BlogLayout_Sidebar">
        {data && (
          <BlogSidebar>
            <BlogTitle
              title={data.site.name}
              headline={data.site.description}
            />
            <SideLinks>
              <SideLinkItem to="/" title="文章" />
              <SideLinkItem to={pagePath({ slug: 'guestbook' })} title="留言" />
              <SideLinkItem to={pagePath({ slug: 'links' })} title="链接" />
              <SideLinkItem to={pagePath({ slug: 'about' })} title="关于" />
              <li className="SideLinks_Item BlogLayout_DetailsTrigger">
                <label htmlFor="BlogLayout_DetailsCheckbox">更多</label>
              </li>
            </SideLinks>
            <input type="checkbox" id="BlogLayout_DetailsCheckbox"></input>
            <div className="BlogLayout_SidebarDetails">
              <SideHeader>分类</SideHeader>
              <SideLinks>
                {data.categories
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
              <TagCloud tags={data.tagCloud} />
              <SideHeader>授权协议</SideHeader>
              <CreativeCommons />
            </div>
          </BlogSidebar>
        )}
      </aside>
      <aside className="BlogLayout_Ainou"></aside>
    </div>
  )
}
