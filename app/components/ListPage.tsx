import React from 'react'

const ListPage: React.FC<{ className?: string }> = ({
  className,
  children,
}) => {
  return <div className={className}>{children}</div>
}

export default ListPage
