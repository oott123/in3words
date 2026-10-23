import { clear as clearCache } from '~/data/cache'
import { commentCacheGroup } from '~/data/comments'
import * as yup from 'yup'
import type { Route } from './+types/comment.$postId'

export async function action({ params }: Route.ActionArgs) {
  const id = await yup
    .number()
    .positive()
    .integer()
    .required()
    .validate(params.postId)

  await clearCache(commentCacheGroup(id))

  return {
    ok: true,
  }
}
