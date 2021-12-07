import { format } from 'date-fns'
import React, { useEffect, useState } from 'react'

const BlogDate: React.FC<{ date: Date | string }> = ({ date }) => {
  const shortDate = format(new Date(date), 'yyyy-MM-dd HH:mm')
  const [dateStr, setDateStr] = useState(shortDate)

  useEffect(() => {
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
