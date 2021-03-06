import React, { useEffect, useLayoutEffect, useRef, useState } from 'react'

// React currently throws a warning when using useLayoutEffect on the server.
// To get around it, we can conditionally useEffect on the server (no-op) and
// useLayoutEffect in the browser.
const useIsomorphicLayoutEffect =
  typeof window !== 'undefined' ? useLayoutEffect : useEffect

const BlogSidebar: React.FC = ({ children }) => {
  const [isFixed, setIsFixed] = useState(false)
  const [top, setTop] = useState<number | undefined>(undefined)
  const sidebar = useRef<HTMLDivElement>(null)
  const sidebarHeight = useRef(0)
  const screenHeight = useRef(0)
  const lastHeight = useRef(0)

  useIsomorphicLayoutEffect(() => {
    function resizeSidebar() {
      sidebarHeight.current = sidebar.current?.clientHeight ?? 0
      document.body.style.minHeight = `${sidebarHeight.current}px`
    }

    function setScroll() {
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop
      const scrollDiff = lastHeight.current - scrollTop

      if (screenHeight.current < sidebarHeight.current) {
        setTop((t) =>
          Math.max(
            Math.min((t ?? 0) + scrollDiff, 0),
            screenHeight.current - sidebarHeight.current,
          ),
        )
      }
      lastHeight.current = scrollTop
    }

    const handleScroll = () => {
      setScroll()
    }

    const handleResize = () => {
      screenHeight.current = window.innerHeight

      setScroll()
      if (screenHeight.current >= sidebarHeight.current) {
        setTop(0)
      }
    }

    const resizeObserver = new ResizeObserver(() => {
      resizeSidebar()
    })

    setIsFixed(true)
    window.addEventListener('scroll', handleScroll, { passive: true })
    window.addEventListener('resize', handleResize, { passive: true })
    resizeSidebar()
    handleResize()
    resizeObserver.observe(sidebar.current!)

    const scrollTop = window.pageYOffset || document.documentElement.scrollTop
    if (screenHeight.current < sidebarHeight.current) {
      setTop(Math.max(-scrollTop, screenHeight.current - sidebarHeight.current))
    }

    return () => {
      window.removeEventListener('scroll', handleScroll)
      window.removeEventListener('resize', handleResize)
      resizeObserver.disconnect()
    }
  }, [])

  return (
    <div
      className="BlogSidebar"
      style={{
        position: isFixed ? 'fixed' : undefined,
        transform:
          top === undefined ? undefined : `translate3d(0, ${top ?? 0}px, 0)`,
      }}
      ref={sidebar}
    >
      {children}
    </div>
  )
}

export default BlogSidebar
