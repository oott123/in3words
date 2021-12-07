import { format } from 'date-fns'
import React, { useEffect, useLayoutEffect, useState } from 'react'

// React currently throws a warning when using useLayoutEffect on the server.
// To get around it, we can conditionally useEffect on the server (no-op) and
// useLayoutEffect in the browser.
const useIsomorphicLayoutEffect =
  typeof window !== 'undefined' ? useLayoutEffect : useEffect

const BlogDate: React.FC<{ date: Date | string }> = ({ date }) => {
  const shortDate = format(new Date(date), 'yyyy-MM-dd HH:mm')
  const [dateStr, setDateStr] = useState(shortDate)

  useIsomorphicLayoutEffect(() => {
    // local timezone
    setDateStr(format(new Date(date), 'yyyy-MM-dd HH:mm'))
  }, [date])

  return (
    <time dateTime={date.toString()} title={date.toString()}>
      {dateStr}
    </time>
  )
}

export default BlogDate
