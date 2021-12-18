import * as yup from 'yup'
import { ActionFunction, redirect, useLoaderData } from 'remix'
import { postComment } from '~/data/comments'
import { blogTitle } from '~/utils/meta'
import { MetaFunction } from '~/types'
import React, { useCallback } from 'react'
import BlogCard from '~/components/BlogCard'
import BlogPage from '~/components/BlogPage'
import { SingleComment } from '~/components/BlogComment'

export const meta: MetaFunction = ({ parentsData: { root } }) => {
  return {
    title: blogTitle('提交评论', root),
  }
}

export const action: ActionFunction = async ({ request }) => {
  const formData = await request.formData()
  const returnPath = await yup.string().validate(formData.get('return_path'))

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

  const created = await postComment(postBody)

  return redirect(`${returnPath || '/'}#comment-${created.id}`)
}
