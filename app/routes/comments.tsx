import * as yup from 'yup'
import React from 'react'
import { ActionFunction, json, Link, useActionData } from 'remix'
import { postComment, Comment } from '~/data/comments'
import { blogTitle } from '~/utils/meta'
import { MetaFunction } from '~/types'
import BlogPage from '~/components/BlogPage'
import BlogCard from '~/components/BlogCard'
import { isErrorResponse } from '~/data/base'

export const meta: MetaFunction = ({ parentsData: { root } }) => {
  return {
    title: blogTitle('提交评论', root),
  }
}

type CommentResult = {
  returnPath: string
  success: boolean
  error?: string
  comment?: Comment
}

export const action: ActionFunction = async ({ request }) => {
  const formData = await request.formData()
  const returnPath = await yup.string().validate(formData.get('return_path'))
  try {
    const commentSchema = yup
      .object()
      .required()
      .shape({
        content: yup.string().required(),
        post: yup.number().positive().integer().required(),
        parent: yup.number().positive().integer().notRequired().nullable(),
        authorName: yup.string().required(),
        authorEmail: yup.string().email().required(),
        authorUrl: yup.string().url().notRequired().nullable(),
      })
      .noUnknown()

    const comment = await commentSchema.validate({
      content: formData.get('content'),
      post: formData.get('post'),
      parent: formData.get('parent'),
      authorName: formData.get('author_name'),
      authorEmail: formData.get('author_email'),
      authorUrl: formData.get('author_url'),
    })

    const postBody = {
      content: comment.content,
      post: comment.post,
      parent: comment.parent,
      author_name: comment.authorName,
      author_email: comment.authorEmail,
      author_url: comment.authorUrl,
    }
    const created = await postComment(postBody, {
      realIp: request.headers.get('X-Real-IP') || '0.0.0.0',
      userAgent: request.headers.get('User-Agent') || 'Mozilla/5.0',
    })

    return {
      returnPath,
      success: true,
      comment: created,
    } as CommentResult
  } catch (e) {
    return {
      returnPath,
      success: false,
      error: isErrorResponse(e) ? e.message : '未知错误',
    } as CommentResult
  }
}

const CommentPosted: React.FC = () => {
  const data = useActionData<CommentResult>()

  return (
    <BlogPage>
      <BlogCard>
        {data?.success ? (
          <>
            <h1>评论已提交</h1>
            <p>我已收到你的评论。</p>
            {data?.comment?.approved ? (
              <p>由于缓存问题，评论可能不会立即显示，请不要担心。</p>
            ) : (
              <p>
                你的邮箱可能是第一次评论，或者评论未能通过本站的自动审核系统。因此，该评论无法立即可见。
              </p>
            )}
            <p>
              <Link
                to={`${data?.returnPath || '/'}#comment-${
                  data?.comment?.id || 'new'
                }`}
              >
                返回文章
              </Link>
            </p>
          </>
        ) : (
          <>
            <h1>评论提交失败</h1>
            <p>{data?.error || '发生了不明错误'}</p>
            <p>请点击浏览器的返回按钮以返回修改评论。</p>
          </>
        )}
      </BlogCard>
    </BlogPage>
  )
}

export default CommentPosted
