import classNames from 'classnames'
import React from 'react'
import { Link, useTransition } from 'remix'

const PageNavigation: React.FC<{
  totalPages: number
  page: number
  path: (page: number) => string
}> = ({ totalPages, page, path, children }) => {
  const transition = useTransition()
  const isLoading = transition.state === 'loading'
  const showPages = getShowPages(page, totalPages)

  return isLoading ? null : (
    <div className="PageNavigation">
      {children}
      <nav className="PageNavigation_Nav">
        {showPages.map((p) => (
          <Link
            key={p}
            to={path(p)}
            className={classNames({
              PageNavigation_Link: true,
              'PageNavigation_Link-Active': p === page,
              'PageNavigation_Link-Disabled': p === page || p <= 0,
            })}
          >
            {p > 0
              ? p === page + 1
                ? '下一页'
                : p === page - 1
                ? '上一页'
                : p
              : '...'}
          </Link>
        ))}
      </nav>
    </div>
  )
}

export default PageNavigation

function getShowPages(page: number, totalPages: number) {
  const showPages = [
    1,
    // page - 2,
    page - 1,
    page,
    page + 1,
    // page + 2,
    totalPages,
  ]
    .filter((p) => p > 0 && p <= totalPages)
    .filter((p, i, a) => a.indexOf(p) === i)

  for (let i = 1; i < showPages.length; i++) {
    if (showPages[i] - showPages[i - 1] > 1) {
      showPages.splice(i, 0, -i)
      i++
    }
  }
  return showPages
}
