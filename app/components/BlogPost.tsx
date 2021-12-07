import classNames from 'classnames'
import React from 'react'
import { Link } from 'remix'
import { Post, SummarizedPost } from '~/data/posts'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'

const BlogPost: React.FC<{ post: SummarizedPost | Post }> = ({ post }) => {
  const summarized = 'summary' in post

  return (
    <article
      className={classNames('BlogPost', summarized && 'BlogPost-Summarized')}
    >
      <div className="BlogPost_Head">
        <h1 className="BlogPost_Title">
          <Link
            to={`/${post.id}.moe`}
            dangerouslySetInnerHTML={{ __html: post.title }}
          />
        </h1>
        <div className="BlogPost_Meta">
          <BlogMeta label={<BlogIcon>{BlogIcons.User}</BlogIcon>}>
            <Link to={`/author/${post.author.slug}`}>{post.author.name}</Link>
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Date}</BlogIcon>}>
            <Link to={`/${post.id}.moe`}>
              <BlogDate date={post.createdAt} />
            </Link>
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Category}</BlogIcon>}>
            {post.categories.map((x) => (
              <Link to={`/category/${x.slug}`} key={x.slug}>
                {x.name}
              </Link>
            ))}
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Tag}</BlogIcon>}>
            {post.tags.map((x) => (
              <Link to={`/tag/${x.slug}`} key={x.slug}>
                {x.name}
              </Link>
            ))}
          </BlogMeta>
        </div>
      </div>
      <div
        className="BlogPost_Body"
        dangerouslySetInnerHTML={{
          __html: summarized ? post.summary : post.content,
        }}
      ></div>
      {summarized && (
        <div className="BlogPost_ReadMore">
          <Link to={`/${post.id}.moe`}>阅读全文»</Link>
        </div>
      )}
    </article>
  )
}

export default BlogPost

const BlogMeta: React.FC<{
  label: React.ReactNode
  children: React.ReactNode
}> = ({ label, children }) => {
  return (
    <div className="BlogMeta">
      <div className="BlogMeta_Label">{label}</div>
      <div className="BlogMeta_Value">{children}</div>
    </div>
  )
}
