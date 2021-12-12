import classNames from 'classnames'
import React from 'react'
import { Link, useLocation, useResolvedPath, useTransition } from 'remix'

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
  active: activeInput,
  className,
}) => {
  const location = useLocation()
  const transition = useTransition()
  const path = useResolvedPath(to)

  const locationPathname = (
    transition.state === 'loading'
      ? transition.location.pathname
      : location.pathname
  ).toLowerCase()
  const toPathname = path.pathname.toLowerCase()

  const isActive = checkActive(locationPathname, toPathname)

  const active = activeInput ?? isActive

  return (
    <li className="SideLinks_Item">
      <Link
        to={to}
        className={classNames(className, active && 'SideLinks_Item-Active')}
        aria-current={active ? 'page' : undefined}
      >
        {title}
      </Link>
    </li>
  )
}
function checkActive(locationPathname: string, toPathname: string) {
  return (
    locationPathname === toPathname ||
    (locationPathname.startsWith(toPathname) &&
      locationPathname.charAt(toPathname.length) === '/')
  )
}
