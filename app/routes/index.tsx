import { useEffect } from 'react'
import type { MetaFunction, LoaderFunction } from 'remix'
import { useLoaderData, json, Link } from 'remix'
import BlogPost from '~/components/BlogPost'
import { getPosts, Post } from '~/data/posts'

type IndexData = {
  posts: Post[]
}

// https://remix.run/api/conventions#loader
export const loader: LoaderFunction = async () => {
  return json({ posts: [] })
}

// https://remix.run/api/conventions#meta
export const meta: MetaFunction = () => {
  return {
    title: 'Remix Starter',
    description: 'Welcome to remix!',
  }
}

// https://remix.run/guides/routing#index-routes
export default function Index() {
  const data = useLoaderData<IndexData>()
  useEffect(() => {
    console.log(data)
  }, [data])

  return (
    <div className="BlogIndex">
      <BlogPost
        post={{
          id: 1,
          createdAt: '2005-06-07',
          updatedAt: '2005-06-07',
          title: '我如何用 Windows 开发 —— 2021 我的开发环境',
          summary: `<p>由于搭载 macOS 的设备越来越昂贵且槽点颇多，加上微软近些年开始发力开发者体验，我的工作环境已经从 MacBook 切换到了运行 Windows 系统的非苹果机器。这是多方面因素决定的：平时玩的游戏需要在 Windows 上运行，桌面环境也是 Windows 的比较舒适耐用。</p>

          <p>使用 Windows 进行开发工作虽说并不像在 macOS 上那样可以使用诸多 POSIX 标准的工具，但鉴于硬件性能的提升和虚拟化技术的成熟，使用虚拟机或者 WSL2 也未必是不可接受的方案。此外，随着 Visual Studio Code 的 Remote 功能越发完善，在 Windows 的窗口环境下享受和 Linux 一样的开发体验也并非不可能。</p>`,
          author: {
            name: '三三',
            avatar: 'about:blank',
          },
          categories: [
            { name: '分类1', slug: 'category-1', taxonomy: 'category' },
          ],
          tags: [{ name: '标签1', slug: 'tag-1', taxonomy: 'post_tag' }],
        }}
      />
      <BlogPost
        post={{
          id: 1,
          createdAt: '2005-06-07',
          updatedAt: '2005-06-07',
          title: '我如何用 Windows 开发 —— 2021 我的开发环境',
          summary: `<p>由于搭载 macOS 的设备越来越昂贵且槽点颇多，加上微软近些年开始发力开发者体验，我的工作环境已经从 MacBook 切换到了运行 Windows 系统的非苹果机器。这是多方面因素决定的：平时玩的游戏需要在 Windows 上运行，桌面环境也是 Windows 的比较舒适耐用。</p>

          <p>使用 Windows 进行开发工作虽说并不像在 macOS 上那样可以使用诸多 POSIX 标准的工具，但鉴于硬件性能的提升和虚拟化技术的成熟，使用虚拟机或者 WSL2 也未必是不可接受的方案。此外，随着 Visual Studio Code 的 Remote 功能越发完善，在 Windows 的窗口环境下享受和 Linux 一样的开发体验也并非不可能。</p>`,
          author: {
            name: '三三',
            avatar: 'about:blank',
          },
          categories: [
            { name: '分类1', slug: 'category-1', taxonomy: 'category' },
          ],
          tags: [{ name: '标签1', slug: 'tag-1', taxonomy: 'post_tag' }],
        }}
      />
      <BlogPost
        post={{
          id: 1,
          createdAt: '2005-06-07',
          updatedAt: '2005-06-07',
          title: '我如何用 Windows 开发 —— 2021 我的开发环境',
          summary: `<p>由于搭载 macOS 的设备越来越昂贵且槽点颇多，加上微软近些年开始发力开发者体验，我的工作环境已经从 MacBook 切换到了运行 Windows 系统的非苹果机器。这是多方面因素决定的：平时玩的游戏需要在 Windows 上运行，桌面环境也是 Windows 的比较舒适耐用。</p>

          <p>使用 Windows 进行开发工作虽说并不像在 macOS 上那样可以使用诸多 POSIX 标准的工具，但鉴于硬件性能的提升和虚拟化技术的成熟，使用虚拟机或者 WSL2 也未必是不可接受的方案。此外，随着 Visual Studio Code 的 Remote 功能越发完善，在 Windows 的窗口环境下享受和 Linux 一样的开发体验也并非不可能。</p>`,
          author: {
            name: '三三',
            avatar: 'about:blank',
          },
          categories: [
            { name: '分类1', slug: 'category-1', taxonomy: 'category' },
          ],
          tags: [{ name: '标签1', slug: 'tag-1', taxonomy: 'post_tag' }],
        }}
      />
    </div>
  )
}
