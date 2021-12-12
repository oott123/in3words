import React from 'react'

const BlogPage: React.FC<{ className?: string }> = ({
  className,
  children,
}) => {
  return <div className={className}>{children}</div>
}

export default BlogPage
