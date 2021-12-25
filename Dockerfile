FROM node:16 AS builder
WORKDIR /app
COPY package.json yarn.lock .yarnrc.yml /app/
COPY .yarn /app/.yarn/
RUN yarn && yarn cache clean --all
COPY . /app/
RUN yarn build

FROM node:16-slim
WORKDIR /app
COPY package.json yarn.lock .yarnrc.yml /app/
COPY .yarn /app/.yarn/
ENV NODE_ENV production
RUN DISABLE_POST_INSTALL=yes yarn workspaces focus --all --production && yarn cache clean --all
COPY --from=builder /app/build /app/build/
COPY --from=builder /app/public /app/public/
CMD node node_modules/.bin/remix-serve build
