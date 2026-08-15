# Plano de Implementação — StoryCode MVP

## Objetivo do MVP

Construir uma ferramenta local-first, distribuível como binário único, que:

1. Analisa um repositório Python local.
2. Detecta uma rota FastAPI.
3. Extrai o handler e chamadas diretas relevantes.
4. Gera uma história em rascunho.
5. Permite abrir a história em browser local.
6. Exibe uma timeline visual e evidências de código.
7. Detecta quando uma evidência mudou após reindexação.
8. Não executa código do repositório analisado.

## Regras obrigatórias

- Não executar código, scripts, testes, hooks Git ou imports do repositório analisado.
- Não exigir Docker, Node.js, Python ou banco externo em runtime.
- Usar SQLite local.
- Servir a UI em `127.0.0.1` por padrão.
- Não enviar nenhum dado externamente.
- Tratar arquivos indexados como conteúdo não confiável.
- Nunca expor segredos em logs, browser ou exportações.
- Implementar uma funcionalidade por tarefa e validar antes de seguir.
- Não criar abstrações, plugins ou integrações futuras antes do vertical slice funcionar.

## Marco 0 — Bootstrap do repositório

### Tarefa 0.1 — Criar estrutura base

Criar um monorepo com:

```text
storycode/
├── cmd/storycode/
├── internal/
│   ├── cli/
│   ├── config/
│   ├── domain/
│   ├── storage/
│   ├── repository/
│   ├── indexer/
│   ├── analyzers/
│   ├── stories/
│   ├── verification/
│   ├── server/
│   └── assets/
├── migrations/
├── web/
├── fixtures/
│   └── fastapi-rag-demo/
├── docs/
├── scripts/
├── .github/workflows/
├── CHANGELOG.md
├── README.md
├── LICENSE
├── go.mod
└── Makefile
```

Critérios de aceite:

- `go test ./...` executa sem falha.
- `go vet ./...` executa sem falha.
- `go run ./cmd/storycode --help` mostra ajuda.
- `README.md`, `LICENSE` e `CHANGELOG.md` existem.
- Nenhuma dependência de runtime externo é exigida.

### Tarefa 0.2 — Definir convenções técnicas

Criar:

```text
.editorconfig
.gitattributes
.gitignore
docs/development/coding-standards.md
docs/development/definition-of-done.md
```

Definir:

- Go formatado por `gofmt`.
- TypeScript formatado por Prettier.
- Lint obrigatório em CI.
- Nomes de entidades e campos em inglês.
- Textos de interface inicialmente em inglês.
- Contratos HTTP usando `snake_case`.
- Erros estruturados e códigos estáveis.
- Testes obrigatórios para lógica de domínio e handlers HTTP.

Critérios de aceite:

- `make format` formata Go e frontend.
- `make lint` valida Go e frontend.
- O CI executa format check, lint, testes e build.

### Tarefa 0.3 — Criar fixture FastAPI

Criar `fixtures/fastapi-rag-demo/` com uma aplicação Python pequena, estática e segura para análise.

Estrutura mínima:

```text
fixtures/fastapi-rag-demo/
├── app/
│   ├── main.py
│   ├── api/
│   │   └── chat.py
│   ├── services/
│   │   ├── retrieval.py
│   │   └── generation.py
│   ├── repositories/
│   │   └── vector_store.py
│   └── models/
│       └── chat.py
├── tests/
│   └── test_chat.py
├── README.md
└── pyproject.toml
```

O fluxo da fixture deve conter:

```text
POST /v1/chat
→ create_chat()
→ RetrievalService.retrieve()
→ VectorStore.search()
→ GenerationService.generate()
→ resposta ChatResponse
```

A fixture deve possuir ao menos:

- Uma rota FastAPI com decorator `@router.post`.
- Um handler assíncrono.
- Duas chamadas diretas entre serviços.
- Um teste com nome relacionado ao endpoint.
- Um caminho alternativo simples, como ausência de contexto.

Critérios de aceite:

- O fixture não precisa ser executado pelo StoryCode.
- O parser encontra os arquivos Python.
- O fixture possui documentação explicando o fluxo esperado.
- Testes Go validam o fixture como fonte de análise.

---

## Marco 1 — CLI, configuração e armazenamento

### Tarefa 1.1 — Implementar CLI base

Implementar comandos vazios:

```bash
storycode init
storycode status
storycode index
storycode discover
storycode serve
storycode story list
storycode story show <key>
storycode verify
```

Critérios de aceite:

- Todos os comandos possuem `--help`.
- Comandos retornam códigos de saída corretos.
- Erros são legíveis e acionáveis.
- `storycode --version` retorna versão e commit de build.
- Nenhum comando altera o repositório sem operação explícita.

### Tarefa 1.2 — Implementar `storycode init`

O comando deve criar:

```text
.storycode/
├── config.yaml
├── stories/
├── index/
└── cache/
```

Configuração mínima:

```yaml
version: 1

repository:
  include:
    - "**/*.py"
    - "tests/**/*.py"
    - "docs/**/*.md"
  exclude:
    - ".git/**"
    - ".venv/**"
    - "venv/**"
    - "__pycache__/**"
    - "node_modules/**"

analysis:
  languages:
    - python
  follow_symlinks: false
  max_file_size_bytes: 5242880

storage:
  mode: repository
  engine: sqlite
```

Critérios de aceite:

- É idempotente.
- Não sobrescreve configuração existente sem `--force`.
- Funciona em caminhos Windows com espaço.
- Cria arquivos com mensagens úteis em caso de erro de permissão.

### Tarefa 1.3 — Implementar SQLite e migrações

Criar banco local:

```text
.storycode/index/storycode.db
```

Implementar migrações para tabelas mínimas:

```text
repositories
index_runs
source_files
code_symbols
code_relations
entry_points
stories
story_triggers
story_actors
story_paths
scenes
scene_transitions
evidences
evidence_references
evidence_verifications
```

Critérios de aceite:

- Banco é criado automaticamente no `init` ou primeiro `index`.
- Migrações são versionadas e transacionais.
- Erro de migração preserva banco anterior.
- Teste de integração cria banco temporário e aplica todas as migrações.
- SQLite deve funcionar sem CGO.

### Tarefa 1.4 — Implementar `storycode status`

Exibir:

```text
Repository: fastapi-rag-demo
Index status: not indexed
Stories: 0
Database: .storycode/index/storycode.db
Config: .storycode/config.yaml
```

Critérios de aceite:

- Funciona antes e depois de indexação.
- Possui saída humana e `--json`.
- Não falha se Git não estiver instalado.
- Não revela caminho completo no modo `--privacy`.

---

## Marco 2 — Analisador Python mínimo

### Tarefa 2.1 — Implementar descoberta de arquivos

Criar scanner de diretório que:

- Respeita `include` e `exclude`.
- Ignora `.git`, ambientes virtuais e diretórios configurados.
- Não segue symlinks por padrão.
- Normaliza caminhos internamente com `/`.
- Detecta arquivos maiores que o limite.
- Não carrega todo o conteúdo do repositório em memória.

Critérios de aceite:

- Testes para caminhos Windows, Linux, Unicode e espaços.
- Testes para exclusões.
- Testes para symlink, quando suportado no ambiente de teste.
- Arquivos com erro são avisos, não falhas globais.

### Tarefa 2.2 — Extrair símbolos Python

Usar Tree-sitter ou parser Python equivalente para extrair:

```text
module
class
function
method
decorator
import
```

Para cada símbolo, persistir:

```text
qualified_name
display_name
kind
source_file
start_line
end_line
semantic_hash
```

Critérios de aceite:

- Extrai `create_chat`, `RetrievalService.retrieve` e `GenerationService.generate` da fixture.
- Preserva linhas corretas.
- Erro de sintaxe em um arquivo não interrompe demais arquivos.
- Testes usam fixtures reais, não apenas mocks.

### Tarefa 2.3 — Detectar rotas FastAPI

Detectar decoradores FastAPI:

```python
@router.get(...)
@router.post(...)
@router.put(...)
@router.patch(...)
@router.delete(...)
```

Extrair:

```text
HTTP method
path
handler symbol
framework = fastapi
entry point key
```

Critérios de aceite:

- Detecta `POST /v1/chat` da fixture.
- Associa a rota ao handler `create_chat`.
- Normaliza rotas com prefixos quando eles forem explícitos no mesmo arquivo.
- Quando prefixo não puder ser resolvido, cria aviso e reduz confiança.
- Não executa código Python.

### Tarefa 2.4 — Extrair chamadas diretas

Implementar análise estática mínima de chamadas dentro de funções e métodos.

A primeira versão deve cobrir:

```python
service.retrieve(...)
vector_store.search(...)
generator.generate(...)
```

Persistir relações:

```text
from_symbol
to_symbol ou external_ref
kind = calls
line
confidence
```

Critérios de aceite:

- Detecta relações do fluxo da fixture.
- Relações não resolvidas são persistidas como `external_ref`.
- Não afirmar resolução completa quando houver ambiguidade.
- Retornar nível de confiança por relação.

### Tarefa 2.5 — Implementar `storycode index`

O comando deve executar:

```text
Criar IndexRun
→ escanear arquivos
→ persistir SourceFile
→ extrair símbolos
→ detectar FastAPI routes
→ extrair chamadas
→ persistir relações
→ finalizar IndexRun
```

Critérios de aceite:

- `storycode index` funciona sobre a fixture.
- O comando mostra progresso por fase.
- Uma segunda execução sem mudanças é incremental.
- A primeira execução pode ser repetida sem duplicar dados.
- Cancelamento não corrompe banco.
- `storycode status` mostra arquivos, símbolos, relações e entry points encontrados.

---

## Marco 3 — Descoberta e persistência de histórias

### Tarefa 3.1 — Implementar formato YAML de história

Criar schema versionado para:

```text
.storycode/stories/<story-key>.yaml
```

O parser deve validar:

```yaml
version: 1
key: post-v1-chat
title: Processar solicitação de chat
intent: Responder perguntas com evidências recuperadas.
status: draft

trigger:
  type: http
  method: POST
  path: /v1/chat

actors: []
paths: []
scenes: []
invariants: []
```

Critérios de aceite:

- Validação retorna linha/campo/motivo.
- YAML inválido não corrompe o banco.
- `storycode story show <key>` lê e apresenta o arquivo.
- O formato é determinístico ao ser regravado.

### Tarefa 3.2 — Implementar descoberta por entry point

Implementar:

```bash
storycode discover --entry POST:/v1/chat
```

Algoritmo inicial:

```text
EntryPoint
→ handler
→ chamadas diretas resolvidas
→ profundidade máxima configurável
→ cria atores por componente/tipo
→ cria cenas na ordem encontrada
→ cria relações sequenciais
→ associa evidências de símbolos
→ salva história como draft
```

Regra do MVP: seguir somente chamadas internas resolvidas e limitar profundidade padrão a `5`.

Critérios de aceite:

- Gera uma história `post-v1-chat`.
- Possui trigger HTTP.
- Possui ator usuário e ator Chat API.
- Possui cenas para handler, retrieval e generation.
- Cada cena possui ao menos uma evidência de símbolo.
- A história é salva em YAML e persistida no SQLite.
- Reexecutar não sobrescreve edição manual sem `--force`.

### Tarefa 3.3 — Implementar sincronização YAML ↔ SQLite

Definir o SQLite como índice de consulta e os arquivos YAML como fonte versionável das histórias.

Criar:

```bash
storycode story sync
```

Comportamento:

```text
YAML alterado
→ validar
→ atualizar SQLite

SQLite alterado via API/UI
→ regravar YAML deterministicamente
→ atualizar versão
```

Critérios de aceite:

- Mudança em YAML aparece na UI após sync.
- Edição via API atualiza o YAML.
- Conflitos de versão retornam erro claro.
- Arquivo YAML nunca é parcialmente gravado.
- Alteração manual inválida mantém último estado válido no SQLite e informa erro.

### Tarefa 3.4 — Implementar comandos de histórias

Implementar:

```bash
storycode story list
storycode story show <key>
storycode story create <key>
storycode tell <key>
```

Critérios de aceite:

- `list` mostra chave, título, status e gatilho.
- `show` exibe estrutura completa.
- `tell` gera narrativa textual ordenada.
- Todos suportam `--json`.
- O comando `tell` mostra evidências por cena.

---

## Marco 4 — Verificação e drift

### Tarefa 4.1 — Criar hashes semânticos

Para cada símbolo e evidência, armazenar:

```text
file_content_hash
symbol_semantic_hash
snapshot_hash
last_seen_index_run_id
```

O hash semântico deve ser mais estável do que número de linha, sempre que possível.

Critérios de aceite:

- Alterar linhas fora de um símbolo não invalida automaticamente a evidência do símbolo.
- Alterar corpo ou assinatura do símbolo marca a evidência como `changed`.
- Remover símbolo marca a evidência como `missing`.
- Renomear arquivo pode ser reportado como `ambiguous` inicialmente.

### Tarefa 4.2 — Implementar `storycode verify`

Implementar:

```bash
storycode verify
storycode verify post-v1-chat
```

O processo deve:

```text
Reindexar somente o necessário
→ localizar evidências
→ comparar hashes
→ persistir EvidenceVerification
→ recalcular Scene.status
→ recalcular Story.status
→ produzir relatório
```

Critérios de aceite:

- História recém-descoberta pode ser marcada como `verified` após validação.
- Alterar `RetrievalService.retrieve` na fixture marca a cena correspondente como `stale`.
- Remover `GenerationService.generate` marca história como `broken`.
- Saída informa evidência, arquivo, cena e ação recomendada.
- `--format json` retorna estrutura consumível por automação.

### Tarefa 4.3 — Exibir relatório de drift

Implementar:

```bash
storycode verify post-v1-chat --format markdown
```

O relatório deve incluir:

```text
História
Status anterior e atual
Cenas afetadas
Evidências alteradas/removidas
Resumo de impacto
Próxima ação recomendada
```

Critérios de aceite:

- Markdown é legível e determinístico.
- JSON segue contrato estável.
- Não inclui segredos ou conteúdo completo de arquivos.

---

## Marco 5 — API HTTP local

### Tarefa 5.1 — Criar servidor HTTP

Implementar:

```bash
storycode serve
```

Requisitos:

```text
Host padrão: 127.0.0.1
Porta padrão: 7331
Fallback para porta livre opcional
Assets web incorporados no binário
Encerramento gracioso
```

Critérios de aceite:

- `GET /health` retorna 200.
- API inicia sem Node.js em runtime.
- Servidor não expõe `0.0.0.0` por padrão.
- Ctrl+C encerra servidor sem corromper banco.
- URL local é exibida no terminal.

### Tarefa 5.2 — Implementar endpoints de leitura

Implementar primeiro:

```text
GET /api/v1/repository
GET /api/v1/stories
GET /api/v1/stories/{story_key}
GET /api/v1/stories/{story_key}/player
GET /api/v1/stories/{story_key}/scenes/{scene_key}
GET /api/v1/stories/{story_key}/scenes/{scene_key}/evidences
GET /api/v1/evidences/{evidence_id}
GET /api/v1/files/content
GET /api/v1/entry-points
GET /api/v1/search
```

Critérios de aceite:

- Respostas seguem contratos documentados.
- Erros usam envelope padronizado.
- Arquivos não podem ser lidos fora do repositório.
- Conteúdo é sanitizado e segredos são mascarados.
- Testes HTTP cobrem sucesso, 404, validação e path traversal.

### Tarefa 5.3 — Implementar endpoint de verificação

Implementar:

```text
POST /api/v1/verifications
GET /api/v1/verifications/{id}
GET /api/v1/stories/{story_key}/drift
```

Critérios de aceite:

- O frontend consegue iniciar e consultar uma verificação.
- A operação não bloqueia a API.
- Somente uma indexação/verificação mutável ocorre por repositório no MVP.
- Operações concorrentes retornam `409 Conflict`.
- Resultados atualizam status das histórias.

### Tarefa 5.4 — Implementar SSE

Implementar:

```text
GET /api/v1/events
GET /api/v1/index-runs/{id}/events
GET /api/v1/verifications/{id}/events
```

Critérios de aceite:

- A UI recebe progresso da indexação.
- A UI recebe atualização de status de história.
- Reconexão SSE não quebra servidor.
- Eventos não incluem conteúdo de código ou segredos.

---

## Marco 6 — Frontend Story Player

### Tarefa 6.1 — Criar frontend embutido

Criar `web/` com:

```text
React
TypeScript
Vite
Tailwind CSS
Zustand
React Router
Shiki
Motion
```

O build do frontend deve gerar assets incorporados pelo binário Go.

Critérios de aceite:

- `make web-build` gera `web_dist/`.
- `go build ./cmd/storycode` incorpora os assets.
- `storycode serve` funciona sem `npm` ou Node.js instalado na máquina usuária.
- A UI é aberta em `http://127.0.0.1:7331`.

### Tarefa 6.2 — Criar tela inicial

A home deve mostrar:

```text
Nome do repositório
Status do índice
Contagem de histórias
Histórias verificadas, stale e broken
Lista de histórias
Busca por título, gatilho e tag
Botão para indexar/verificar
```

Critérios de aceite:

- Carrega via `GET /api/v1/repository` e `GET /api/v1/stories`.
- Mostra estados de carregamento, vazio e erro.
- Funciona por teclado.
- Não depende exclusivamente de cor para status.

### Tarefa 6.3 — Criar Story Player em timeline

Criar rota:

```text
/stories/:storyKey
```

A tela deve exibir:

```text
Título
Intenção
Gatilho
Atores
Timeline de cenas
Cena selecionada
Controles anterior/próxima/reiniciar
Status de verificação
Invariantes
Atalho para evidências
```

Critérios de aceite:

- Usa `GET /api/v1/stories/{story_key}/player`.
- Permite avançar e voltar por teclado.
- Cena atual destaca atores de origem e destino.
- Animações obedecem `prefers-reduced-motion`.
- Caminho principal funciona para a história da fixture.
- Não é necessário auto-play no MVP.

### Tarefa 6.4 — Criar painel de evidências

Ao clicar em uma cena, permitir abrir drawer lateral ou modal com:

```text
Título da evidência
Tipo
Status
Arquivo
Símbolo
Linhas
Trecho de código com highlight
Teste associado, quando disponível
```

Critérios de aceite:

- Dados vêm do endpoint de evidências.
- Código é renderizado como texto seguro.
- Conteúdo longo possui scroll.
- Linhas são exibidas.
- Evidência `changed`, `missing` ou `ambiguous` possui destaque acessível.

### Tarefa 6.5 — Criar visualização de drift

Na página da história, apresentar:

```text
Badge de status
Cenas afetadas
Evidências alteradas/removidas
Botão “Verificar novamente”
Mensagem de próxima ação
```

Critérios de aceite:

- Atualiza após término de verificação por SSE ou polling.
- A história da fixture fica `stale` após mudança controlada.
- A tela explica o motivo do status, não apenas a cor.

---

## Marco 7 — Segurança, portabilidade e release

### Tarefa 7.1 — Implementar proteção de API local

Implementar:

```text
Loopback only
CORS restrito
Token efêmero de sessão
Cookie HttpOnly/SameSite ou header local
Proteção de endpoints mutáveis
Headers de segurança
```

Critérios de aceite:

- Escritas sem token retornam 401 ou 403.
- Origin externo é rejeitado.
- Server não responde em interface de rede externa por padrão.
- Token não é salvo em disco.

### Tarefa 7.2 — Implementar sanitização e redaction

Criar regras para:

```text
.env
.env.*
*.pem
*.key
id_rsa
credentials.json
secrets.*
```

Implementar mascaramento básico de padrões comuns:

```text
API keys
Bearer tokens
private keys
password assignments
connection strings
```

Critérios de aceite:

- Arquivos sensíveis são ignorados por padrão.
- Preview não exibe valores secretos conhecidos.
- Logs não exibem conteúdo de arquivos.
- Testes incluem fixture com segredo falso e validam mascaramento.

### Tarefa 7.3 — Cross-platform

Validar em:

```text
Windows 10/11 amd64
Linux amd64
```

Critérios de aceite:

- Caminhos com espaço funcionam.
- `.exe` funciona em PowerShell e CMD.
- Banco SQLite abre e fecha corretamente.
- Build não requer CGO.
- `storycode serve` abre browser de maneira opcional e compatível.
- Documentação fornece exemplos de PowerShell e Bash.

### Tarefa 7.4 — CI e releases

Criar GitHub Actions para:

```text
Pull request:
  gofmt check
  go vet
  go test
  frontend lint
  frontend test
  frontend build

Release:
  build Windows amd64
  build Linux amd64
  gerar SHA-256
  publicar artefatos
  criar release notes a partir do CHANGELOG
```

Critérios de aceite:

- Artefatos possuem versão e checksums.
- `storycode --version` inclui versão de release.
- `CHANGELOG.md` é atualizado em cada release.
- O projeto possui instruções de instalação e upgrade.

---

## Ordem de validação

Após cada marco, executar:

```bash
make format
make lint
make test
make build
```

Após os marcos 2, 3, 4 e 6, executar também:

```bash
./storycode init
./storycode index
./storycode discover --entry POST:/v1/chat
./storycode tell post-v1-chat
./storycode verify post-v1-chat
./storycode serve
```

## Demo obrigatória do MVP

A demonstração final deve funcionar assim:

```bash
git clone <storycode>
cd storycode

make build

cd fixtures/fastapi-rag-demo
../../bin/storycode init
../../bin/storycode index
../../bin/storycode discover --entry POST:/v1/chat
../../bin/storycode verify post-v1-chat
../../bin/storycode serve
```

No browser:

1. Abrir a história `Processar solicitação de chat`.
2. Ver o gatilho `POST /v1/chat`.
3. Avançar por handler, retrieval, vector store e generation.
4. Abrir a evidência de `RetrievalService.retrieve`.
5. Alterar o método `retrieve` no fixture.
6. Rodar indexação/verificação.
7. Ver a história marcada como `stale`.
8. Abrir o painel e entender qual evidência mudou.

## Fora do escopo deste plano

Não implementar antes do MVP:

- Suporte além de Python/FastAPI.
- LLM, embeddings ou RAG.
- Integração de GitHub/GitLab/Jira/Notion.
- Runtime traces e OpenTelemetry.
- Multi-repositório.
- Multiusuário ou sincronização cloud.
- Editor visual completo de grafo.
- Geração de vídeo ou GIF.
- Plugins externos.
- Aplicativo desktop Tauri/Electron.
- Neo4j, Qdrant, Kafka, Redis ou infraestrutura remota.