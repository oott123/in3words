import { reactRouter } from '@react-router/dev/vite'
import { defineConfig } from 'vite'

export default defineConfig(({ command }) => ({
  // 构建产物放在 /_i3w_remix/ 下，避免和 WordPress 的路径冲突；
  // 开发服务器设置 base 后会把整个应用挂到该路径下，所以只在构建时设置
  base: command === 'build' ? '/_i3w_remix/' : '/',
  plugins: [reactRouter()],
  resolve: {
    tsconfigPaths: true,
  },
  server: {
    port: 3000,
  },
}))
