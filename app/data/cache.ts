import { createClient } from 'redis'
import * as msgpack from 'msgpack-lite'

async function connectRedis() {
  const client = createClient({
    url: process.env.REDIS_URL,
  })
  await client.connect()
  globalThis.blogRedisClient = client
  return client
}

function getRedis(): Promise<ReturnType<typeof createClient>> {
  if (!globalThis.blogRedisClient) {
    globalThis.blogRedisClient = connectRedis()
  }
  return Promise.resolve(globalThis.blogRedisClient)
}

function prefix(k: string) {
  return `i3w:${k}`
}

export async function set<T = any>(key: string, value: T): Promise<void> {
  const redis = await getRedis()
  await redis.setEx(prefix(key), 86400, msgpack.encode(value) as any)
}

export async function get<T = any>(key: string): Promise<T | null> {
  const redis = await getRedis()
  const value = await redis.getBuffer(prefix(key))
  if (!value) {
    return null
  }
  return msgpack.decode(value)
}

export async function del(key: string): Promise<void> {
  const redis = await getRedis()
  await redis.del(prefix(key))
}

export async function clear(group?: string): Promise<void> {
  const redis = await getRedis()
  const keys = await redis.keys(prefix(group ? `${group}:*` : '*'))
  if (keys.length) await redis.del(keys)
}
