import { Parser } from 'htmlparser2'
import {
  createErrorResponse,
  encodeHtmlAttr,
  encodeHtmlText,
  get,
  getList,
  parseGmt,
  replaceMediaUrl,
  voidTags,
} from './base'
import hljs from 'highlight.js'
import classNames from 'classnames'
import { getCategory, getTag, getUser } from './site'
import { postPath } from '~/path'

export interface Author {
  name: string
  avatar: string
}

export interface Term<T extends string> {
  name: string
  slug: string
  taxonomy: T
}

export interface Post {
  id: number
  canonicalUrl: string
  createdAt: string
  updatedAt: string
  title: string
  content: string
  author: {
    name: string
    slug: string
    avatar: string
  }
  newCommentsAllowed: boolean
  summaryText: string
  categories?: Term<'category'>[]
  tags?: Term<'post_tag'>[]
}

export type SummarizedPost = Omit<Post, 'content'> & { summary: string }

export async function getPosts(
  page = 1,
  queryArgs: Record<string, any> = {},
): Promise<{ posts: SummarizedPost[]; total: number; totalPages: number }> {
  const { items, total, totalPages } = await getList(`/posts`, {
    _embed: 'author,wp:term',
    page,
    ...queryArgs,
  })

  const posts = items.map((post: any) => {
    const transformed = transformPost(post)
    const { summary, summaryText } = postProcessContent(
      transformed.content,
      true,
    )
    return {
      ...transformed,
      summary,
      summaryText,
      content: undefined,
    } as SummarizedPost
  })

  return { posts, total, totalPages }
}

export async function getCategoryPosts(slug: string, page = 1) {
  const category = await getCategory(slug)
  return {
    ...(await getPosts(page, { categories: category.id })),
    category,
  }
}

export async function getTagPosts(slug: string, page = 1) {
  const tag = await getTag(slug)
  return {
    ...(await getPosts(page, { tags: tag.id })),
    tag,
  }
}

export async function getAuthorPosts(slug: string, page = 1) {
  const author = await getUser(slug)
  return {
    ...(await getPosts(page, { author: author.id })),
    author,
  }
}

export async function getPost(id: number): Promise<Post> {
  const post = await get(
    `/posts/${id}`,
    {
      _embed: 'author,wp:term',
    },
    { cacheGroup: `post:${id}` },
  )

  const transformed = transformPost(post)
  const { content, summaryText } = postProcessContent(transformed.content)
  transformed.content = content
  transformed.summaryText = summaryText

  return transformed
}

export async function getPage(slug: string): Promise<Post> {
  const post = await get(
    `/pages`,
    {
      _embed: 'author,wp:term',
      slug,
    },
    { cacheGroup: `page:${slug}` },
  )

  if (!post[0]) {
    throw createErrorResponse('请求的页面不存在', 'unknown_page', 404)
  }

  const transformed = transformPost(post[0])
  const { content, summaryText } = postProcessContent(transformed.content)
  transformed.content = content
  transformed.summaryText = summaryText

  return transformed
}

function transformPost(post: any): Post {
  return {
    id: post.id,
    canonicalUrl: `${process.env.PUBLIC_URL || ''}${
      post ? postPath(post) : '/'
    }`,
    createdAt: parseGmt(post.date_gmt),
    updatedAt: parseGmt(post.modified_gmt),
    title: post.title.rendered,
    content: post.content.rendered,
    author: {
      name: findEmbedded(post._embedded, 'author', post.author).name,
      slug: findEmbedded(post._embedded, 'author', post.author).slug,
      avatar: findEmbedded(post._embedded, 'author', post.author).avatar_urls[
        '96'
      ],
    } as Author,
    categories: post.categories?.map(
      mapTermsByTaxonomy(post._embedded['wp:term'], 'category'),
    ),
    tags: post.tags?.map(
      mapTermsByTaxonomy(post._embedded['wp:term'], 'post_tag'),
    ),
    newCommentsAllowed: post.comment_status === 'open',
  } as Post
}

function findEmbedded(embedded: any, resource: string, id: number) {
  const result = embedded[resource].find((item: any) => item.id === id)
  if (!result) {
    throw new Error(`Could not find ${resource} ${id}`)
  }
  return result
}

function mapTermsByTaxonomy(terms: any, taxonomy: string) {
  return function (id: number) {
    const newTerms = terms
      .flat()
      .filter((term: any) => term.taxonomy === taxonomy)
    const result: Record<number, Term<any>> = {}
    for (const term of newTerms) {
      result[term.id] = {
        name: term.name,
        slug: term.slug,
        taxonomy: term.taxonomy,
      }
    }
    return result[id]
  }
}

function postProcessContent(html: string, summaryOnly?: boolean) {
  const content: string[] = []
  const contentText: string[] = []
  let summary = ''
  let currentTag = ''
  let currentLanguage = ''
  let paragraphs = 0

  const parser = new Parser({
    onopentag(name, attribs) {
      if (currentTag === 'pre' && name === 'code') {
        currentLanguage =
          `${attribs.class}`.match(/language-(\w+)/)?.[1]?.toLowerCase() ||
          'AUTO'
      }
      currentTag = name

      if (name === 'a' && attribs.href && attribs.href.startsWith('http')) {
        attribs.target = attribs.target || '_blank'
        attribs.rel = attribs.rel || 'noopener noreferrer'
        attribs.class = classNames(attribs.class, 'BlogLink-External')
        attribs.href = replaceMediaUrl(attribs.href)
      }

      if (name === 'img' && typeof attribs.src === 'string') {
        attribs.src = replaceMediaUrl(attribs.src)
        if (attribs.srcset) {
          attribs.srcset = replaceMediaUrl(attribs.srcset)
        }
        attribs.loading = attribs.loading || 'lazy'
      }

      // SyntaxHighlighter 兼容
      if (name === 'pre' && attribs.class?.includes('brush:')) {
        const language = attribs.class
          .match(/brush:\s*(\w+)/)?.[1]
          ?.toLowerCase()
        if (language) {
          currentLanguage = language
        }
      }

      content.push(`<${name}`)
      for (const key in attribs) {
        content.push(
          ` ${encodeHtmlAttr(key)}="${encodeHtmlAttr(attribs[key])}"`,
        )
      }
      content.push('>')
    },
    ontext(text) {
      if (currentLanguage) {
        if (currentLanguage === 'AUTO') {
          content.push(hljs.highlightAuto(text).value)
        } else {
          try {
            content.push(
              hljs.highlight(text, { language: currentLanguage }).value,
            )
          } catch {
            content.push(hljs.highlightAuto(text).value)
          }
        }
      } else {
        content.push(encodeHtmlText(text))
      }
      if (currentTag !== 'script' && currentTag !== 'style') {
        contentText.push(text)
      }
    },
    onclosetag(name) {
      currentTag = ''
      currentLanguage = ''
      if (!voidTags.includes(name)) {
        content.push(`</${name}>`)
      }

      if (name === 'p') {
        paragraphs++
      }
      if (paragraphs >= 2) {
        summary = content.join('')
        if (summaryOnly) {
          parser.pause()
        }
      }
    },
  })

  parser.write(html)
  parser.end()

  if (!summary) {
    summary = content.join('')
  }

  return {
    content: content.join(''),
    summary: summary,
    summaryText: contentText.join('').substring(0, 200),
  }
}
