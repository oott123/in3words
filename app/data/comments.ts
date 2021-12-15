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
  indent: number
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
    order: 'desc',
    per_page: 100,
  })

  const comments = commentsData
    .map(
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
          indent: 0,
        } as Comment),
    )
    .sort((a, b) => a.id - b.id)

  const unknownCommentsId: number[] = []
  const sortedCommentsId: number[] = []
  const commentMap = new Map<number, Comment>()
  for (const comment of comments) {
    commentMap.set(comment.id, comment)
    if (!comment.parent) {
      sortedCommentsId.push(comment.id)
    } else {
      const foundIndex = sortedCommentsId.indexOf(comment.parent)
      if (foundIndex < 0) {
        unknownCommentsId.push(comment.parent)
        comment.indent = 1
        sortedCommentsId.push(comment.id)
      } else {
        comment.indent = commentMap.get(comment.parent)!.indent + 1
        sortedCommentsId.splice(foundIndex + 1, 0, comment.id)
      }
    }
  }

  const sortedComments = sortedCommentsId.map((id) => commentMap.get(id)!)

  return { comments: sortedComments, total, totalPages }
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
