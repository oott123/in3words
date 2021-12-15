import React from 'react'
import BlogCard from './BlogCard'
import type { Comment } from '~/data/comments'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'
import PageNavigation from './PageNavigation'
import { useLocation, useTransition } from 'remix'

const BlogComment: React.FC<{
  comments: Comment[]
  total: number
  page: number
  totalPages: number
}> = ({ comments, page, totalPages, total }) => {
  const { state, location } = useTransition()
  const currentLocation = useLocation()

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
          <BlogCard key={comment.id} data-indent={comment.indent}>
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
                <button>回复</button>
              </footer>
            </article>
          </BlogCard>
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
    </section>
  )
}

export default BlogComment
