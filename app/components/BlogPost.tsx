import classNames from 'classnames'
import React from 'react'
import { Link, useTransition } from 'remix'
import { Post, SummarizedPost } from '~/data/posts'
import { authorPath, categoryPath, tagPath } from '~/path'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'

const BlogPost: React.FC<{ post: SummarizedPost | Post; postPath: string }> = ({
  post,
  postPath,
}) => {
  const summarized = 'summary' in post
  const transition = useTransition()
  const isLoadingFull = summarized && transition.location?.pathname === postPath

  return (
    <article
      className={classNames('BlogPost', summarized && 'BlogPost-Summarized')}
    >
      <div className="BlogPost_Head">
        <h1 className="BlogPost_Title">
          <Link
            to={postPath}
            dangerouslySetInnerHTML={{ __html: post.title }}
          />
        </h1>
        <div className="BlogPost_Meta">
          <BlogMeta label={<BlogIcon>{BlogIcons.User}</BlogIcon>}>
            <Link to={authorPath(post.author)}>{post.author.name}</Link>
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Date}</BlogIcon>}>
            <Link to={postPath}>
              <BlogDate date={post.createdAt} />
            </Link>
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Category}</BlogIcon>}>
            {post.categories.map((c) => (
              <Link to={categoryPath(c)} key={c.slug}>
                {c.name}
              </Link>
            ))}
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Tag}</BlogIcon>}>
            {post.tags.map((t) => (
              <Link to={tagPath(t)} key={t.slug}>
                {t.name}
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
          <Link to={postPath}>
            {isLoadingFull ? '正在加载……' : '阅读全文»'}
          </Link>
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
