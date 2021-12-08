export function postPath({ id }: { id: number }) {
  return `/${id}.moe`
}

export function authorPath({ slug }: { slug: string }) {
  return `/author/${slug}`
}

export function categoryPath({ slug }: { slug: string }) {
  return `/category/${slug}`
}

export function tagPath({ slug }: { slug: string }) {
  return `/tag/${slug}`
}

export function pagePath({ slug }: { slug: string }) {
  return `/${slug}`
}
