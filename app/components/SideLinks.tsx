import classNames from 'classnames'
import React from 'react'
import { Link } from 'remix'

const SideLinks: React.FC = ({ children }) => {
  return <ul className="SideLinks">{children}</ul>
}

export default SideLinks

export type SideLink = {
  title: React.ReactNode
  to: string
  active?: boolean
  className?: string
}

export const SideLinkItem: React.FC<SideLink> = ({
  title,
  to,
  active,
  className,
}) => {
  return (
    <li className="SideLinks_Item">
      <Link
        to={to}
        className={classNames(className, active && 'SideLinks_Item-Active')}
      >
        {title}
      </Link>
    </li>
  )
}
