import React, { useState } from 'react'
import BlogCard from './BlogCard'
import type { Comment } from '~/data/comments'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'
import PageNavigation from './PageNavigation'
import { Form, useLocation, useTransition } from 'remix'

const BlogComment: React.FC<{
  comments: Comment[]
  total: number
  page: number
  totalPages: number
  postId: number
}> = ({ comments, page, totalPages, total, postId }) => {
  const { state, location } = useTransition()
  const currentLocation = useLocation()
  const [replyTo, setReplyTo] = useState(0)

  const switchingPage =
    state === 'loading' && currentLocation.pathname === location?.pathname

  return (
    <section className="BlogComment">
      <BlogCard className="BlogComment_H1">
        <h1>评论</h1>
      </BlogCard>
      {switchingPage ? (
        <BlogCard>读取评论……</BlogCard>
      ) : comments.length > 0 ? (
        comments.map((comment) => (
          <React.Fragment key={comment.id}>
            <BlogCard data-indent={comment.indent}>
              <article className="BlogComment_Comment">
                <link href={`#comment-${comment.id}`} />
                <a id={`comment-${comment.id}`} />
                <header className="BlogComment_CommentHeader">
                  <div className="BlogComment_AuthorAvatar">
                    <img src={comment.authorAvatar} alt={comment.authorName} />
                  </div>
                  <div className="BlogComment_AuthorInfo">
                    <span className="BlogComment_AuthorName">
                      <span>{comment.authorName}</span>
                      {comment.author ? (
                        <span
                          title={
                            comment.author === 1 ? '博主' : '身份经过博主确认'
                          }
                        >
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
              </article>
            </BlogCard>
            {replyTo === comment.id && (
              <BlogCard>
                <CommentForm post={postId} parent={comment.id} />
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
      <BlogCard>
        <CommentForm post={postId} />
      </BlogCard>
    </section>
  )
}

const CommentForm: React.FC<{
  post: number
  parent?: number
}> = ({ post, parent }) => {
  return (
    <form method="post" action="/comments" className="CommentForm">
      <h2>{parent ? '回复评论' : '发表评论'}</h2>
      <fieldset>
        <textarea
          name="content"
          rows={3}
          placeholder="评论内容"
          required
        ></textarea>
        <input type="hidden" name="post" value={post} />
        {parent && <input type="hidden" name="parent" value={parent} />}
        <input
          type="text"
          name="author_name"
          placeholder="昵称（公开显示）"
          required
          autoComplete="nickname"
        />
        <input
          type="email"
          name="author_email"
          placeholder="邮箱（显示头像和接收回复邮件）"
          required
          autoComplete="email"
        />
        <input
          type="url"
          name="author_url"
          placeholder="网站（显示在昵称旁边）"
          autoComplete="url"
        />
        <button type="submit">提交</button>
      </fieldset>
      <p className="CommentForm_Tip">
        发表评论代表您授权本网站存储并在必要情况下使用您输入的邮箱地址、连接本站服务器使用的
        IP 地址和用户代理字符串 (User Agent)
        用于发送评论回复邮件，以及将上述信息分享给{' '}
        <a
          href="https://en.gravatar.com/site/terms-of-service"
          target="_blank"
          rel="noopener noreferrer nofollow"
        >
          Gravatar
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
    </form>
  )
}

export default BlogComment
