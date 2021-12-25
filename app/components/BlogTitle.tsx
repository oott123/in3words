import { Link } from 'remix'
import React from 'react'

const BlogTitle: React.FC<{ title: string; headline: string }> = ({
  title,
  headline,
}) => {
  return (
    <div className="BlogTitle">
      <h1>
        <Link to="/" prefetch="render">
          {title}
        </Link>
      </h1>
      <h2>{headline}</h2>
    </div>
  )
}

export default BlogTitle
