export function postPath({ id }: { id: number }) {
  return `/${id}.moe`
}

export function authorPath({ slug }: { slug: string }) {
  return `/author/${slug}`
}

export function categoryPath({ slug }: { slug: string }, page: number) {
  if (page > 1) {
    return `/category/${slug}/page/${page}`
  }
  return `/category/${slug}`
}

export function tagPath({ slug }: { slug: string }, page: number) {
  if (page > 1) {
    return `/tag/${slug}/page/${page}`
  }
  return `/tag/${slug}`
}

export function pagePath({ slug }: { slug: string }) {
  return `/${slug}`
}

export function indexPath(page: number) {
  if (page === 1) {
    return '/'
  } else if (page > 0) {
    return `/page/${page}`
  } else {
    return `/`
  }
}
