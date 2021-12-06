/* eslint-disable no-var */

declare var blogApiGetCache: Map<string, any> | undefined

// polyfill pre 4.5
type Awaited<T> = T extends PromiseLike<infer U> ? U : T
