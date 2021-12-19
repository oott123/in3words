import React from 'react'
import { useTransition } from 'remix'
import { SummarizedPost } from '~/data/posts'
import { postPath } from '~/path'
import BlogCard from './BlogCard'
import BlogLoading from './BlogLoading'
import BlogPost from './BlogPost'

const PostList: React.FC<{ posts: Array<SummarizedPost> }> = ({ posts }) => {
  const transition = useTransition()
  const isLoading = transition.state === 'loading'
  const loadingPath = transition.location?.pathname
  const hasLoadingPost = posts.some((p) => postPath(p) === loadingPath)

  return (
    <section className="PostList">
      {isLoading && !hasLoadingPost ? (
        <BlogLoading />
      ) : posts.length > 0 ? (
        posts.map((post) => {
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
        })
      ) : (
        <BlogCard>
          <p>没有结果</p>
        </BlogCard>
      )}
    </section>
  )
}

export default PostList
