import BlogListPage, { loader as indexLoader, meta as indexMeta } from './$year'
export default BlogListPage

export const loader = (arg: any) => indexLoader(arg)
export const meta = (arg: any) => indexMeta(arg)
