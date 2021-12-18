import * as yup from 'yup'
import { ActionFunction, json, redirect, useLoaderData } from 'remix'
import { postComment } from '~/data/comments'
import { pagePath, postPath } from '~/path'
import { blogTitle } from '~/utils/meta'
import { MetaFunction } from '~/types'

export const meta: MetaFunction = ({ parentsData: { root } }) => {
  return {
    title: blogTitle('提交评论', root),
  }
}

export const action: ActionFunction = async ({ request }) => {
  const formData = await request.formData()

  const commentSchema = yup
    .object()
    .required()
    .shape({
      content: yup.string().required(),
      post: yup.number().positive().integer().required(),
      parent: yup.number().positive().integer().notRequired(),
      authorName: yup.string().required(),
      authorEmail: yup.string().email().required(),
      authorUrl: yup.string().url().notRequired(),
    })
    .noUnknown()

  const comment = await commentSchema.validate({
    content: formData.get('content'),
    post: formData.get('post'),
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

  await postComment(postBody)

  return {}
}

import React, { useCallback } from 'react'
import BlogCard from '~/components/BlogCard'
import BlogPage from '~/components/BlogPage'

const Comments: React.FC = () => {
  const back = useCallback(() => {
    history.back()
  }, [])

  return (
    <BlogPage>
      <BlogCard>
        <h1>谢谢！</h1>
        <p>评论已经提交成功。</p>
        <p>由于缓存和审核问题，您的评论可能无法马上显示在文章下方。</p>
        <p>
          您可以现在可以
          <button className="Button-Link" onClick={back}>
            返回上一页
          </button>
          了。
        </p>
      </BlogCard>
    </BlogPage>
  )
}

export default Comments
