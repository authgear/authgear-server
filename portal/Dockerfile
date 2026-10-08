# syntax=docker/dockerfile:1

FROM node:24.16.0-bookworm AS build
WORKDIR /usr/src/app
COPY package.json package-lock.json .npmrc ./
# file: dependency. npm ci fails if this tree is missing.
COPY vendor/storybook-addon-comments ./vendor/storybook-addon-comments
COPY scripts/. ./scripts/.
RUN npm ci
COPY . .
RUN npm run build-storybook

FROM nginx:1.27-alpine
COPY --from=build /usr/src/app/storybook-static /usr/share/nginx/html
COPY default.conf.template /etc/nginx/templates/default.conf.template
