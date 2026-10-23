FROM node:24 AS builder
RUN npm install -g pnpm@12.15.0
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
RUN pnpm install --frozen-lockfile
COPY . /app/
RUN pnpm build

FROM node:24-slim AS runner
RUN npm install -g pnpm@12.15.0
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml /app/
ENV NODE_ENV=production
RUN pnpm install --frozen-lockfile --prod && rm -rf "$(pnpm store path)"
COPY --from=builder /app/build /app/build/
COPY --from=builder /app/public /app/public/
CMD ["node_modules/.bin/react-router-serve", "build/server/index.js"]
