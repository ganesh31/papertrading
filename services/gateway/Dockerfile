FROM node:20-alpine AS build

WORKDIR /repo
RUN corepack enable

COPY . .

RUN pnpm install --frozen-lockfile
RUN pnpm -C services/gateway build
RUN pnpm --filter @papertrading/gateway deploy --prod --legacy /out/gateway

FROM node:20-alpine
WORKDIR /app

ENV NODE_ENV=production
COPY --from=build /out/gateway ./

EXPOSE 4000
CMD ["node", "--import", "./dist/otel.js", "./dist/index.js"]

