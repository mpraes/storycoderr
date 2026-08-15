# API HTTP Local — StoryCode

## 1. Visão geral

A API HTTP do StoryCode expõe os dados indexados do repositório e as operações necessárias para a interface web local.

A API deve ser usada principalmente pelo Studio embutido no binário. Ela também pode ser utilizada por ferramentas locais, plugins de IDE e automações no futuro.

Base URL padrão:

```text
http://127.0.0.1:7331/api/v1
```

O servidor deve escutar somente em loopback por padrão.

---

## 2. Convenções

### 2.1 Formato de dados

Todos os corpos de requisição e resposta devem usar JSON UTF-8.

```http
Content-Type: application/json; charset=utf-8
Accept: application/json
```

Campos devem usar `snake_case`.

Datas devem usar ISO 8601 em UTC.

```json
{
  "created_at": "2026-08-15T18:11:00Z"
}
```

### 2.2 Identificadores

A API deve utilizar:

- `id`: ULID interno.
- `key`: identificador legível e estável, como `answer-with-rag`.
- `ref`: referência técnica, como `src/api/chat.py::create_chat`.

Endpoints externos devem preferir `key` para histórias e `id` para entidades internas.

### 2.3 Envelope de erro

Todos os erros devem seguir o formato:

```json
{
  "error": {
    "code": "story_not_found",
    "message": "A história 'answer-with-rag' não foi encontrada.",
    "details": {
      "story_key": "answer-with-rag"
    },
    "request_id": "01J5H9W4PGJ3P4Y1YJPK7J8HAA"
  }
}
```

Campos:

| Campo | Tipo | Descrição |
|---|---|---|
| `code` | string | Código estável para automação |
| `message` | string | Mensagem legível |
| `details` | object | Dados específicos do erro |
| `request_id` | string | Identificador para diagnóstico |

### 2.4 Códigos HTTP

| Código | Uso |
|---:|---|
| `200` | Consulta ou atualização bem-sucedida |
| `201` | Recurso criado |
| `202` | Operação assíncrona iniciada |
| `204` | Operação concluída sem corpo |
| `400` | Payload ou parâmetro inválido |
| `404` | Recurso não encontrado |
| `409` | Conflito de versão, estado ou operação concorrente |
| `422` | Erro de validação semântica |
| `429` | Limite local de operação excedido |
| `500` | Erro interno |
| `503` | Índice, migração ou serviço ainda indisponível |

### 2.5 Paginação

Listagens devem aceitar:

```text
limit
cursor
sort
order
```

Exemplo:

```http
GET /stories?limit=20&sort=updated_at&order=desc
```

Resposta:

```json
{
  "data": [],
  "page": {
    "limit": 20,
    "next_cursor": "eyJvZmZzZXQiOjIwfQ",
    "total": 42
  }
}
```

O `total` pode ser omitido em operações caras, usando `null`.

### 2.6 Concorrência e versão

Recursos editáveis devem expor:

```json
{
  "version": 4,
  "updated_at": "2026-08-15T18:11:00Z"
}
```

Atualizações devem receber uma das opções:

```http
If-Match: "4"
```

ou:

```json
{
  "expected_version": 4
}
```

Quando a versão divergir, a API deve retornar:

```http
409 Conflict
```

```json
{
  "error": {
    "code": "version_conflict",
    "message": "A história foi modificada desde o último carregamento.",
    "details": {
      "expected_version": 4,
      "actual_version": 5
    }
  }
}
```

---

## 3. Endpoints de sistema

## 3.1 Health check

```http
GET /health
```

Resposta:

```json
{
  "status": "ok",
  "version": "0.1.0",
  "api_version": "v1",
  "started_at": "2026-08-15T18:00:00Z"
}
```

## 3.2 Informações do servidor

```http
GET /system/info
```

Resposta:

```json
{
  "version": "0.1.0",
  "api_version": "v1",
  "platform": "windows/amd64",
  "storage_engine": "sqlite",
  "server": {
    "host": "127.0.0.1",
    "port": 7331,
    "network_exposed": false
  },
  "features": {
    "ai_enabled": false,
    "git_available": true,
    "runtime_traces_enabled": false
  }
}
```

## 3.3 Diagnóstico

```http
GET /system/doctor
```

Resposta:

```json
{
  "status": "warning",
  "checks": [
    {
      "name": "storage_writable",
      "status": "pass",
      "message": "O diretório .storycode possui permissão de escrita."
    },
    {
      "name": "git_available",
      "status": "pass",
      "message": "Git encontrado na versão 2.45.1."
    },
    {
      "name": "index_integrity",
      "status": "warning",
      "message": "Há 2 arquivos que falharam durante a última indexação."
    }
  ]
}
```

---

## 4. Repositório e índice

## 4.1 Obter repositório atual

```http
GET /repository
```

Resposta:

```json
{
  "data": {
    "id": "01J5H9W4PGJ3P4Y1YJPK7J8HAA",
    "name": "rag-service",
    "root_path": "C:/projects/rag-service",
    "default_branch": "main",
    "head_commit_sha": "8a2bc63e2e",
    "status": "ready",
    "last_indexed_at": "2026-08-15T18:09:30Z",
    "statistics": {
      "files": 432,
      "symbols": 1250,
      "relations": 3670,
      "entry_points": 24,
      "stories": 12,
      "stories_verified": 8,
      "stories_stale": 2,
      "stories_broken": 1
    }
  }
}
```

> `root_path` pode ser omitido ou mascarado em um futuro modo de privacidade reforçada.

## 4.2 Obter configuração efetiva

```http
GET /repository/config
```

Resposta:

```json
{
  "data": {
    "version": 1,
    "repository": {
      "include": ["src/**", "tests/**", "docs/**"],
      "exclude": [".git/**", ".venv/**", "node_modules/**"]
    },
    "analysis": {
      "languages": ["python"],
      "follow_symlinks": false,
      "max_file_size_bytes": 5242880
    },
    "storage": {
      "mode": "repository",
      "engine": "sqlite"
    }
  }
}
```

## 4.3 Atualizar configuração

```http
PUT /repository/config
```

Requisição:

```json
{
  "expected_version": 1,
  "repository": {
    "include": ["src/**", "app/**", "tests/**", "docs/**"],
    "exclude": [".git/**", ".venv/**", "node_modules/**", "dist/**"]
  },
  "analysis": {
    "languages": ["python"],
    "follow_symlinks": false,
    "max_file_size_bytes": 5242880
  }
}
```

Resposta:

```http
200 OK
```

```json
{
  "data": {
    "version": 2,
    "restart_or_reindex_required": true,
    "changed_fields": [
      "repository.include",
      "repository.exclude"
    ]
  }
}
```

## 4.4 Listar execuções de indexação

```http
GET /index-runs?limit=20
```

Resposta:

```json
{
  "data": [
    {
      "id": "01J5HAA7B9BQ9GCFZ22EGXQRKD",
      "kind": "incremental",
      "status": "completed_with_warnings",
      "started_at": "2026-08-15T18:09:10Z",
      "finished_at": "2026-08-15T18:09:30Z",
      "files_scanned": 18,
      "files_indexed": 12,
      "files_failed": 1,
      "symbols_found": 64,
      "relations_found": 122
    }
  ],
  "page": {
    "limit": 20,
    "next_cursor": null,
    "total": 1
  }
}
```

## 4.5 Obter detalhes de indexação

```http
GET /index-runs/{index_run_id}
```

Resposta:

```json
{
  "data": {
    "id": "01J5HAA7B9BQ9GCFZ22EGXQRKD",
    "kind": "incremental",
    "status": "completed_with_warnings",
    "head_commit_sha": "8a2bc63e2e",
    "started_at": "2026-08-15T18:09:10Z",
    "finished_at": "2026-08-15T18:09:30Z",
    "statistics": {
      "files_scanned": 18,
      "files_indexed": 12,
      "files_failed": 1,
      "symbols_found": 64,
      "relations_found": 122
    },
    "issues": [
      {
        "path": "src/legacy/broken_file.py",
        "kind": "parse_error",
        "message": "Erro de sintaxe na linha 43."
      }
    ]
  }
}
```

## 4.6 Iniciar indexação

```http
POST /index-runs
```

Requisição:

```json
{
  "kind": "incremental",
  "force": false
}
```

Tipos permitidos:

```text
incremental
full
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HAA7B9BQ9GCFZ22EGXQRKD",
    "kind": "incremental",
    "status": "queued",
    "events_url": "/api/v1/index-runs/01J5HAA7B9BQ9GCFZ22EGXQRKD/events"
  }
}
```

## 4.7 Cancelar indexação

```http
POST /index-runs/{index_run_id}/cancel
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HAA7B9BQ9GCFZ22EGXQRKD",
    "status": "cancelling"
  }
}
```

## 4.8 Acompanhar indexação em tempo real

```http
GET /index-runs/{index_run_id}/events
Accept: text/event-stream
```

Eventos SSE:

```text
event: progress
data: {
  "phase": "extract_symbols",
  "processed": 182,
  "total": 432,
  "percent": 42,
  "message": "Extraindo símbolos Python"
}

event: warning
data: {
  "path": "src/legacy/broken_file.py",
  "kind": "parse_error",
  "message": "Erro de sintaxe na linha 43"
}

event: completed
data: {
  "status": "completed_with_warnings",
  "finished_at": "2026-08-15T18:09:30Z"
}
```

---

## 5. Histórias

## 5.1 Listar histórias

```http
GET /stories
```

Parâmetros suportados:

```text
q
status
tag
entry_type
component_id
actor_kind
sort
order
limit
cursor
```

Exemplo:

```http
GET /stories?q=rag&status=verified&sort=updated_at&order=desc
```

Resposta:

```json
{
  "data": [
    {
      "id": "01J5HBD40XKYZ1D92Y8Q1PQZP4",
      "key": "answer-with-rag",
      "title": "Responder uma pergunta com contexto recuperado",
      "summary": "Fluxo de consulta RAG com isolamento por tenant.",
      "intent": "Entregar uma resposta baseada em documentos autorizados.",
      "status": "verified",
      "verification_status": "verified",
      "confidence": "high",
      "primary_trigger": {
        "kind": "http",
        "label": "POST /v1/chat"
      },
      "tags": ["rag", "chat", "security"],
      "scene_count": 6,
      "updated_at": "2026-08-15T18:00:00Z"
    }
  ],
  "page": {
    "limit": 20,
    "next_cursor": null,
    "total": 1
  }
}
```

## 5.2 Criar história

```http
POST /stories
```

Requisição mínima:

```json
{
  "key": "answer-with-rag",
  "title": "Responder uma pergunta com contexto recuperado",
  "intent": "Entregar uma resposta baseada em documentos autorizados.",
  "status": "draft",
  "tags": ["rag", "chat"]
}
```

Resposta:

```http
201 Created
Location: /api/v1/stories/answer-with-rag
```

```json
{
  "data": {
    "id": "01J5HBD40XKYZ1D92Y8Q1PQZP4",
    "key": "answer-with-rag",
    "status": "draft",
    "version": 1,
    "created_at": "2026-08-15T18:11:00Z"
  }
}
```

## 5.3 Obter história

```http
GET /stories/{story_key}
```

Resposta:

```json
{
  "data": {
    "id": "01J5HBD40XKYZ1D92Y8Q1PQZP4",
    "key": "answer-with-rag",
    "title": "Responder uma pergunta com contexto recuperado",
    "summary": "Fluxo de consulta RAG com isolamento por tenant.",
    "intent": "Entregar uma resposta baseada em documentos autorizados.",
    "outcome": "A API devolve resposta e fontes rastreáveis.",
    "status": "verified",
    "verification_status": "verified",
    "confidence": "high",
    "owner": null,
    "tags": ["rag", "chat", "security"],
    "version": 4,
    "created_at": "2026-08-14T15:00:00Z",
    "updated_at": "2026-08-15T18:00:00Z",
    "last_verified_at": "2026-08-15T18:09:30Z",
    "statistics": {
      "actors": 4,
      "paths": 3,
      "scenes": 6,
      "evidences": 12,
      "broken_evidences": 0
    }
  }
}
```

## 5.4 Atualizar história

```http
PUT /stories/{story_key}
If-Match: "4"
```

Requisição:

```json
{
  "title": "Responder uma pergunta usando RAG",
  "summary": "Fluxo de resposta fundamentada com recuperação vetorial.",
  "intent": "Permitir consultas seguras sobre documentos autorizados.",
  "outcome": "A API devolve uma resposta com fontes.",
  "status": "review",
  "tags": ["rag", "chat", "tenant-isolation"],
  "owner": "architecture-team"
}
```

Resposta:

```http
200 OK
```

```json
{
  "data": {
    "key": "answer-with-rag",
    "version": 5,
    "updated_at": "2026-08-15T18:11:45Z"
  }
}
```

## 5.5 Arquivar história

```http
POST /stories/{story_key}/archive
If-Match: "5"
```

Requisição opcional:

```json
{
  "reason": "Fluxo substituído pela história rag-answer-v2."
}
```

Resposta:

```http
200 OK
```

```json
{
  "data": {
    "key": "answer-with-rag",
    "status": "archived",
    "archived_at": "2026-08-15T18:12:00Z"
  }
}
```

## 5.6 Restaurar história

```http
POST /stories/{story_key}/restore
```

Resposta:

```http
200 OK
```

```json
{
  "data": {
    "key": "answer-with-rag",
    "status": "draft"
  }
}
```

## 5.7 Duplicar história

```http
POST /stories/{story_key}/duplicate
```

Requisição:

```json
{
  "key": "answer-with-rag-v2",
  "title": "Responder uma pergunta usando RAG v2",
  "include_evidence": true
}
```

Resposta:

```http
201 Created
```

## 5.8 Descobrir histórias

```http
POST /story-discoveries
```

Requisição:

```json
{
  "entry_point_ids": [
    "01J5H8KDXFADAEZBMDN7MJ2V2R"
  ],
  "entry_types": ["http"],
  "max_depth": 12,
  "include_tests": true,
  "create_drafts": true
}
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HCE2F0YB7S4H9KJ6M3DP0J",
    "status": "queued",
    "events_url": "/api/v1/story-discoveries/01J5HCE2F0YB7S4H9KJ6M3DP0J/events"
  }
}
```

## 5.9 Consultar descoberta

```http
GET /story-discoveries/{discovery_id}
```

Resposta:

```json
{
  "data": {
    "id": "01J5HCE2F0YB7S4H9KJ6M3DP0J",
    "status": "completed",
    "created_story_keys": [
      "post-v1-chat"
    ],
    "candidates": [
      {
        "entry_point": {
          "kind": "http",
          "label": "POST /v1/chat"
        },
        "suggested_key": "post-v1-chat",
        "suggested_title": "Processar solicitação de chat",
        "confidence": "medium",
        "scene_count": 5
      }
    ]
  }
}
```

---

## 6. Story Player

## 6.1 Obter modelo completo para reprodução

```http
GET /stories/{story_key}/player
```

Esse endpoint deve devolver o contrato otimizado para o frontend. A UI não deve precisar montar o grafo consultando múltiplos endpoints.

Resposta:

```json
{
  "data": {
    "story": {
      "id": "01J5HBD40XKYZ1D92Y8Q1PQZP4",
      "key": "answer-with-rag",
      "title": "Responder uma pergunta com contexto recuperado",
      "intent": "Permitir consultas seguras sobre documentos autorizados.",
      "outcome": "A API devolve uma resposta com fontes.",
      "status": "verified",
      "verification_status": "verified",
      "confidence": "high",
      "tags": ["rag", "chat", "security"]
    },
    "triggers": [
      {
        "id": "01J5H8KDXFADAEZBMDN7MJ2V2R",
        "kind": "http",
        "label": "POST /v1/chat",
        "method": "POST",
        "path": "/v1/chat",
        "is_primary": true,
        "confidence": "confirmed"
      }
    ],
    "actors": [
      {
        "id": "actor-user",
        "key": "user",
        "label": "Usuário",
        "kind": "human",
        "component_id": null,
        "position_hint": "left"
      },
      {
        "id": "actor-api",
        "key": "chat-api",
        "label": "Chat API",
        "kind": "service",
        "component_id": "component-chat-api",
        "position_hint": "center"
      },
      {
        "id": "actor-qdrant",
        "key": "vector-store",
        "label": "Vector Store",
        "kind": "database",
        "component_id": "component-qdrant",
        "position_hint": "right"
      },
      {
        "id": "actor-llm",
        "key": "llm-provider",
        "label": "LLM Provider",
        "kind": "external_api",
        "component_id": null,
        "position_hint": "right"
      }
    ],
    "paths": [
      {
        "id": "path-happy",
        "key": "happy-path",
        "label": "Caminho principal",
        "kind": "happy_path",
        "is_default": true,
        "entry_scene_id": "scene-validate",
        "exit_scene_id": "scene-response"
      },
      {
        "id": "path-no-context",
        "key": "no-context",
        "label": "Sem contexto suficiente",
        "kind": "fallback",
        "is_default": false,
        "entry_scene_id": "scene-retrieve",
        "exit_scene_id": "scene-no-context"
      }
    ],
    "scenes": [
      {
        "id": "scene-validate",
        "key": "validate-request",
        "type": "validate",
        "title": "Validar solicitação",
        "narration": "A API valida identidade, payload e escopo do tenant.",
        "technical_summary": "O handler executa validação do request e autenticação.",
        "from_actor_id": "actor-user",
        "to_actor_id": "actor-api",
        "primary_symbol": {
          "id": "symbol-create-chat",
          "ref": "src/api/chat.py::create_chat"
        },
        "status": "verified",
        "confidence": "high",
        "input_summary": "ChatRequest",
        "output_summary": "AuthenticatedChatRequest",
        "evidence_summary": {
          "total": 3,
          "verified": 3,
          "changed": 0,
          "missing": 0
        }
      },
      {
        "id": "scene-retrieve",
        "key": "retrieve-context",
        "type": "read",
        "title": "Recuperar contexto autorizado",
        "narration": "O serviço busca chunks relevantes dentro do tenant solicitado.",
        "from_actor_id": "actor-api",
        "to_actor_id": "actor-qdrant",
        "primary_symbol": {
          "id": "symbol-retrieve",
          "ref": "src/services/retrieval.py::retrieve"
        },
        "status": "verified",
        "confidence": "high",
        "input_summary": "Embedding + tenant_id",
        "output_summary": "RetrievedChunk[]",
        "evidence_summary": {
          "total": 2,
          "verified": 2,
          "changed": 0,
          "missing": 0
        }
      }
    ],
    "transitions": [
      {
        "id": "transition-1",
        "from_scene_id": "scene-validate",
        "to_scene_id": "scene-retrieve",
        "kind": "sequence",
        "is_default": true
      }
    ],
    "invariants": [
      {
        "id": "invariant-tenant-isolation",
        "key": "tenant-isolation",
        "statement": "Dados de outro tenant não podem ser usados como contexto.",
        "kind": "authorization",
        "severity": "critical",
        "status": "verified"
      }
    ],
    "failure_modes": [
      {
        "id": "failure-no-context",
        "scene_id": "scene-retrieve",
        "title": "Nenhum contexto suficiente encontrado",
        "category": "business_rule",
        "impact": "low",
        "handling": "Responder sem inventar fontes.",
        "status": "tested"
      }
    ],
    "visualization": {
      "default_view": "timeline",
      "recommended_scene_order": [
        "scene-validate",
        "scene-retrieve",
        "scene-generate",
        "scene-response"
      ]
    }
  }
}
```

## 6.2 Obter cena

```http
GET /stories/{story_key}/scenes/{scene_key}
```

Resposta:

```json
{
  "data": {
    "id": "scene-retrieve",
    "key": "retrieve-context",
    "type": "read",
    "title": "Recuperar contexto autorizado",
    "narration": "O serviço busca chunks relevantes dentro do tenant solicitado.",
    "from_actor": {
      "id": "actor-api",
      "label": "Chat API",
      "kind": "service"
    },
    "to_actor": {
      "id": "actor-qdrant",
      "label": "Vector Store",
      "kind": "database"
    },
    "data_flows": [
      {
        "artifact": "EmbeddingVector",
        "direction": "input",
        "operation": "filter"
      },
      {
        "artifact": "RetrievedChunk[]",
        "direction": "output",
        "operation": "retrieve"
      }
    ],
    "invariants": [
      {
        "key": "tenant-isolation",
        "statement": "Dados de outro tenant não podem ser usados como contexto.",
        "severity": "critical",
        "status": "verified"
      }
    ],
    "evidences_url": "/api/v1/stories/answer-with-rag/scenes/retrieve-context/evidences",
    "transitions": {
      "incoming": [],
      "outgoing": [
        {
          "to_scene_key": "generate-answer",
          "kind": "sequence",
          "label": null
        },
        {
          "to_scene_key": "no-context",
          "kind": "fallback",
          "condition": "Nenhum chunk passa o score mínimo."
        }
      ]
    }
  }
}
```

## 6.3 Criar cena

```http
POST /stories/{story_key}/scenes
If-Match: "4"
```

Requisição:

```json
{
  "key": "validate-request",
  "type": "validate",
  "title": "Validar solicitação",
  "narration": "A API valida identidade, payload e escopo do tenant.",
  "from_actor_key": "user",
  "to_actor_key": "chat-api",
  "primary_symbol_ref": "src/api/chat.py::create_chat",
  "input_summary": "ChatRequest",
  "output_summary": "AuthenticatedChatRequest"
}
```

Resposta:

```http
201 Created
```

## 6.4 Atualizar cena

```http
PUT /stories/{story_key}/scenes/{scene_key}
If-Match: "4"
```

Requisição:

```json
{
  "type": "validate",
  "title": "Validar solicitação e tenant",
  "narration": "A API valida o token, o payload e o tenant antes de iniciar a recuperação.",
  "from_actor_key": "user",
  "to_actor_key": "chat-api",
  "primary_symbol_ref": "src/api/chat.py::create_chat",
  "status": "draft"
}
```

## 6.5 Excluir cena

```http
DELETE /stories/{story_key}/scenes/{scene_key}
If-Match: "4"
```

A API deve recusar exclusão quando houver transições dependentes, salvo se for enviada a opção explícita:

```http
DELETE /stories/{story_key}/scenes/{scene_key}?cascade=true
```

---

## 7. Caminhos e transições

## 7.1 Listar caminhos

```http
GET /stories/{story_key}/paths
```

## 7.2 Criar caminho

```http
POST /stories/{story_key}/paths
If-Match: "4"
```

Requisição:

```json
{
  "key": "happy-path",
  "label": "Caminho principal",
  "kind": "happy_path",
  "is_default": true,
  "description": "Fluxo bem-sucedido de resposta com fontes."
}
```

## 7.3 Atualizar caminho

```http
PUT /stories/{story_key}/paths/{path_key}
If-Match: "4"
```

## 7.4 Excluir caminho

```http
DELETE /stories/{story_key}/paths/{path_key}
If-Match: "4"
```

O caminho padrão não pode ser excluído sem que outro caminho seja promovido a padrão na mesma operação.

## 7.5 Criar transição

```http
POST /stories/{story_key}/transitions
If-Match: "4"
```

Requisição:

```json
{
  "from_scene_key": "validate-request",
  "to_scene_key": "retrieve-context",
  "kind": "sequence",
  "label": "Solicitação validada",
  "is_default": true
}
```

Para bifurcação:

```json
{
  "from_scene_key": "retrieve-context",
  "to_scene_key": "no-context",
  "kind": "fallback",
  "condition": "Nenhum chunk passa o score mínimo.",
  "is_default": false
}
```

## 7.6 Atualizar transição

```http
PUT /stories/{story_key}/transitions/{transition_id}
If-Match: "4"
```

## 7.7 Excluir transição

```http
DELETE /stories/{story_key}/transitions/{transition_id}
If-Match: "4"
```

---

## 8. Atores e componentes

## 8.1 Listar atores da história

```http
GET /stories/{story_key}/actors
```

## 8.2 Criar ator

```http
POST /stories/{story_key}/actors
If-Match: "4"
```

Requisição:

```json
{
  "key": "vector-store",
  "label": "Vector Store",
  "kind": "database",
  "component_id": "component-qdrant",
  "description": "Armazena e recupera embeddings e chunks.",
  "sort_order": 3
}
```

## 8.3 Atualizar ator

```http
PUT /stories/{story_key}/actors/{actor_key}
If-Match: "4"
```

## 8.4 Excluir ator

```http
DELETE /stories/{story_key}/actors/{actor_key}
If-Match: "4"
```

A API deve recusar exclusão de ator utilizado por cenas, salvo com `cascade=true` ou após atualização das cenas dependentes.

## 8.5 Listar componentes

```http
GET /components
```

Parâmetros:

```text
q
kind
source_type
confidence
limit
cursor
```

Resposta:

```json
{
  "data": [
    {
      "id": "component-chat-api",
      "key": "chat-api",
      "name": "Chat API",
      "kind": "service",
      "description": "Camada HTTP da aplicação de chat.",
      "confidence": "high",
      "source_type": "static_analysis",
      "symbol_count": 28,
      "story_count": 4
    }
  ],
  "page": {
    "limit": 20,
    "next_cursor": null,
    "total": 1
  }
}
```

## 8.6 Obter componente

```http
GET /components/{component_id}
```

A resposta deve incluir símbolos, relações, histórias e pontos de entrada associados, respeitando paginação para coleções grandes.

## 8.7 Obter mapa de componente

```http
GET /components/{component_id}/map?depth=2
```

Resposta:

```json
{
  "data": {
    "nodes": [
      {
        "id": "component-chat-api",
        "type": "component",
        "label": "Chat API",
        "kind": "service",
        "highlighted": true
      },
      {
        "id": "component-qdrant",
        "type": "component",
        "label": "Vector Store",
        "kind": "database"
      }
    ],
    "edges": [
      {
        "id": "relation-1",
        "from": "component-chat-api",
        "to": "component-qdrant",
        "kind": "reads_from",
        "confidence": "high"
      }
    ]
  }
}
```

---

## 9. Evidências

## 9.1 Listar evidências da história

```http
GET /stories/{story_key}/evidences
```

Parâmetros:

```text
kind
status
scene_key
role
limit
cursor
```

## 9.2 Listar evidências da cena

```http
GET /stories/{story_key}/scenes/{scene_key}/evidences
```

Resposta:

```json
{
  "data": [
    {
      "id": "evidence-create-chat",
      "kind": "code_symbol",
      "role": "implements",
      "claim": "O handler valida a solicitação antes de iniciar o fluxo.",
      "status": "verified",
      "confidence": "high",
      "source": {
        "path": "src/api/chat.py",
        "symbol": "create_chat",
        "start_line": 31,
        "end_line": 69
      },
      "excerpt": "@router.post('/v1/chat')\nasync def create_chat(...):\n    ...",
      "last_verified_at": "2026-08-15T18:09:30Z"
    }
  ],
  "page": {
    "limit": 20,
    "next_cursor": null,
    "total": 1
  }
}
```

## 9.3 Obter evidência

```http
GET /evidences/{evidence_id}
```

Resposta:

```json
{
  "data": {
    "id": "evidence-create-chat",
    "kind": "code_symbol",
    "title": "src/api/chat.py::create_chat",
    "status": "verified",
    "confidence": "high",
    "locator": {
      "path": "src/api/chat.py",
      "symbol": "create_chat",
      "start_line": 31,
      "end_line": 69
    },
    "excerpt": "@router.post('/v1/chat')\nasync def create_chat(...):\n    ...",
    "content_hash": "sha256:...",
    "last_seen_index_run_id": "01J5HAA7B9BQ9GCFZ22EGXQRKD",
    "verifications": [
      {
        "status": "verified",
        "verified_at": "2026-08-15T18:09:30Z",
        "message": "Símbolo localizado e hash semântico compatível."
      }
    ],
    "used_by": {
      "stories": [
        {
          "key": "answer-with-rag",
          "title": "Responder uma pergunta com contexto recuperado"
        }
      ],
      "scenes": [
        {
          "story_key": "answer-with-rag",
          "scene_key": "validate-request",
          "title": "Validar solicitação"
        }
      ]
    }
  }
}
```

## 9.4 Associar evidência à cena

```http
POST /stories/{story_key}/scenes/{scene_key}/evidences
If-Match: "4"
```

Requisição usando uma evidência existente:

```json
{
  "evidence_id": "evidence-create-chat",
  "role": "implements",
  "claim": "O handler valida a solicitação antes de iniciar o fluxo.",
  "is_primary": true
}
```

Ou criando referência técnica:

```json
{
  "source": {
    "kind": "code_symbol",
    "ref": "src/api/chat.py::create_chat"
  },
  "role": "implements",
  "claim": "O handler valida a solicitação antes de iniciar o fluxo.",
  "is_primary": true
}
```

## 9.5 Remover evidência da cena

```http
DELETE /stories/{story_key}/scenes/{scene_key}/evidences/{evidence_id}
If-Match: "4"
```

## 9.6 Abrir evidência no editor

```http
POST /evidences/{evidence_id}/open
```

Requisição:

```json
{
  "editor": "vscode"
}
```

Resposta:

```http
204 No Content
```

Se não houver editor configurado, deve retornar:

```http
422 Unprocessable Entity
```

```json
{
  "error": {
    "code": "editor_not_configured",
    "message": "Nenhum editor local foi configurado para abrir evidências."
  }
}
```

---

## 10. Código, símbolos e pontos de entrada

## 10.1 Buscar no índice

```http
GET /search?q=retrieve&types=symbol,story,entry_point
```

Resposta:

```json
{
  "data": [
    {
      "type": "symbol",
      "id": "symbol-retrieve",
      "label": "RetrievalService.retrieve",
      "description": "src/services/retrieval.py:41",
      "score": 0.98,
      "ref": "src/services/retrieval.py::RetrievalService.retrieve"
    },
    {
      "type": "story",
      "key": "answer-with-rag",
      "label": "Responder uma pergunta com contexto recuperado",
      "description": "Fluxo de consulta RAG com isolamento por tenant.",
      "score": 0.83
    }
  ]
}
```

## 10.2 Listar pontos de entrada

```http
GET /entry-points
```

Parâmetros:

```text
kind
framework
q
component_id
limit
cursor
```

## 10.3 Obter ponto de entrada

```http
GET /entry-points/{entry_point_id}
```

Resposta:

```json
{
  "data": {
    "id": "01J5H8KDXFADAEZBMDN7MJ2V2R",
    "kind": "http",
    "key": "http:POST:/v1/chat",
    "label": "POST /v1/chat",
    "method": "POST",
    "path": "/v1/chat",
    "framework": "fastapi",
    "handler": {
      "id": "symbol-create-chat",
      "ref": "src/api/chat.py::create_chat"
    },
    "confidence": "confirmed",
    "stories": [
      {
        "key": "answer-with-rag",
        "title": "Responder uma pergunta com contexto recuperado",
        "status": "verified"
      }
    ]
  }
}
```

## 10.4 Obter símbolo

```http
GET /symbols/{symbol_id}
```

Resposta:

```json
{
  "data": {
    "id": "symbol-retrieve",
    "ref": "src/services/retrieval.py::RetrievalService.retrieve",
    "display_name": "retrieve",
    "kind": "method",
    "signature": "async def retrieve(self, query: str, tenant_id: str) -> list[Chunk]",
    "location": {
      "path": "src/services/retrieval.py",
      "start_line": 41,
      "end_line": 88
    },
    "component": {
      "id": "component-retrieval",
      "name": "Retrieval Service"
    },
    "relations": {
      "calls": 4,
      "called_by": 2,
      "reads": 1,
      "tests": 3,
      "stories": 2
    }
  }
}
```

## 10.5 Obter relações de um símbolo

```http
GET /symbols/{symbol_id}/relations?direction=both&kind=calls&depth=2
```

Resposta:

```json
{
  "data": {
    "nodes": [
      {
        "id": "symbol-retrieve",
        "type": "symbol",
        "label": "RetrievalService.retrieve",
        "highlighted": true
      }
    ],
    "edges": [
      {
        "id": "relation-123",
        "from": "symbol-retrieve",
        "to": "symbol-qdrant-search",
        "kind": "calls",
        "confidence": "high",
        "line": 53
      }
    ]
  }
}
```

## 10.6 Obter conteúdo sanitizado de arquivo

```http
GET /files/content?path=src/services/retrieval.py&start_line=35&end_line=95
```

Resposta:

```json
{
  "data": {
    "path": "src/services/retrieval.py",
    "language": "python",
    "start_line": 35,
    "end_line": 95,
    "content": "class RetrievalService:\n    async def retrieve(...):\n        ...",
    "redactions": []
  }
}
```

A API deve aplicar mascaramento de segredos antes de devolver conteúdo.

---

## 11. Invariantes, falhas e dados

## 11.1 Listar invariantes

```http
GET /stories/{story_key}/invariants
```

## 11.2 Criar invariante

```http
POST /stories/{story_key}/invariants
If-Match: "4"
```

Requisição:

```json
{
  "key": "tenant-isolation",
  "statement": "Dados de outro tenant não podem ser usados como contexto.",
  "kind": "authorization",
  "severity": "critical",
  "verification_method": "automated_test",
  "scene_keys": ["validate-request", "retrieve-context"]
}
```

## 11.3 Atualizar invariante

```http
PUT /stories/{story_key}/invariants/{invariant_key}
If-Match: "4"
```

## 11.4 Excluir invariante

```http
DELETE /stories/{story_key}/invariants/{invariant_key}
If-Match: "4"
```

## 11.5 Listar modos de falha

```http
GET /stories/{story_key}/failure-modes
```

## 11.6 Criar modo de falha

```http
POST /stories/{story_key}/failure-modes
If-Match: "4"
```

Requisição:

```json
{
  "key": "no-context",
  "scene_key": "retrieve-context",
  "title": "Nenhum contexto suficiente encontrado",
  "description": "Nenhum chunk atingiu o score mínimo configurado.",
  "category": "business_rule",
  "impact": "low",
  "handling": "Retornar resposta informando ausência de evidência.",
  "recovery_path_key": "no-context"
}
```

## 11.7 Listar fluxo de dados da cena

```http
GET /stories/{story_key}/scenes/{scene_key}/data-flows
```

Resposta:

```json
{
  "data": [
    {
      "id": "data-flow-embedding-input",
      "artifact": {
        "id": "artifact-embedding",
        "name": "EmbeddingVector",
        "kind": "dto",
        "classification": "internal"
      },
      "direction": "input",
      "operation": "filter",
      "is_sensitive": false,
      "confidence": "high"
    }
  ]
}
```

---

## 12. Verificação e drift

## 12.1 Iniciar verificação

```http
POST /verifications
```

Requisição:

```json
{
  "story_keys": ["answer-with-rag"],
  "scope": "stories"
}
```

Ou para todas as histórias:

```json
{
  "scope": "repository"
}
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HDZAE2P31XWWJVJTR25QTD",
    "status": "queued",
    "events_url": "/api/v1/verifications/01J5HDZAE2P31XWWJVJTR25QTD/events"
  }
}
```

## 12.2 Consultar verificação

```http
GET /verifications/{verification_id}
```

Resposta:

```json
{
  "data": {
    "id": "01J5HDZAE2P31XWWJVJTR25QTD",
    "status": "completed",
    "started_at": "2026-08-15T18:13:00Z",
    "finished_at": "2026-08-15T18:13:08Z",
    "summary": {
      "stories_checked": 12,
      "verified": 8,
      "stale": 2,
      "broken": 1,
      "unavailable": 1
    },
    "affected_stories": [
      {
        "key": "answer-with-rag",
        "previous_status": "verified",
        "current_status": "stale",
        "issues": [
          {
            "scene_key": "retrieve-context",
            "evidence_id": "evidence-retrieve",
            "status": "changed",
            "message": "O hash semântico do símbolo mudou."
          }
        ]
      }
    ]
  }
}
```

## 12.3 Eventos de verificação

```http
GET /verifications/{verification_id}/events
Accept: text/event-stream
```

Eventos:

```text
event: progress
data: {
  "processed_stories": 5,
  "total_stories": 12,
  "current_story_key": "answer-with-rag"
}

event: story_updated
data: {
  "story_key": "answer-with-rag",
  "previous_status": "verified",
  "current_status": "stale"
}

event: completed
data: {
  "status": "completed"
}
```

## 12.4 Obter relatório de drift da história

```http
GET /stories/{story_key}/drift
```

Resposta:

```json
{
  "data": {
    "story_key": "answer-with-rag",
    "status": "stale",
    "last_verified_at": "2026-08-14T18:00:00Z",
    "detected_at": "2026-08-15T18:13:08Z",
    "issues": [
      {
        "severity": "warning",
        "scene_key": "retrieve-context",
        "evidence": {
          "id": "evidence-retrieve",
          "kind": "code_symbol",
          "ref": "src/services/retrieval.py::retrieve"
        },
        "status": "changed",
        "message": "O símbolo foi alterado desde a última verificação.",
        "recommended_action": "Revise a narrativa e as evidências da cena."
      }
    ]
  }
}
```

---

## 13. Impacto de mudanças

## 13.1 Analisar impacto por arquivo, símbolo ou diff Git

```http
POST /impact-analysis
```

Por referências explícitas:

```json
{
  "targets": [
    {
      "type": "file",
      "path": "src/services/retrieval.py"
    },
    {
      "type": "symbol",
      "ref": "src/services/generation.py::generate"
    }
  ],
  "depth": 3
}
```

Por diff Git:

```json
{
  "git": {
    "base_ref": "main",
    "head_ref": "HEAD"
  },
  "depth": 3
}
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HEVQY8J70BJEAEXEZTWQRV",
    "status": "queued",
    "events_url": "/api/v1/impact-analysis/01J5HEVQY8J70BJEAEXEZTWQRV/events"
  }
}
```

## 13.2 Consultar análise de impacto

```http
GET /impact-analysis/{analysis_id}
```

Resposta:

```json
{
  "data": {
    "id": "01J5HEVQY8J70BJEAEXEZTWQRV",
    "status": "completed",
    "targets": [
      {
        "type": "file",
        "path": "src/services/retrieval.py"
      }
    ],
    "affected_stories": [
      {
        "key": "answer-with-rag",
        "title": "Responder uma pergunta com contexto recuperado",
        "impact": "direct",
        "affected_scenes": [
          {
            "key": "retrieve-context",
            "title": "Recuperar contexto autorizado"
          }
        ],
        "affected_invariants": [
          {
            "key": "tenant-isolation",
            "severity": "critical"
          }
        ]
      }
    ],
    "affected_tests": [
      {
        "ref": "tests/integration/test_chat.py::test_filters_by_tenant",
        "relation": "directly_tests"
      }
    ]
  }
}
```

---

## 14. Exportação e importação

## 14.1 Exportar história

```http
POST /stories/{story_key}/exports
```

Requisição:

```json
{
  "format": "markdown",
  "include": {
    "evidences": true,
    "invariants": true,
    "failure_modes": true,
    "code_excerpts": false
  }
}
```

Formatos aceitos:

```text
markdown
json
mermaid
html
```

Resposta:

```http
202 Accepted
```

```json
{
  "data": {
    "id": "01J5HFC6CJ4E3C2T6HAT9P9C2P",
    "status": "completed",
    "format": "markdown",
    "download_url": "/api/v1/exports/01J5HFC6CJ4E3C2T6HAT9P9C2P/content"
  }
}
```

> Como o sistema é local, `download_url` é uma URL temporária local; a CLI pode salvar diretamente em disco.

## 14.2 Baixar exportação

```http
GET /exports/{export_id}/content
```

O `Content-Type` deve variar por formato:

```text
text/markdown; charset=utf-8
application/json; charset=utf-8
text/plain; charset=utf-8
text/html; charset=utf-8
```

## 14.3 Importar histórias

```http
POST /story-imports
```

`Content-Type`:

```text
multipart/form-data
```

Campos:

```text
file
mode = validate | preview | apply
conflict_strategy = reject | skip | overwrite
```

Resposta de preview:

```json
{
  "data": {
    "status": "preview",
    "valid": true,
    "stories_found": 2,
    "conflicts": [
      {
        "key": "answer-with-rag",
        "reason": "Já existe uma história com esta chave."
      }
    ]
  }
}
```

---

## 15. Git, documentos e linha do tempo

## 15.1 Listar commits relacionados à história

```http
GET /stories/{story_key}/commits?limit=20
```

Resposta:

```json
{
  "data": [
    {
      "sha": "8a2bc63e2e",
      "short_sha": "8a2bc63",
      "message": "feat: add tenant filter to retrieval",
      "author_name": "Developer",
      "committed_at": "2026-08-12T14:33:00Z",
      "relation": "changed_evidence"
    }
  ]
}
```

## 15.2 Listar documentos relacionados

```http
GET /stories/{story_key}/documents
```

## 15.3 Associar documento a uma história

```http
POST /stories/{story_key}/documents
If-Match: "4"
```

Requisição:

```json
{
  "document_id": "document-adr-qdrant",
  "role": "explains_decision"
}
```

## 15.4 Obter linha do tempo da história

```http
GET /stories/{story_key}/timeline
```

Resposta:

```json
{
  "data": [
    {
      "at": "2026-08-10T10:00:00Z",
      "type": "story_created",
      "title": "História criada",
      "description": "Criada a partir de descoberta automática."
    },
    {
      "at": "2026-08-12T14:33:00Z",
      "type": "evidence_changed",
      "title": "Evidência alterada",
      "description": "O símbolo RetrievalService.retrieve foi modificado.",
      "commit": {
        "short_sha": "8a2bc63",
        "message": "feat: add tenant filter to retrieval"
      }
    }
  ]
}
```

---

## 16. API de mapa e visualização

## 16.1 Mapa da história

```http
GET /stories/{story_key}/map?view=components
```

Valores possíveis de `view`:

```text
components
symbols
data_flow
evidence
```

Resposta:

```json
{
  "data": {
    "story_key": "answer-with-rag",
    "view": "components",
    "nodes": [
      {
        "id": "actor-user",
        "type": "actor",
        "label": "Usuário",
        "kind": "human",
        "story_actor_key": "user"
      },
      {
        "id": "component-chat-api",
        "type": "component",
        "label": "Chat API",
        "kind": "service"
      }
    ],
    "edges": [
      {
        "id": "edge-user-api",
        "from": "actor-user",
        "to": "component-chat-api",
        "kind": "receives_http_request",
        "scene_keys": ["validate-request"],
        "confidence": "high"
      }
    ]
  }
}
```

## 16.2 Layout persistido do usuário

```http
GET /stories/{story_key}/layout
```

```http
PUT /stories/{story_key}/layout
If-Match: "4"
```

Requisição:

```json
{
  "view": "system_map",
  "nodes": {
    "actor-user": {
      "x": 80,
      "y": 220
    },
    "actor-api": {
      "x": 360,
      "y": 220
    }
  },
  "viewport": {
    "x": 0,
    "y": 0,
    "zoom": 1
  }
}
```

No MVP, esse layout pode ficar em `visual_metadata` da história; futuramente poderá ser separado por perfil local.

---

## 17. Eventos globais

Para atualizar a UI após indexação, verificação ou alterações manuais, o frontend pode abrir uma conexão SSE global.

```http
GET /events
Accept: text/event-stream
```

Eventos possíveis:

```text
repository.index.started
repository.index.progress
repository.index.completed
repository.index.failed

story.created
story.updated
story.archived
story.verified
story.stale
story.broken

evidence.changed
evidence.missing

verification.started
verification.completed

impact_analysis.completed
```

Exemplo:

```text
event: story.updated
data: {
  "story_key": "answer-with-rag",
  "version": 5,
  "updated_at": "2026-08-15T18:11:45Z"
}
```

---

## 18. Segurança da API local

### 18.1 Escopo de rede

A API deve escutar apenas em:

```text
127.0.0.1
::1
localhost
```

Exposição em rede exige flag explícita:

```bash
storycode serve --host 0.0.0.0 --allow-network
```

### 18.2 Proteção de sessão local

Mesmo em loopback, endpoints mutáveis devem exigir um token efêmero de sessão.

Ao iniciar `storycode serve`, o processo deve gerar um token aleatório em memória e entregar a interface web por cookie `HttpOnly`, `SameSite=Strict`, ou exigir header:

```http
X-StoryCode-Session: <token>
```

O token:

- Não deve ser salvo em disco por padrão.
- Deve expirar ao encerrar o processo.
- Deve ser rotacionado a cada inicialização.
- Deve ser obrigatório para `POST`, `PUT`, `PATCH` e `DELETE`.

### 18.3 CORS e origem

A API deve aceitar por padrão somente a origem local do Studio atual.

Exemplo:

```text
Origin: http://127.0.0.1:7331
```

Origens externas devem ser bloqueadas, salvo configuração explícita.

### 18.4 Conteúdo seguro

Endpoints de arquivo e evidência devem:

- Restringir caminhos ao diretório do repositório.
- Bloquear path traversal.
- Aplicar mascaramento de segredos.
- Escapar conteúdo para renderização HTML.
- Não permitir que o browser execute conteúdo do repositório como script.

---

## 19. Priorização do MVP

### Obrigatórios

```text
GET  /health
GET  /repository
GET  /repository/config
POST /index-runs
GET  /index-runs/{id}
GET  /index-runs/{id}/events

GET  /stories
POST /stories
GET  /stories/{story_key}
PUT  /stories/{story_key}

GET  /stories/{story_key}/player
GET  /stories/{story_key}/scenes/{scene_key}
POST /stories/{story_key}/scenes
PUT  /stories/{story_key}/scenes/{scene_key}

GET  /stories/{story_key}/scenes/{scene_key}/evidences
GET  /evidences/{evidence_id}

GET  /entry-points
GET  /symbols/{symbol_id}
GET  /files/content

POST /verifications
GET  /verifications/{id}
GET  /stories/{story_key}/drift

GET  /search
GET  /events
```

### Segunda entrega

```text
POST /story-discoveries
GET  /story-discoveries/{id}

POST /stories/{story_key}/actors
POST /stories/{story_key}/paths
POST /stories/{story_key}/transitions

GET  /components
GET  /components/{id}/map

POST /impact-analysis
GET  /impact-analysis/{id}

POST /stories/{story_key}/exports
GET  /exports/{id}/content

GET  /stories/{story_key}/commits
GET  /stories/{story_key}/timeline
```

### Pós-MVP

```text
POST /story-imports
POST /evidences/{id}/open
PUT  /stories/{story_key}/layout
GET  /stories/{story_key}/layout

Integração de runtime traces
Integração de IDE
API de plugins
Autenticação para uso em rede
```

---

## 20. Decisão de design

A API não deve expor o banco SQLite diretamente nem obrigar o frontend a entender o modelo relacional completo.

O endpoint central para a experiência visual é:

```http
GET /stories/{story_key}/player
```

Ele retorna um **view model narrativo** pronto para timeline, mapa, animação, painéis de evidência e caminhos alternativos.

Os endpoints técnicos — símbolos, arquivos, relações, evidências, indexação e verificação — existem para aprofundar a narrativa, editar conteúdo e manter tudo verificável.