import React from 'react'
import { useIsPageLoading } from '~/hooks/useIsPageLoading'
import { useScrollToTop } from '~/hooks/useScrollToTop'
import BlogLoading from './BlogLoading'

const BlogPage: React.FC<{ className?: string }> = ({
  className,
  children,
}) => {
  const isLoading = useIsPageLoading()
  useScrollToTop(isLoading)

  return (
    <div className={className}>{isLoading ? <BlogLoading /> : children}</div>
  )
}

export default BlogPage
