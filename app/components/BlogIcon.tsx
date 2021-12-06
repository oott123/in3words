import { AllApplication, Calendar, TagOne, User } from '@icon-park/svg'
import type { Icon as IconPark } from '@icon-park/svg/lib/runtime'

import React from 'react'

export enum BlogIcons {
  Category,
  Date,
  Tag,
  User,
}

function getIcons(): Record<
  BlogIcons,
  { icon: IconPark } & Parameters<IconPark>[0]
> {
  return {
    [BlogIcons.Category]: { icon: AllApplication, size: 18, strokeWidth: 3 },
    [BlogIcons.Date]: { icon: Calendar, size: 18, strokeWidth: 3 },
    [BlogIcons.Tag]: { icon: TagOne, size: 18, strokeWidth: 3 },
    [BlogIcons.User]: { icon: User, size: 18, strokeWidth: 3 },
  }
}

const symbolPrefix = '__BlogIcon_'

export const BlogIconsDef: React.FC = () => {
  const html = Object.entries(getIcons())
    .map(([id, cfg]) =>
      svgToSymbol(cfg.icon(cfg), symbolPrefix + BlogIcons[parseInt(id)]),
    )
    .join('')

  return (
    <svg style={{ display: 'none' }}>
      <defs
        dangerouslySetInnerHTML={{
          __html: html,
        }}
      ></defs>
    </svg>
  )
}

const BlogIcon = ({ children }: { children: BlogIcons }) => {
  const size = getIcons()[children].size
  return (
    <span className={`BlogIcon BlogIcon-${BlogIcons[children]}`}>
      <svg width={size ?? 18} height={size ?? 18}>
        <use xlinkHref={`#${symbolPrefix}${BlogIcons[children]}`} />
      </svg>
    </span>
  )
}

function svgToSymbol(xml: string, id: string) {
  return xml
    .replace(/^<\?xml version="1\.0" encoding="UTF-8"\?>/, '')
    .replace(/^<svg/, '<symbol id=' + JSON.stringify(id))
    .replace(/<\/svg>$/, '</symbol>')
}

export default BlogIcon
