import React from 'react'
import { Link } from 'react-router'
import BlogCard from './BlogCard'
import ExternalLink from './ExternalLink'

const BlogFooter: React.FC = () => {
  return (
    <footer className="BlogFooter">
      <BlogCard>
        <p>
          <Link to="/">三言三语</Link>，由{' '}
          <ExternalLink href="https://oott123.com">oott123</ExternalLink>{' '}
          设计与创作。
        </p>
        <p>
          <ExternalLink href="https://github.com/oott123/in3words">
            in3words
          </ExternalLink>
          ，使用{' '}
          <ExternalLink href="https://wordpress.org">WordPress</ExternalLink>，
          <ExternalLink href="https://react.dev">React</ExternalLink> 和{' '}
          <ExternalLink href="https://reactrouter.com">
            React Router
          </ExternalLink>{' '}
          构建。
        </p>
      </BlogCard>
    </footer>
  )
}

export default BlogFooter
