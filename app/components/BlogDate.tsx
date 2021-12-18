import { format, parseISO } from 'date-fns'
import React, { useEffect, useLayoutEffect, useState } from 'react'

// React currently throws a warning when using useLayoutEffect on the server.
// To get around it, we can conditionally useEffect on the server (no-op) and
// useLayoutEffect in the browser.
const useIsomorphicLayoutEffect =
  typeof window !== 'undefined' ? useLayoutEffect : useEffect

const BlogDate: React.FC<{ date: Date | string }> = ({ date }) => {
  date = date instanceof Date ? date : parseISO(date)
  const shortDate = format(date, 'yyyy-MM-dd HH:mm')
  return (
    <time dateTime={date.toISOString()} title={date.toISOString()}>
      {shortDate}
    </time>
  )
}

export default BlogDate
