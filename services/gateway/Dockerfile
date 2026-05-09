FROM node:20-alpine AS build

WORKDIR /repo
RUN corepack enable
RUN apk add --no-cache git

COPY . .

# Container builds don't include .git (see .dockerignore); skip hooks/bootstrap scripts.
RUN pnpm install --frozen-lockfile --ignore-scripts
RUN pnpm -C services/gateway build
RUN pnpm --filter @papertrading/gateway deploy --prod --legacy --ignore-scripts /out/gateway

FROM node:20-alpine
WORKDIR /app

ENV NODE_ENV=production
COPY --from=build /out/gateway ./

EXPOSE 4000
CMD ["node", "--import", "./dist/otel.js", "./dist/index.js"]

