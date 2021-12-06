import React, { useLayoutEffect, useRef, useState } from 'react'

const BlogSidebar: React.FC = ({ children }) => {
  const [isFixed, setIsFixed] = useState(false)
  const [top, setTop] = useState<number | undefined>(undefined)
  const sidebar = useRef<HTMLDivElement>(null)
  const sidebarHeight = useRef(0)
  const screenHeight = useRef(0)
  const lastHeight = useRef(0)

  useLayoutEffect(() => {
    setIsFixed(true)

    sidebarHeight.current = sidebar.current?.clientHeight ?? 0

    function setScroll() {
      const scrollTop = window.pageYOffset || document.documentElement.scrollTop
      const scrollDiff = lastHeight.current - scrollTop

      setTop((t) =>
        Math.max(
          Math.min((t ?? 0) + scrollDiff, 0),
          screenHeight.current - sidebarHeight.current,
        ),
      )
      lastHeight.current = scrollTop
    }

    const handleScroll = () => {
      setScroll()
    }

    const handleResize = () => {
      screenHeight.current = window.innerHeight
      setScroll()
    }

    window.addEventListener('scroll', handleScroll, { passive: true })
    window.addEventListener('resize', handleResize, { passive: true })
    handleResize()

    return () => {
      window.removeEventListener('scroll', handleScroll)
      window.removeEventListener('resize', handleResize)
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
