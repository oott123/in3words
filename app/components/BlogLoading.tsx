import React from 'react'
import BlogCard from './BlogCard'

const BlogLoading: React.FC = () => {
  return (
    <div>
      <BlogCard>
        <p>正在加载……</p>
      </BlogCard>
    </div>
  )
}

export default BlogLoading
