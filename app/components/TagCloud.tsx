import React from 'react'
import { TagMeta } from '~/data/site'

const TagCloud: React.FC<{ tags: TagMeta[] }> = ({ tags }) => {
  const sortedTags = [...tags].sort((a, b) => b.count - a.count)
  const maxCount =
    sortedTags.length > 1 ? sortedTags[1].count : sortedTags[0].count
  const minCount = Math.min(...tags.map((tag) => tag.count))
  const tagWithRate = tags.map((tag) => ({
    ...tag,
    rate: Math.min((tag.count - minCount) / (maxCount - minCount), 1.4),
  }))

  return (
    <ul className="TagCloud">
      {tagWithRate.map((tag) => (
        <li
          key={tag.id}
          style={{ fontSize: `${(tag.rate * 0.5 + 1).toFixed(2)}em` }}
        >
          <a href={`/tag/${tag.slug}`} title={`${tag.name} (${tag.count})`}>
            {tag.name}
          </a>
        </li>
      ))}
    </ul>
  )
}

export default TagCloud
