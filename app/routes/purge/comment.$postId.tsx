import { ActionFunction } from 'remix'
import { clear as clearCache } from '~/data/cache'
import { commentCacheGroup } from '~/data/comments'
import * as yup from 'yup'

export const action: ActionFunction = async ({ params }) => {
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
