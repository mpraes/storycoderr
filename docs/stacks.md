Aqui estão três stacks viáveis para o StoryCode. As três mantêm o princípio de binário local, baixo atrito para Windows/Linux e UI visual no browser; a diferença é onde você aceita mais complexidade ou ganha mais velocidade.

## Alternativas de stack

| Stack | Melhor para | Trade-off principal |
|---|---|---|
| **A. Go + React embutido** | Produto CLI-first robusto e distribuível em binário único | Build frontend/backend mais estruturado |
| **B. Go + HTMX/templ** | MVP muito leve, poucas dependências e execução simples | Menor liberdade para animações/edição visual sofisticada |
| **C. Tauri + Rust + React** | Experiência desktop premium com UI rica e acesso local seguro | Maior complexidade de distribuição e curva Rust |

***

## A. Go + React + SQLite

**Recomendada para o projeto.** É a melhor combinação entre distribuição simples, análise de código, CLI forte e interface de storytelling visual.

```text
┌──────────────────────────────────────────────────────┐
│                 storycode (binário Go)               │
│                                                      │
│  CLI · Indexador · Tree-sitter · SQLite · API REST   │
│  SSE · Exportadores · Assets web incorporados        │
│                                                      │
│                    go:embed                          │
│                      │                               │
│                      ▼                               │
│             React SPA compilada                       │
│       Timeline · Canvas · Evidence Viewer             │
└──────────────────────────────────────────────────────┘
```

### Componentes

| Área | Tecnologia |
|---|---|
| Linguagem principal | Go |
| CLI | Cobra ou Kong |
| HTTP API | `net/http` + `chi` |
| Banco local | SQLite |
| Driver SQLite | `modernc.org/sqlite` para evitar CGO |
| Migrações | Goose ou Atlas |
| Análise AST | Tree-sitter |
| Git | `go-git` ou execução segura/read-only de Git local |
| UI | React + TypeScript + Vite |
| Grafo/mapa | React Flow |
| Timeline storytelling | SVG + React |
| Layout automático | ELK.js |
| Animações | Motion |
| Highlight de código | Shiki |
| Estado no frontend | Zustand |
| Estilos | Tailwind CSS + shadcn/ui |
| Comunicação em tempo real | SSE |
| Empacotamento | `go:embed` |
| Release | GoReleaser + GitHub Actions |

### Estrutura sugerida

```text
storycode/
├── cmd/
│   └── storycode/
│       └── main.go
├── internal/
│   ├── app/
│   ├── cli/
│   ├── config/
│   ├── domain/
│   ├── repository/
│   ├── indexer/
│   ├── analyzers/
│   │   ├── python/
│   │   └── treesitter/
│   ├── stories/
│   ├── evidence/
│   ├── verification/
│   ├── impact/
│   ├── export/
│   ├── git/
│   ├── server/
│   └── storage/
├── migrations/
├── web/
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── web_dist/
├── docs/
├── examples/
├── CHANGELOG.md
├── go.mod
└── Makefile
```

### Como distribuir

```bash
# Windows
storycode_0.1.0_windows_amd64.zip
  └── storycode.exe

# Linux
storycode_0.1.0_linux_amd64.tar.gz
  └── storycode
```

O React é compilado no pipeline de release e incorporado no binário:

```go
//go:embed all:web_dist
var webAssets embed.FS
```

A pessoa usuária baixa um único binário e executa:

```bash
storycode init
storycode index
storycode serve
```

### Vantagens

- Binário único e sem Node/Python/Docker em runtime.
- Excelente para CLI e SQLite.
- Bom suporte a concorrência para indexação.
- Dependências operacionais mínimas.
- Funciona muito bem em Windows e Linux.
- A UI React permite timeline, pan/zoom, SVG, animações e painel de evidências.
- Você pode manter a UI evoluindo sem transformar o core em um app desktop pesado.

### Riscos

- Tree-sitter e grammars exigem atenção no build cross-platform.
- React Flow + ELK pode aumentar bastante o tamanho do asset web.
- Ter Go e TypeScript exige pipeline de build com duas etapas.

### Escolhas concretas

```text
Go:                1.23+
HTTP:              chi
CLI:               cobra
SQLite:            modernc.org/sqlite
Migrations:        goose
Config:            koanf ou viper
Logging:           slog
Tree-sitter:       go-tree-sitter
Frontend:          React 19 + TypeScript + Vite
UI:                Tailwind + shadcn/ui
Graph:             React Flow
Layout:             ELK.js
Animation:         Motion
Code highlighting:  Shiki
State:             Zustand
E2E:               Playwright
Go tests:          testing + testify opcional
Release:           GoReleaser
```

***

## B. Go + templ + HTMX + SQLite

Escolha esta stack se a prioridade absoluta for colocar uma versão funcional na mão das pessoas rapidamente, com binário pequeno e quase nenhum tooling web complexo.

```text
┌───────────────────────────────────────────────────┐
│                storycode (binário Go)             │
│                                                   │
│ CLI · Indexador · SQLite · Server                 │
│                                                   │
│ HTML renderizado no servidor                       │
│ templ + HTMX + Alpine.js pontual + SVG            │
└───────────────────────────────────────────────────┘
```

### Componentes

| Área | Tecnologia |
|---|---|
| Linguagem | Go |
| CLI | Cobra |
| HTTP API e páginas | `net/http` + `chi` |
| Templates tipados | templ |
| Interatividade | HTMX |
| Estado visual local | Alpine.js, apenas quando necessário |
| Diagramas | SVG gerado no backend ou browser |
| Layout de grafos | Dagre compilado/embutido ou layout simples no Go |
| Banco local | SQLite |
| Análise | Tree-sitter |
| Estilos | CSS próprio, Pico CSS ou Tailwind compilado |
| Atualizações | SSE |
| Empacotamento | `go:embed` |
| Releases | GoReleaser |

### Estrutura sugerida

```text
storycode/
├── cmd/storycode/
├── internal/
│   ├── cli/
│   ├── indexer/
│   ├── analyzers/
│   ├── domain/
│   ├── storage/
│   ├── stories/
│   ├── server/
│   └── views/
├── web/
│   ├── templates/
│   │   ├── layout.templ
│   │   ├── stories.templ
│   │   ├── player.templ
│   │   └── evidence.templ
│   ├── static/
│   │   ├── app.css
│   │   ├── htmx.min.js
│   │   └── alpine.min.js
│   └── generated/
├── migrations/
├── docs/
└── CHANGELOG.md
```

### Vantagens

- Um único ecossistema principal: Go.
- Menos tempo configurando bundler, roteamento SPA, state management e integração API.
- HTML inicial muito rápido e simples.
- Baixo consumo de RAM e assets pequenos.
- Muito fácil incorporar tudo no binário.
- Excelente para CRUD de histórias, evidências, status, busca e relatórios.
- A interface pode funcionar bem mesmo com JavaScript limitado.

### Limitações

- A timeline cinematográfica exigirá JavaScript customizado.
- Pan/zoom, minimap, seleção múltipla e auto-layout sofisticado demandam mais trabalho.
- Editor visual de grafos será mais difícil que com React Flow.
- Se o produto migrar para experiência altamente visual, a UI pode precisar ser reescrita.

### Quando escolher

Escolha esta alternativa se você quer validar em uma ou duas semanas:

```text
Indexar FastAPI
→ descobrir rota
→ gerar história
→ abrir player HTML/SVG
→ clicar em cena
→ abrir evidência
→ detectar drift
```

Ela é ideal para provar a tese de produto antes de investir em uma SPA rica.

***

## C. Rust + Tauri + React + SQLite

Escolha esta stack se você quer que StoryCode seja, desde o início, um aplicativo desktop completo e altamente polido — mais parecido com um “IDE narrativo” do que com uma CLI que abre um browser.

```text
┌─────────────────────────────────────────────────────────┐
│                  StoryCode Desktop                       │
│                                                         │
│ Tauri shell                                              │
│   ├── React UI: timeline, mapa, animações                │
│   ├── Rust core: análise, SQLite, Git, indexação         │
│   ├── IPC commands/eventos                               │
│   └── Integração opcional com editor e filesystem        │
└─────────────────────────────────────────────────────────┘
```

### Componentes

| Área | Tecnologia |
|---|---|
| Core | Rust |
| Desktop shell | Tauri 2 |
| UI | React + TypeScript + Vite |
| Banco local | SQLite via SQLx ou rusqlite |
| Migrações | SQLx migrations |
| Análise AST | Tree-sitter Rust bindings |
| CLI | Clap |
| Git | Git CLI read-only ou `git2` |
| Comunicação UI/core | Tauri commands + events |
| Grafo | React Flow |
| Layout | ELK.js |
| Animação | Motion |
| Estilos | Tailwind CSS + shadcn/ui |
| Code highlighting | Shiki |
| Testes Rust | `cargo test` |
| Testes web | Vitest + Playwright |
| Releases | GitHub Actions + Tauri bundler |

### Estrutura sugerida

```text
storycode/
├── src-tauri/
│   ├── src/
│   │   ├── main.rs
│   │   ├── commands/
│   │   ├── domain/
│   │   ├── indexer/
│   │   ├── analyzers/
│   │   ├── storage/
│   │   ├── stories/
│   │   └── integrations/
│   ├── migrations/
│   ├── Cargo.toml
│   └── tauri.conf.json
├── web/
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── docs/
├── examples/
└── CHANGELOG.md
```

### Vantagens

- UI altamente rica, com integração nativa ao sistema operacional.
- Pode abrir arquivos no editor, arrastar projetos, gerenciar múltiplos repositórios e manter workspace local.
- Segurança forte por padrão no modelo Tauri.
- Menor consumo de memória que Electron em muitos casos.
- Excelente base para funcionalidades futuras de desktop: diff visual, watch mode, múltiplas janelas e integração com IDE.

### Riscos

- O usuário instala um aplicativo, não apenas um binário CLI.
- Empacotamento Windows pode envolver WebView2; embora comum no Windows moderno, ainda é mais uma variável.
- Curva de Rust e Tauri é maior.
- Mais difícil manter foco no MVP devido às possibilidades da plataforma.
- A CLI ainda precisa existir como interface separada ou modo complementar.

### Quando escolher

Escolha se a visão de longo prazo for:

```text
StoryCode como um aplicativo desktop
para onboarding, arquitetura, análise de impacto
e exploração visual profunda de codebases.
```

***

## Recomendação

Começaria com a **Stack A: Go + React + SQLite**, em um monorepo simples.

Ela preserva exatamente o que você pediu:

```text
Uma CLI leve
→ roda em Windows/Linux
→ não exige Docker, Node ou Python do usuário
→ inicia um servidor local
→ abre uma UI visual rica no browser
→ suporta storytelling com timeline, desenho, animação e evidências
```

A Stack B é uma forma excelente de fazer um spike de UX em poucos dias, mas eu só a escolheria se você quiser conscientemente validar o produto antes de investir numa experiência visual muito rica.

A Stack C tem o maior potencial de produto desktop, mas é prematura para a primeira entrega: primeiro precisamos provar que uma história visual gerada a partir de código realmente faz alguém entender um sistema mais rápido.