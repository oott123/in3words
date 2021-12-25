import React from 'react'

const ExternalLink: React.FC<{ href: string }> = ({ href, children }) => {
  return (
    <a href={href} target="_blank" rel="noopener noreferrer nofollow">
      {children}
    </a>
  )
}

export default ExternalLink
