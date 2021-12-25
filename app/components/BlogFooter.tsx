import React from 'react'
import { Link } from 'remix'
import BlogCard from './BlogCard'
import ExternalLink from './ExternalLink'

const BlogFooter: React.FC = () => {
  return (
    <footer className="BlogFooter">
      <BlogCard>
        <p>
          <Link to="/">三言三语</Link>，由{' '}
          <ExternalLink href="https://oott123.com">oott123</ExternalLink>{' '}
          设计与创作，
          <ExternalLink href="https://www.upyun.com/">又拍云</ExternalLink>
          提供 CDN 服务。
        </p>
        <p>
          <ExternalLink href="https://github.com/oott123/in3words">
            in3words
          </ExternalLink>
          ，使用{' '}
          <ExternalLink href="https://wordpress.org">WordPress</ExternalLink>，
          <ExternalLink href="https://reactjs.org">React</ExternalLink> 和{' '}
          <ExternalLink href="https://remix.run">Remix</ExternalLink> 构建。
        </p>
      </BlogCard>
    </footer>
  )
}

export default BlogFooter
