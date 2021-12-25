FROM node:16 AS builder
WORKDIR /app
COPY package.json yarn.lock /app/
RUN yarn
COPY . /app/
RUN yarn build

FROM node:16-slim
WORKDIR /app
COPY package.json yarn.lock /app/
RUN yarn --production --ignore-scripts
COPY --from=builder /app/build /app/public /app/
CMD node node_modules/.bin/remix-serve build
