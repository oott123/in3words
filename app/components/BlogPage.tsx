import React from 'react'
import { useTransition } from 'remix'
import BlogLoading from './BlogLoading'

const BlogPage: React.FC<{ className?: string }> = ({
  className,
  children,
}) => {
  const transition = useTransition()
  const isLoading = transition.state === 'loading'

  return (
    <div className={className}>{isLoading ? <BlogLoading /> : children}</div>
  )
}

export default BlogPage
