import React from 'react'
import { Link } from 'react-router-dom'
import BlogCard from './BlogCard'

const BlogFooter: React.FC = () => {
  return (
    <footer className="BlogFooter">
      <BlogCard>
        <p>
          <Link to="/">三言三语</Link>，由{' '}
          <a href="https://www.oott123.com" target="_blank">
            oott123
          </a>{' '}
          设计与创作，并使用{' '}
          <a
            href="https://wordpress.org"
            target="_blank"
            rel="noopener noreferrer nofollow"
          >
            WordPress
          </a>
          ，
          <a
            href="https://reactjs.org"
            target="_blank"
            rel="noopener noreferrer nofollow"
          >
            React
          </a>{' '}
          和{' '}
          <a
            href="https://remix.run"
            target="_blank"
            rel="noopener noreferrer nofollow"
          >
            Remix
          </a>{' '}
          构建。
        </p>
        <p>
          <a href="https://ainou.com" target="_blank">
            Ainou
          </a>{' '}
          是由 Ashley 绘制的原创角色，受独立版权条款保护，不适用于 CC 协议。
        </p>
      </BlogCard>
    </footer>
  )
}

export default BlogFooter
