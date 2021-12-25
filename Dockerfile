FROM node:16 AS builder
WORKDIR /app
COPY package.json yarn.lock /app/
RUN yarn
COPY . /app/
RUN yarn build
RUN yarn --production --ignore-scripts

FROM node:16-slim
WORKDIR /app
COPY --from=builder /app /app
CMD node node_modules/.bin/remix-serve build
