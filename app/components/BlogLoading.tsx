import React from 'react'
import BlogCard from './BlogCard'

const BlogLoading: React.FC = () => {
  return (
    <div>
      <BlogCard>
        <LoadingSvg />
      </BlogCard>
    </div>
  )
}

export const LoadingSvg = () => (
  <svg
    role="img"
    width="340"
    height="84"
    aria-labelledby="loading-aria"
    viewBox="0 0 340 84"
    preserveAspectRatio="none"
  >
    <title id="loading-aria">正在加载……</title>
    <rect
      x="0"
      y="0"
      width="100%"
      height="100%"
      clip-path="url(#clip-path)"
      style={{ fill: 'url("#fill")' }}
    ></rect>
    <defs>
      <clipPath id="clip-path">
        <rect x="0" y="0" rx="3" ry="3" width="67" height="11" />
        <rect x="76" y="0" rx="3" ry="3" width="140" height="11" />
        <rect x="127" y="48" rx="3" ry="3" width="53" height="11" />
        <rect x="187" y="48" rx="3" ry="3" width="72" height="11" />
        <rect x="18" y="48" rx="3" ry="3" width="100" height="11" />
        <rect x="0" y="71" rx="3" ry="3" width="37" height="11" />
        <rect x="18" y="23" rx="3" ry="3" width="140" height="11" />
        <rect x="166" y="23" rx="3" ry="3" width="173" height="11" />
      </clipPath>
      <linearGradient id="fill">
        <stop offset="0.599964" stopColor="#e6e6e6" stopOpacity="1">
          <animate
            attributeName="offset"
            values="-2; -2; 1"
            keyTimes="0; 0.25; 1"
            dur="2s"
            repeatCount="indefinite"
          ></animate>
        </stop>
        <stop offset="1.59996" stopColor="#c6c4c4" stopOpacity="1">
          <animate
            attributeName="offset"
            values="-1; -1; 2"
            keyTimes="0; 0.25; 1"
            dur="2s"
            repeatCount="indefinite"
          ></animate>
        </stop>
        <stop offset="2.59996" stopColor="#e6e6e6" stopOpacity="1">
          <animate
            attributeName="offset"
            values="0; 0; 3"
            keyTimes="0; 0.25; 1"
            dur="2s"
            repeatCount="indefinite"
          ></animate>
        </stop>
      </linearGradient>
    </defs>
  </svg>
)

export default BlogLoading
