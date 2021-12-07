import { Parser } from 'htmlparser2'
import { get } from './base'
import hljs from 'highlight.js'
import classNames from 'classnames'

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
  createdAt: string
  updatedAt: string
  title: string
  content: string
  author: {
    name: string
    slug: string
    avatar: string
  }
  categories: Term<'category'>[]
  tags: Term<'post_tag'>[]
}

export type SummarizedPost = Omit<Post, 'content'> & { summary: string }

export async function getPosts(page = 1): Promise<Post[]> {
  const posts = await get(`/wp/v2/posts`, {
    _embed: 'author,wp:term',
    page,
  })

  return posts.map((post: any) => {
    const transformed = transformPost(post)
    const { summary } = postProcessContent(transformed.content, true)
    return {
      ...transformed,
      summary,
      content: undefined,
    } as SummarizedPost
  })
}

export async function getPost(id: number): Promise<Post> {
  const post = await get(`/wp/v2/posts/${id}`, {
    _embed: 'author,wp:term',
  })

  const transformed = transformPost(post)
  const { content } = postProcessContent(transformed.content)
  transformed.content = content

  return transformed
}

function transformPost(post: any): Post {
  return {
    id: post.id,
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
    categories: post.categories.map(
      mapTermsByTaxonomy(post._embedded['wp:term'], 'category'),
    ),
    tags: post.tags.map(
      mapTermsByTaxonomy(post._embedded['wp:term'], 'post_tag'),
    ),
  } as Post
}

function findEmbedded(embedded: any, resource: string, id: number) {
  const result = embedded[resource].find((item: any) => item.id === id)
  if (!result) {
    throw new Error(`Could not find ${resource} ${id}`)
  }
  return result
}

function parseGmt(date: string): string {
  return new Date(`${date}Z`).toISOString()
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

const voidTags = [
  'area',
  'base',
  'br',
  'col',
  'embed',
  'hr',
  'img',
  'input',
  'keygen',
  'link',
  'meta',
  'param',
  'source',
  'track',
  'wbr',
]

function postProcessContent(html: string, summaryOnly?: boolean) {
  const content: string[] = []
  const summary: string[] = []
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

      if (name === 'a') {
        attribs.target = attribs.target || '_blank'
        attribs.rel = attribs.rel || 'noopener noreferrer'
        attribs.class = classNames(attribs.class, 'BlogLink-External')
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
          content.push(
            hljs.highlight(text, { language: currentLanguage }).value,
          )
        }
      } else {
        content.push(encodeHtmlText(text))
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
        summary.push(content.join(''))
        if (summaryOnly) {
          parser.pause()
        }
      }
    },
  })

  parser.write(html)
  parser.end()

  return {
    content: content.join(''),
    summary: summary.join(''),
  }
}

function encodeHtmlText(str: string) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function encodeHtmlAttr(str: string) {
  return encodeHtmlText(str).replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}
