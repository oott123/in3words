import {
  type RouteConfig,
  type RouteConfigEntry,
  index,
  route,
} from '@react-router/dev/routes'

// 同一个路由模块挂到多个路径上时，需要给每个路径一个唯一的 id
function paged(path: string, file: string): RouteConfigEntry[] {
  return [
    route(path, file, { id: `${file}:${path}` }),
    route(`${path}/page/:page`, file, { id: `${file}:${path}/page` }),
  ]
}

export default [
  index('routes/index.tsx', { id: 'routes/index' }),
  route('page/:page', 'routes/index.tsx', { id: 'routes/page' }),
  ...paged('author/:slug', 'routes/author/$slug/index.tsx'),
  ...paged('category/:slug', 'routes/category/$slug/index.tsx'),
  ...paged('tag/:slug', 'routes/tag/$slug/index.tsx'),
  ...paged('date/:year', 'routes/date/$year.tsx'),
  ...paged('date/:year/:month', 'routes/date/$year.tsx'),
  ...paged('date/:year/:month/:day', 'routes/date/$year.tsx'),
  route('avatar/:slug', 'routes/avatar/$slug.tsx'),
  route('purge/comment/:postId/server', 'routes/purge/comment.$postId.tsx'),
  route(':id.moe', 'routes/$id[.moe].tsx'),
  route('comments', 'routes/comments.tsx'),
  route('feed', 'routes/feed.tsx'),
  route(':slug', 'routes/$slug.tsx'),
  route('*', 'routes/$.tsx'),
] satisfies RouteConfig
