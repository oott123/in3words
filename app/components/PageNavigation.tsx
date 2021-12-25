import classNames from 'classnames'
import React, { useEffect, useRef } from 'react'
import { Link, PrefetchPageLinks, useTransition } from 'remix'

const PageNavigation: React.FC<{
  totalPages: number
  page: number
  path: (page: number) => string
}> = ({ totalPages, page, path, children }) => {
  const transition = useTransition()
  const isLoading = transition.state === 'loading'
  const showPages = getShowPages(page, totalPages)
  const hasNextPage = page + 1 <= totalPages
  const nextPagePath = path(page + 1)
  const [prefetchNextPage, setPrefetchNextPage] = React.useState(false)
  const navRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!navRef.current || !hasNextPage || prefetchNextPage) {
      return
    }
    const intersectionObserver = new IntersectionObserver(function (entries) {
      if (entries[0].intersectionRatio <= 0) {
        return
      }
      setPrefetchNextPage(true)
    })
    intersectionObserver.observe(navRef.current)
    return () => intersectionObserver.disconnect()
  }, [hasNextPage, prefetchNextPage])

  return isLoading ? null : (
    <div className="PageNavigation">
      {children}
      <nav className="PageNavigation_Nav" ref={navRef}>
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
      {hasNextPage && prefetchNextPage && (
        <PrefetchPageLinks page={nextPagePath} />
      )}
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
