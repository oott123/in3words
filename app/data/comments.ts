import classNames from 'classnames'
import { Parser } from 'htmlparser2'
import {
  encodeHtmlAttr,
  encodeHtmlText,
  getList,
  parseGmt,
  post,
  replaceMediaUrl,
  RequestOptions,
  voidTags,
} from './base'
import { clear as clearCache } from './cache'

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
  approved: boolean
  isSpam: boolean
}

export type NewComment = {
  content: string
  post: number
  parent?: number | null
  author_name: string
  author_email: string
  author_url?: string | null
}

export function commentCacheGroup(postId: number) {
  return `comments:post-${postId}`
}

export async function getComments(
  postId: number,
  page = 1,
): Promise<{ comments: Comment[]; total: number; totalPages: number }> {
  const {
    items: commentsData,
    totalPages,
    total,
  } = await getList(
    '/comments',
    {
      post: postId,
      page: page,
      order: 'desc',
      per_page: 100,
    },
    { cacheGroup: commentCacheGroup(postId) },
  )

  const comments = commentsData.map(mapComment).sort((a, b) => a.id - b.id)

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

function mapComment(comment: any): Comment {
  return {
    id: comment.id,
    post: comment.post,
    parent: comment.parent,
    author: comment.author,
    authorAvatar: replaceMediaUrl(comment.author_avatar_urls['96']),
    authorUrl: comment.author_url,
    authorName: comment.author_name,
    date: parseGmt(comment.date_gmt),
    content: postProcessContent(comment.content.rendered).content,
    indent: 0,
    approved: comment.status === 'approved',
    isSpam: comment.status === 'spam',
  }
}

export async function postComment(body: NewComment, options?: RequestOptions) {
  const resp = await post('/comments', body, options)

  await clearCache(commentCacheGroup(body.post))

  return mapComment(resp)
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
