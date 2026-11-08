# Última Pista

Jogo web de enigmas cooperativos em português brasileiro. Uma pessoa cria a sala e vira **narrador**; as demais entram como **detetives** por link ou código. Um navegador separado pode abrir o **telão**. Sem conta.

## Stack

- Frontend: React + TypeScript + Vite
- Backend: Go (HTTP + SSE)
- Persistência: SQLite em disco (`data/ultima-pista.db` por padrão)
- Mesma origem em produção (Go serve `web/dist` + `/api`)

## Requisitos

- Go 1.22+ (toolchain automática se necessário)
- Node.js 20+

## Instalação

```bash
# Dependências do frontend
cd web && npm install && cd ..

# Dependências Go
go mod download
```

## Desenvolvimento

Terminal 1 — API:

```bash
go run ./cmd/server -addr :8080 -db data/ultima-pista.db
```

Terminal 2 — Vite (proxy `/api` → `:8080`):

```bash
cd web && npm run dev
```

Abra o endereço do Vite (geralmente `http://127.0.0.1:5173`).

## Produção (mesma origem)

```bash
cd web && npm run build && cd ..
go run ./cmd/server -addr :8080 -db data/ultima-pista.db -static web/dist
```

Variáveis de ambiente opcionais:

| Variável | Padrão | Descrição |
|---|---|---|
| `ADDR` | `:8080` | Endereço HTTP |
| `DB_PATH` | `data/ultima-pista.db` | Caminho do SQLite |
| `STATIC_DIR` | `web/dist` | Frontend compilado |
| `SECURE_COOKIES` | vazio | Defina `1` para cookies `Secure` (HTTPS) |

Salas expiram após **24 horas** sem atividade.

## Simular narrador, detetive e telão

1. **Narrador:** abra o app em um perfil/janela anônima → **Criar sala** (escolha Presencial ou À distância).
2. **Detetive:** copie o convite (`/entrar/CODIGO`) e abra em outro navegador/perfil → informe um apelido.
3. **Telão:** no lobby, **Abrir telão** (ou `/telao/CODIGO`) em um terceiro contexto. O telão é somente leitura mesmo se o cookie do narrador estiver presente.
4. Narrador escolhe um caso → detetive pergunta → narrador responde → teoria → revelação. No presencial, use **Registrar pergunta oral** e **Confirmar solução**.

## Scripts

```bash
# Backend
go test ./...
go build -o bin/ultima-pista ./cmd/server

# Frontend
cd web
npm run lint
npm run build
```

## Papéis e sigilo

- Cookie de sessão HttpOnly (`up_session`), SameSite=Lax; `Secure` em produção via `SECURE_COOKIES=1`.
- Solução e notas do caso só em `GET /api/rooms/:code/private` (narrador).
- Snapshots públicos, SSE e telão nunca incluem a solução antes do encerramento.
- SSE emite apenas `invalidate` + `version`; o cliente recarrega o snapshot HTTP.
