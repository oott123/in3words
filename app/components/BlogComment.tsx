import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import BlogCard from './BlogCard'
import type { Comment } from '~/data/comments'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'
import PageNavigation from './PageNavigation'
import { Form, useFetcher, useLocation, useTransition } from 'remix'
import classNames from 'classnames'

export const SingleComment: React.FC<{ comment: Comment }> = ({
  comment,
  children,
}) => {
  const preventClick = useCallback((e) => e.preventDefault(), [])

  return (
    <article className="BlogComment_Comment">
      <link href={`#comment-${comment.id}`} />
      <a id={`comment-${comment.id}`} />
      <header className="BlogComment_CommentHeader">
        <div className="BlogComment_AuthorAvatar">
          <img
            src={comment.authorAvatar}
            alt={comment.authorName}
            loading="lazy"
            referrerPolicy="no-referrer"
          />
        </div>
        <div className="BlogComment_AuthorInfo">
          <span className="BlogComment_AuthorName">
            <span>{comment.authorName}</span>
            {comment.author ? (
              <span title={comment.author === 1 ? '博主' : '身份经过博主确认'}>
                <BlogIcon>{BlogIcons.Verified}</BlogIcon>
              </span>
            ) : null}
            {comment.authorUrl ? (
              <a
                className="BlogComment_AuthorLink"
                href={comment.authorUrl}
                rel="noopener noreferrer nofollow"
                target="_blank"
              >
                {new URL(comment.authorUrl).hostname}
              </a>
            ) : null}
          </span>
          <a
            href={`#comment-${comment.id}`}
            onClick={preventClick}
            className="BlogComment_Date"
          >
            <BlogDate date={comment.date} />
          </a>
        </div>
      </header>
      <section
        className="BlogComment_Content"
        dangerouslySetInnerHTML={{ __html: comment.content }}
      ></section>
      {children}
    </article>
  )
}

const BlogComment: React.FC<{
  comments: Comment[]
  total: number
  page: number
  totalPages: number
  postId: number
  newCommentsAllowed: boolean
}> = ({ comments, page, totalPages, total, postId, newCommentsAllowed }) => {
  const { state, location } = useTransition()
  const currentLocation = useLocation()
  const [replyTo, setReplyTo] = useState(0)
  const [showCommentTip, setShowCommentTip] = useState(false)

  const switchingPage =
    state === 'loading' && currentLocation.pathname === location?.pathname

  useEffect(() => {
    const locationCommentId = currentLocation.hash.startsWith('#comment-')
      ? Number(currentLocation.hash.substring('#comment-'.length))
      : null
    const showCommentTip =
      locationCommentId !== null &&
      !comments.some((c) => c.id === locationCommentId)
    setShowCommentTip(showCommentTip)
  }, [currentLocation.hash, comments])

  return (
    <section className="BlogComment">
      {comments.length < 1 && !newCommentsAllowed ? null : (
        <>
          <BlogCard className="BlogComment_H1">
            <h1>评论</h1>
          </BlogCard>
          {switchingPage ? (
            <BlogCard>读取评论……</BlogCard>
          ) : comments.length > 0 ? (
            comments.map((comment) => (
              <React.Fragment key={comment.id}>
                <BlogCard data-indent={comment.indent}>
                  <SingleComment comment={comment}>
                    <footer className="BlogComment_Actions">
                      {comment.id === replyTo ? (
                        <button
                          className="Button-Link"
                          onClick={() => setReplyTo(0)}
                        >
                          取消回复
                        </button>
                      ) : (
                        <button
                          className="Button-Link"
                          onClick={() => setReplyTo(comment.id)}
                        >
                          回复
                        </button>
                      )}
                    </footer>
                  </SingleComment>
                </BlogCard>
                {replyTo === comment.id && (
                  <BlogCard>
                    <CommentForm
                      post={postId}
                      parent={comment.id}
                      onSuccess={() => setReplyTo(0)}
                    />
                  </BlogCard>
                )}
              </React.Fragment>
            ))
          ) : (
            <BlogCard>
              <p>还没有评论。</p>
            </BlogCard>
          )}
          {totalPages > 1 ? (
            <PageNavigation
              page={page}
              totalPages={totalPages}
              path={(page) => `?comments_page=${page}`}
            >
              共 {total} 条
            </PageNavigation>
          ) : null}
          {showCommentTip && (
            <BlogCard>
              <p>
                如果你刚才提交的评论没有出现在此处，可能是因为评论需要审核或是缓存尚未刷新，无需重复提交。
              </p>
            </BlogCard>
          )}
          {newCommentsAllowed ? (
            <BlogCard className="BlogComment_CommentForm">
              <CommentForm post={postId} />
            </BlogCard>
          ) : (
            <BlogCard>
              <p>评论已关闭。</p>
            </BlogCard>
          )}
        </>
      )}
    </section>
  )
}

const CommentForm: React.FC<{
  post: number
  parent?: number
  onSuccess?: () => void
}> = ({ post, parent, onSuccess }) => {
  const nameInput = useMemorizedInput('commenterName')
  const emailInput = useMemorizedInput('commenterEmail')
  const urlInput = useMemorizedInput('commenterUrl')
  const commentInput = useMemorizedInput('comment')
  const location = useLocation()
  const [hasMemorized, setHasMemorized] = useState(false)

  const comment = useFetcher()
  const commentSuccess = comment.data?.success
  const commentApproved = commentSuccess && comment.data?.comment?.approved
  const commentIsSpam = commentSuccess && comment.data?.comment?.isSpam
  const commentError = comment.data?.error

  const isSubmitting =
    comment.state === 'submitting' || comment.state === 'loading'

  useEffect(() => {
    if (!isSubmitting && commentSuccess) {
      commentInput.setValue('')
      if (commentApproved) {
        onSuccess && onSuccess()
      }
    }
  }, [
    isSubmitting,
    commentSuccess,
    commentInput,
    commentInput.setValue,
    onSuccess,
    commentApproved,
  ])

  useEffect(() => {
    if (nameInput.value && emailInput.value) {
      setHasMemorized(true)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const handleExpandDetails = useCallback(() => {
    setHasMemorized(false)
  }, [])

  return (
    <comment.Form method="post" action="/comments" className="CommentForm">
      <h2>{parent ? '回复评论' : '发表评论'}</h2>
      <fieldset disabled={isSubmitting}>
        <textarea
          name="content"
          rows={3}
          placeholder="评论内容"
          required
          {...commentInput.dom}
        ></textarea>
        <input type="hidden" name="post" value={post} />
        <input type="hidden" name="return_path" value={location.pathname} />
        {parent && <input type="hidden" name="parent" value={parent} />}
        {hasMemorized && (
          <p>
            以 {nameInput.value} 的身份发表评论。
            <button className="Button-Link" onClick={handleExpandDetails}>
              修改
            </button>
          </p>
        )}
        <div
          className={classNames({ 'CommentForm_Details-Hidden': hasMemorized })}
        >
          <input
            type="text"
            name="author_name"
            placeholder="昵称（公开显示）"
            required
            autoComplete="nickname"
            {...nameInput.dom}
          />
          <input
            type="email"
            name="author_email"
            placeholder="邮箱（显示头像和接收回复邮件）"
            required
            autoComplete="email"
            {...emailInput.dom}
          />
          <input
            type="url"
            name="author_url"
            placeholder="网站（显示在昵称旁边）"
            autoComplete="url"
            {...urlInput.dom}
          />
        </div>
        {commentSuccess && (
          <div className="CommentForm_Status CommentForm_Status-Success">
            {commentApproved
              ? '评论发表成功。'
              : commentIsSpam
              ? '评论已提交，但被系统判定为无用评论。这应该是误判，人工复核后很快就能显示。'
              : '评论已提交，但审核后才会显示。'}
          </div>
        )}
        {commentError && (
          <div className="CommentForm_Status CommentForm_Status-Error">
            评论失败：{commentError}
          </div>
        )}
        <button type="submit">{isSubmitting ? '提交中……' : '提交'}</button>
      </fieldset>
      <p className="CommentForm_Tip">
        发表评论代表你授权本网站存储并在必要情况下使用你输入的邮箱地址、连接本站服务器使用的
        IP 地址和用户代理字符串 (User Agent)
        用于发送评论回复邮件，以及将上述信息分享给{' '}
        <a
          href="https://www.libravatar.org/privacy/"
          target="_blank"
          rel="noopener noreferrer nofollow"
        >
          Libravatar
        </a>{' '}
        和{' '}
        <a
          href="https://akismet.com/privacy/"
          target="_blank"
          rel="noopener noreferrer nofollow"
        >
          Akismet
        </a>
        ，用于显示头像和反垃圾。
      </p>
    </comment.Form>
  )
}

function useMemorizedInput(key: string) {
  const localStorage: Record<string, string | undefined> = useMemo(
    () => (typeof window === 'undefined' ? {} : window.localStorage),
    [],
  )
  const prefix = 'i3w_'
  const lsKey = `${prefix}${key}`
  const [value, setValue] = useState(localStorage[lsKey] || '')

  const onChange = useCallback((e: React.ChangeEvent) => {
    setValue((e.target as HTMLInputElement).value)
  }, [])

  const mySetValue = useCallback(
    (value: string) => {
      setValue(value)
      localStorage[lsKey] = value
    },
    [lsKey, localStorage],
  )

  useEffect(() => {
    localStorage[lsKey] = value
  }, [value, lsKey, localStorage])

  return { dom: { value, onChange }, setValue: mySetValue, value }
}

export default BlogComment
