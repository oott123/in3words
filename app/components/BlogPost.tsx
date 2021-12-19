import classNames from 'classnames'
import React, { useEffect, useRef } from 'react'
import { Link } from 'remix'
import { Post, SummarizedPost } from '~/data/posts'
import { authorPath, categoryPath, tagPath } from '~/path'
import BlogCard from './BlogCard'
import BlogDate from './BlogDate'
import BlogIcon, { BlogIcons } from './BlogIcon'
import { LoadingSvg } from './BlogLoading'

const BlogPost: React.FC<{
  post: SummarizedPost | Post
  postPath: string
  isLoadingFull?: boolean
}> = ({ post, postPath, isLoadingFull }) => {
  const summarized = 'summary' in post
  const html = summarized ? post.summary : post.content
  const postBody = useRef<HTMLDivElement>(null)

  useEffect(() => {
    Array.from(postBody.current!.querySelectorAll('script')).forEach(
      (oldScript) => {
        const newScript = document.createElement('script')
        Array.from(oldScript.attributes).forEach((attr) =>
          newScript.setAttribute(attr.name, attr.value),
        )
        newScript.appendChild(document.createTextNode(oldScript.innerHTML))
        oldScript.parentNode!.replaceChild(newScript, oldScript)
      },
    )
  }, [html])

  return (
    <BlogCard>
      <article
        className={classNames('BlogPost', summarized && 'BlogPost-Summarized')}
      >
        <header className="BlogPost_Head">
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
            {!!post.categories?.length && (
              <BlogMeta label={<BlogIcon>{BlogIcons.Category}</BlogIcon>}>
                {post.categories.map((c) => (
                  <Link to={categoryPath(c)} key={c.slug}>
                    {c.name}
                  </Link>
                ))}
              </BlogMeta>
            )}
            {!!post.tags?.length && (
              <BlogMeta label={<BlogIcon>{BlogIcons.Tag}</BlogIcon>}>
                {post.tags.map((t) => (
                  <Link to={tagPath(t)} key={t.slug}>
                    {t.name}
                  </Link>
                ))}
              </BlogMeta>
            )}
          </div>
        </header>
        <section
          className="BlogPost_Body"
          dangerouslySetInnerHTML={{
            __html: html,
          }}
          ref={postBody}
        ></section>
        {summarized && (
          <footer className="BlogPost_ReadMore">
            {isLoadingFull ? (
              <LoadingSvg />
            ) : (
              <Link to={postPath}>阅读全文»</Link>
            )}
          </footer>
        )}
      </article>
    </BlogCard>
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
