import classNames from 'classnames'
import React from 'react'

const BlogCard: React.FC<{ className?: string }> = ({
  children,
  className,
}) => {
  return <div className={classNames('BlogCard', className)}>{children}</div>
}

export default BlogCard
