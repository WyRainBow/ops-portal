FROM node:20-alpine AS build
WORKDIR /src

COPY apps/web/package.json apps/web/package-lock.json* apps/web/pnpm-lock.yaml* apps/web/yarn.lock* ./apps/web/
RUN cd apps/web && (npm ci || npm install)

COPY apps/web ./apps/web
RUN cd apps/web && npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production

COPY --from=build /src/apps/web/.next/standalone /app
COPY --from=build /src/apps/web/.next/static /app/.next/static
COPY --from=build /src/apps/web/public /app/public

ENV PORT=18082
EXPOSE 18082
CMD ["node", "server.js"]
