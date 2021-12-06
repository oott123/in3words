import { get } from './base'

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

  return posts.map(
    (post: any) =>
      ({
        id: post.id,
        createdAt: parseGmt(post.date_gmt),
        updatedAt: parseGmt(post.modified_gmt),
        title: post.title.rendered,
        content: post.content.rendered,
        author: {
          name: findEmbedded(post._embedded, 'author', post.author).name,
          avatar: findEmbedded(post._embedded, 'author', post.author)
            .avatar_urls['96'],
        } as Author,
        categories: post.categories.map(
          mapTermsByTaxonomy(post._embedded['wp:term'], 'category'),
        ),
        tags: post.tags.map(
          mapTermsByTaxonomy(post._embedded['wp:term'], 'post_tag'),
        ),
      } as Post),
  )
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
