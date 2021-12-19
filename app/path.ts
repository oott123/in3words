export function postPath({ id }: { id: number }) {
  return `/${id}.moe`
}

export function authorPath({ slug }: { slug: string }, page = 1) {
  if (page > 1) {
    return `/author/${slug}/page/${page}`
  }
  return `/author/${slug}`
}

export function categoryPath({ slug }: { slug: string }, page = 1) {
  if (page > 1) {
    return `/category/${slug}/page/${page}`
  }
  return `/category/${slug}`
}

export function tagPath({ slug }: { slug: string }, page = 1) {
  if (page > 1) {
    return `/tag/${slug}/page/${page}`
  }
  return `/tag/${slug}`
}

export function pagePath({ slug }: { slug: string }) {
  return `/${slug}`
}

export function indexPath(page = 1, keyword?: string) {
  if (page > 1) {
    return `/page/${page}${keyword ? `?s=${keyword}` : ''}`
  } else {
    return `/${keyword ? `?s=${keyword}` : ''}`
  }
}

export function archivePath(
  { year, month, day }: { year: number; month?: number; day?: number },
  page = 1,
) {
  const segments = ['/date', year]
  if (month) {
    segments.push(month)
    if (day) {
      segments.push(day)
    }
  }
  if (page > 1) {
    segments.push(`page/${page}`)
  }
  return segments.join('/')
}
