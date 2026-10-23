import classNames from 'classnames'
import React from 'react'

const BlogCard: React.FC<React.PropsWithChildren<{ className?: string }>> = ({
  children,
  className,
  ...props
}) => {
  return (
    <div className={classNames('BlogCard', className)} {...props}>
      {children}
    </div>
  )
}

export default BlogCard
