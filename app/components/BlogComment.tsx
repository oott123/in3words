import React from 'react'
import BlogCard from './BlogCard'
import type { Comment } from '~/data/comments'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'

const BlogComment: React.FC<{
  comments: Comment[]
  page: number
  totalPages: number
}> = ({ comments, page, totalPages }) => {
  return (
    <section className="BlogComment">
      <BlogCard className="BlogComment_H1">
        <h1>评论</h1>
      </BlogCard>
      {comments.length > 0 ? (
        comments.map((comment) => (
          <BlogCard data-indent={comment.indent}>
            <article key={comment.id} className="BlogComment_Comment">
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
    </section>
  )
}

export default BlogComment
