import React from 'react'
import { useTransition } from 'remix'
import { SummarizedPost } from '~/data/posts'
import { postPath } from '~/path'
import BlogPost from './BlogPost'

const PostList: React.FC<{ posts: Array<SummarizedPost> }> = ({ posts }) => {
  const transition = useTransition()
  const isLoading = transition.state === 'loading'
  const loadingPath = transition.location?.pathname

  return (
    <section className="PostList">
      {posts.map((post) => {
        const path = postPath(post)
        const isPostLoading = loadingPath === path
        if (isLoading && !isPostLoading) return null
        return (
          <BlogPost
            key={post.id}
            post={post}
            postPath={path}
            isLoadingFull={isLoading && isPostLoading}
          />
        )
      })}
    </section>
  )
}

export default PostList
