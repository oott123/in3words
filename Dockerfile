FROM node:16 AS builder
WORKDIR /app
COPY package.json yarn.lock .yarnrc.yml /app/
COPY .yarn /app/.yarn/
RUN yarn && yarn cache clean
COPY . /app/
RUN yarn build

FROM node:16-slim
WORKDIR /app
COPY package.json yarn.lock .yarnrc.yml /app/
COPY .yarn /app/.yarn/
RUN yarn --production --ignore-scripts --cache-folder /tmp/yarn-cache && \
  yarn cache clean && \
  rm -rf /tmp/yarn-cache
COPY --from=builder /app/build /app/build/
COPY --from=builder /app/public /app/public/
CMD node node_modules/.bin/remix-serve build
