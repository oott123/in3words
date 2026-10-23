FROM node:16 AS builder
RUN npm install -g pnpm@8.15.9
WORKDIR /app
COPY package.json pnpm-lock.yaml setup.js /app/
RUN pnpm install --frozen-lockfile
COPY . /app/
RUN pnpm build

FROM node:16-slim AS runner
RUN npm install -g pnpm@8.15.9
WORKDIR /app
COPY package.json pnpm-lock.yaml setup.js /app/
ENV NODE_ENV production
RUN DISABLE_POST_INSTALL=yes pnpm install --frozen-lockfile --prod && rm -rf "$(pnpm store path)"
COPY --from=builder /app/build /app/build/
COPY --from=builder /app/public /app/public/
CMD node node_modules/.bin/remix-serve build
