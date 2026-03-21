# ─────────────────────────────────────────────
# Stage 1: Build — transpila o TypeScript com esbuild
# ─────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Instala esbuild globalmente para transpilar o TS do k6
RUN npm install -g esbuild

COPY k6-test.ts .

# Transpila para JS compatível com o runtime do k6 (ES2015, sem node built-ins)
RUN esbuild k6-test.ts \
    --bundle \
    --outfile=k6-test.js \
    --format=esm \
    --target=es2015 \
    --platform=neutral \
    --external:k6 \
    --external:k6/*

# ─────────────────────────────────────────────
# Stage 2: Run — executa o teste com k6
# ─────────────────────────────────────────────
FROM grafana/k6:latest

WORKDIR /tests

# Copia o script transpilado
COPY --from=builder /app/k6-test.js .

# O container go_backend deve estar acessível pela rede Docker.
# Execute com:
#   docker run --rm --network <sua_rede> k6-load-test run /tests/k6-test.js
#
# Ou sobrescreva o host via variável de ambiente:
#   docker run --rm --network <sua_rede> \
#     -e BASE_URL=http://go_backend:8080 \
#     k6-load-test run /tests/k6-test.js

ENTRYPOINT ["k6"]
CMD ["run", "/tests/k6-test.js"]