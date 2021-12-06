import React from 'react'
import { Link } from 'remix'
import { Post, SummarizedPost } from '~/data/posts'
import BlogIcon, { BlogIcons } from './BlogIcon'

const BlogPost: React.FC<{ post: SummarizedPost | Post }> = ({ post }) => {
  return (
    <article className="BlogPost">
      <div className="BlogPost_Head">
        <h1 className="BlogPost_Title">{post.title}</h1>
        <div className="BlogPost_Meta">
          <BlogMeta label={<BlogIcon>{BlogIcons.User}</BlogIcon>}>
            <a href="">{post.author.name}</a>
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Date}</BlogIcon>}>
            {post.updatedAt}
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Category}</BlogIcon>}>
            {post.categories.map((x) => x.name).join(',')}
          </BlogMeta>
          <BlogMeta label={<BlogIcon>{BlogIcons.Tag}</BlogIcon>}>
            {post.tags.map((x) => x.name).join(',')}
          </BlogMeta>
        </div>
      </div>
      <div
        className="BlogPost_Body"
        dangerouslySetInnerHTML={{
          __html: 'summary' in post ? post.summary : post.content,
        }}
      ></div>
      {'summary' in post && (
        <div className="BlogPost_ReadMore">
          <Link to="/">阅读全文»</Link>
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
