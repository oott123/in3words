import classNames from 'classnames'
import { Parser } from 'htmlparser2'
import {
  encodeHtmlAttr,
  encodeHtmlText,
  getList,
  parseGmt,
  voidTags,
} from './base'

export type Comment = {
  id: number
  post: number
  parent: number
  author: number
  authorAvatar: string
  authorUrl: string
  authorName: string
  date: string
  content: string
}

export async function getComments(
  postId: number,
  page = 1,
): Promise<{ comments: Comment[]; total: number; totalPages: number }> {
  const {
    items: commentsData,
    totalPages,
    total,
  } = await getList('/comments', {
    post: postId,
    page: page,
    order: 'asc',
  })
  const comments = commentsData.map(
    (comment: any) =>
      ({
        id: comment.id,
        post: comment.post,
        parent: comment.parent,
        author: comment.author,
        authorAvatar: comment.author_avatar_urls['96'],
        authorUrl: comment.author_url,
        authorName: comment.author_name,
        date: parseGmt(comment.date_gmt),
        content: postProcessContent(comment.content.rendered).content,
      } as Comment),
  )

  return { comments, total, totalPages }
}

function postProcessContent(html: string) {
  const content: string[] = []

  const parser = new Parser({
    onopentag(name, attribs) {
      if (name === 'a') {
        attribs.target = '_blank'
        attribs.rel = 'noopener noreferrer nofollow'
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
      content.push(encodeHtmlText(text))
    },
    onclosetag(name) {
      if (!voidTags.includes(name)) {
        content.push(`</${name}>`)
      }
    },
  })

  parser.write(html)
  parser.end()

  return {
    content: content.join(''),
  }
}
