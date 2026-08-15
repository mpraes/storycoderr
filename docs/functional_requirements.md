# Requisitos Funcionais — StoryCode

## 1. Propósito

O StoryCode deve permitir que pessoas entendam um repositório de software como um conjunto de histórias interativas e verificáveis.

Uma história descreve uma jornada do sistema — por exemplo, uma requisição HTTP, um job assíncrono, um comando CLI ou o consumo de um evento — conectando:

- Intenção e resultado esperado
- Gatilho de entrada
- Atores e componentes envolvidos
- Etapas e decisões do fluxo
- Dados e efeitos colaterais
- Evidências no código, testes, contratos e histórico Git
- Possíveis falhas, caminhos alternativos e invariantes

O produto deve priorizar entendimento rápido de sistemas desconhecidos, onboarding técnico e documentação que se mantenha vinculada ao código.

---

## 2. Personas

### 2.1 Pessoa desenvolvedora em onboarding

Deseja entender como uma funcionalidade funciona sem percorrer manualmente dezenas de arquivos, símbolos e chamadas.

### 2.2 Pessoa mantenedora do sistema

Deseja identificar impacto, dependências, fluxos afetados e documentação desatualizada antes ou depois de uma mudança.

### 2.3 Pessoa arquiteta ou tech lead

Deseja comunicar arquitetura, jornadas críticas, integrações e decisões técnicas de maneira acessível e auditável.

### 2.4 Pessoa não técnica ou parcialmente técnica

Deseja compreender o comportamento de uma funcionalidade em uma visão narrativa, sem precisar interpretar o código-fonte.

---

## 3. Glossário

| Termo | Definição |
|---|---|
| Repositório | Diretório Git local analisado pelo StoryCode |
| Índice | Base local com informações extraídas do repositório |
| História | Jornada narrável e verificável de um comportamento do sistema |
| Jornada | Fluxo iniciado por um gatilho e concluído por um resultado ou falha |
| Cena | Etapa visual e narrativa dentro de uma história |
| Gatilho | Evento que inicia uma jornada: rota HTTP, evento, cron, CLI, webhook ou fila |
| Ator | Pessoa, serviço, banco, fila, API externa ou processo participante |
| Evidência | Referência verificável a código, teste, contrato, documento, commit ou trace |
| Símbolo | Unidade identificável do código, como função, método, classe, módulo ou rota |
| Invariante | Regra que deve permanecer verdadeira ao longo de uma história |
| Drift | Divergência entre a história documentada e as evidências atuais do repositório |
| Rascunho | História descoberta ou gerada que ainda não foi confirmada por uma pessoa |
| História verificada | História cujas evidências foram validadas contra o índice atual |

---

## 4. Escopo do MVP

O MVP deve suportar um repositório local por execução, com análise inicial para projetos Python e interface web local.

O MVP deve entregar uma experiência completa para:

1. Indexar um repositório.
2. Descobrir pontos de entrada.
3. Criar ou importar histórias.
4. Visualizar histórias no browser.
5. Navegar das cenas para evidências concretas.
6. Verificar se histórias continuam consistentes após mudanças no código.
7. Exportar uma história em formatos compartilháveis.

O MVP não deve depender de um LLM, serviço cloud, banco de dados remoto ou conta de usuário.

---

## 5. Gestão de repositório

### RF-001 — Inicializar um projeto StoryCode

O sistema deve permitir inicializar a configuração do StoryCode em um repositório local.

```bash
storycode init
```

A inicialização deve criar:

```text
.storycode/
├── config.yaml
├── stories/
├── index/
└── cache/
```

O comando não deve modificar arquivos de código-fonte da aplicação analisada.

### RF-002 — Configurar o repositório

O sistema deve permitir configurar:

- Diretórios incluídos e ignorados.
- Linguagens habilitadas para análise.
- Padrões de arquivos de teste.
- Diretórios de documentação.
- Diretórios de migrations.
- Arquivos de contrato, como OpenAPI, AsyncAPI, GraphQL e protobuf.
- Branch Git de referência.
- Opções de privacidade e providers de IA, quando habilitados futuramente.

Exemplo:

```yaml
repository:
  include:
    - src/**
    - app/**
    - tests/**
  exclude:
    - .git/**
    - node_modules/**
    - .venv/**
    - dist/**
    - build/**

analysis:
  languages:
    - python
  test_paths:
    - tests/**
  documentation_paths:
    - docs/**
  contract_paths:
    - openapi.yaml
    - docs/contracts/**
```

### RF-003 — Exibir status do projeto

O sistema deve fornecer um comando e uma tela para exibir:

- Repositório analisado.
- Branch e commit Git indexados.
- Data e hora da última indexação.
- Linguagens detectadas.
- Quantidade de arquivos, símbolos, rotas, testes e histórias.
- Histórias verificadas, em rascunho, desatualizadas e com erro.
- Avisos de indexação incompleta.

```bash
storycode status
```

---

## 6. Indexação e análise

### RF-004 — Indexar estrutura do repositório

O sistema deve indexar arquivos e diretórios elegíveis, respeitando regras de inclusão e exclusão.

Para cada arquivo indexado, o sistema deve registrar no mínimo:

- Caminho relativo.
- Tipo de arquivo.
- Linguagem detectada.
- Hash do conteúdo.
- Data de modificação.
- Estado de indexação.
- Relações de importação, quando aplicável.

### RF-005 — Extrair símbolos do código

O sistema deve extrair símbolos relevantes de projetos Python, incluindo:

- Módulos.
- Classes.
- Funções.
- Métodos.
- Decorators.
- Imports.
- Chamadas de função, quando identificáveis estaticamente.
- Herança e implementação, quando identificáveis.
- Linhas inicial e final do símbolo.

Cada símbolo deve possuir um identificador estável enquanto seu contexto semântico permanecer reconhecível.

### RF-006 — Detectar pontos de entrada

O sistema deve detectar e classificar pontos de entrada do sistema, incluindo:

- Rotas HTTP de FastAPI.
- Rotas HTTP de Flask.
- Rotas HTTP de Django, quando possível.
- Comandos CLI.
- Tarefas agendadas.
- Consumers e producers de mensagens.
- Webhooks.
- Workers assíncronos, como Celery, quando possível.
- Scripts executáveis Python.

Cada ponto de entrada deve conter, quando disponível:

- Tipo.
- Nome.
- Método HTTP, rota ou tópico.
- Símbolo de origem.
- Arquivo e localização.
- Framework detectado.
- Nível de confiança da detecção.

### RF-007 — Construir relações de execução

O sistema deve criar relações navegáveis entre símbolos, incluindo:

- Importa.
- Chama.
- Retorna.
- Lê.
- Escreve.
- Publica evento.
- Consome evento.
- Usa contrato.
- É exercitado por teste.
- Pertence ao mesmo módulo ou componente.

Quando uma relação for inferida e não comprovada diretamente, ela deve indicar seu nível de confiança.

### RF-008 — Indexar evidências complementares

O sistema deve indexar, quando presentes:

- Testes automatizados.
- Arquivos Markdown.
- ADRs.
- OpenAPI e outros contratos suportados.
- Migrations e schemas de banco.
- Histórico Git.
- Commits relacionados a arquivos e símbolos.

A indexação deve preservar a origem e a localização de cada evidência.

### RF-009 — Indexação incremental

O sistema deve identificar arquivos alterados desde a última indexação e reprocessar apenas o necessário.

```bash
storycode index
storycode index --full
```

O sistema deve informar se precisou fazer uma indexação completa e o motivo.

---

## 7. Histórias

### RF-010 — Criar história manualmente

O sistema deve permitir criar uma história manualmente por CLI, arquivo ou interface web.

```bash
storycode story create checkout-payment
```

Uma história deve suportar:

- Identificador único.
- Título.
- Descrição curta.
- Intenção.
- Gatilho.
- Atores.
- Cenas.
- Resultado.
- Caminhos alternativos.
- Falhas.
- Invariantes.
- Evidências.
- Status.
- Metadados de autoria e atualização.

### RF-011 — Persistir histórias no repositório

As histórias devem ser armazenadas como arquivos legíveis e revisáveis por Git.

Local padrão:

```text
.storycode/stories/
```

Formato inicial obrigatório:

```yaml
version: 1
id: answer-with-rag
title: Responder uma pergunta com contexto recuperado
status: draft
intent: Entregar uma resposta baseada em documentos autorizados.

trigger:
  type: http
  method: POST
  path: /v1/chat

actors:
  - id: user
    type: human
    label: Usuário
  - id: api
    type: service
    label: Chat API
  - id: vector-store
    type: database
    label: Vector Store

scenes:
  - id: validate-request
    type: action
    title: Validar solicitação
    narration: A API valida autenticação, payload e escopo do tenant.
    from: user
    to: api
    evidence:
      - type: symbol
        ref: src/api/chat.py::create_chat

  - id: retrieve-context
    type: action
    title: Recuperar contexto autorizado
    narration: O serviço busca conteúdo relevante dentro do escopo permitido.
    from: api
    to: vector-store
    evidence:
      - type: symbol
        ref: src/services/retrieval.py::retrieve
      - type: test
        ref: tests/integration/test_chat.py::test_filters_by_tenant

outcome: A API devolve uma resposta e as fontes utilizadas.

invariants:
  - Dados de outro tenant não podem ser usados como contexto.

tags:
  - rag
  - chat
  - security
```

### RF-012 — Descobrir histórias automaticamente

O sistema deve gerar histórias em rascunho a partir de pontos de entrada detectados.

```bash
storycode discover
storycode discover --type http
storycode discover --entry POST:/v1/chat
```

A descoberta deve:

- Criar uma história em rascunho para cada fluxo identificado.
- Identificar gatilho, símbolo inicial e chamadas alcançáveis.
- Sugerir atores a partir de componentes internos e externos.
- Criar cenas iniciais a partir das relações de execução.
- Anexar evidências encontradas.
- Informar nível de confiança por cena.
- Evitar sobrescrever histórias editadas manualmente sem confirmação explícita.

### RF-013 — Editar uma história

O sistema deve permitir editar histórias por:

- Arquivo YAML.
- Interface web.
- CLI, para alterações simples.

A edição pela interface deve atualizar o arquivo da história de forma determinística.

### RF-014 — Definir status da história

Uma história deve possuir um dos seguintes estados:

| Status | Significado |
|---|---|
| `draft` | Descoberta ou criada, mas ainda não revisada |
| `review` | Em revisão por pessoa responsável |
| `verified` | Evidências atuais foram verificadas |
| `stale` | Evidências foram alteradas desde a última verificação |
| `broken` | Evidências não puderam ser localizadas ou validadas |
| `archived` | História mantida apenas para referência histórica |

O sistema deve impedir que uma história seja marcada como `verified` quando possuir evidências quebradas.

### RF-015 — Associar histórias

O sistema deve permitir relacionar histórias entre si como:

- Continua.
- Depende de.
- É alternativa de.
- É chamada por.
- É pré-requisito de.
- Substitui.
- Foi substituída por.

Isso deve permitir representar jornadas grandes sem transformar uma única história em um fluxo incompreensível.

---

## 8. Player visual no browser

### RF-016 — Executar servidor local

O sistema deve disponibilizar uma aplicação web local.

```bash
storycode serve
storycode serve --port 7331
```

Ao iniciar, o sistema deve:

- Exibir a URL local no terminal.
- Abrir o navegador quando configurado.
- Disponibilizar API local para a interface.
- Não expor o servidor para a rede externa por padrão.
- Permitir encerrar o servidor de forma segura.

### RF-017 — Listar e buscar histórias

A tela inicial deve permitir:

- Visualizar histórias disponíveis.
- Buscar por título, intenção, tag, gatilho, rota, arquivo, símbolo e ator.
- Filtrar por status.
- Filtrar por tipo de gatilho.
- Filtrar por framework ou componente.
- Ordenar por atualização, criticidade, número de evidências e status.

### RF-018 — Reproduzir uma história

A interface deve apresentar uma história como uma sequência visual de cenas.

A reprodução deve conter:

- Título, intenção e gatilho da história.
- Linha do tempo com progresso atual.
- Atores envolvidos.
- Cena atual com título e narração.
- Movimento visual entre atores.
- Destaque do ator de origem e destino.
- Controles de avançar, voltar, pausar, reiniciar e navegar para uma cena específica.
- Indicação de caminho feliz, falha ou caminho alternativo.
- Estado visual para cena verificada, inferida, desatualizada ou quebrada.

O usuário deve conseguir navegar sem reprodução automática.

### RF-019 — Representar tipos de cena

A interface deve renderizar visualmente os seguintes tipos de cena:

| Tipo | Representação esperada |
|---|---|
| Ação | Movimento entre atores ou execução em um componente |
| Decisão | Bifurcação com condição explícita |
| Leitura | Fluxo de consulta para fonte de dados |
| Escrita | Fluxo de persistência ou atualização |
| Evento publicado | Emissão para fila, tópico ou broker |
| Evento consumido | Consumo a partir de fila, tópico ou broker |
| Falha | Interrupção, exceção, timeout ou rejeição |
| Retry | Repetição controlada da operação |
| Compensação | Ação corretiva após falha parcial |
| Resultado | Entrega de resposta ou conclusão da jornada |

### RF-020 — Exibir mapa do sistema

A interface deve oferecer uma visão complementar de mapa do sistema.

O mapa deve permitir:

- Visualizar atores e componentes relacionados à história atual.
- Alternar entre visão simplificada e detalhada.
- Fazer zoom, pan e centralização.
- Destacar somente os elementos da cena atual.
- Mostrar conexões de entrada e saída de um componente.
- Navegar de um componente para histórias relacionadas.
- Ocultar elementos não relacionados ao contexto atual.

### RF-021 — Exibir painel de evidências

Cada cena deve permitir abrir um painel de evidências.

O painel deve exibir:

- Tipo de evidência.
- Arquivo de origem.
- Símbolo referenciado.
- Intervalo de linhas.
- Trecho de código ou documento.
- Estado de verificação.
- Hash ou versão indexada.
- Relações com testes, contratos, commits e ADRs, quando disponíveis.

A interface deve permitir abrir uma evidência no editor local quando esta integração estiver configurada.

### RF-022 — Exibir caminhos alternativos e falhas

A interface deve permitir alternar entre:

- Caminho feliz.
- Caminhos alternativos.
- Falhas conhecidas.
- Retries e compensações.

A seleção deve atualizar timeline, mapa, texto narrativo e evidências exibidas.

### RF-023 — Exibir invariantes e efeitos

A interface deve apresentar, por história e por cena quando aplicável:

- Invariantes.
- Dados lidos.
- Dados gravados.
- Eventos publicados.
- Serviços externos chamados.
- Efeitos colaterais.
- Riscos ou avisos associados.

---

## 9. Navegação de código e impacto

### RF-024 — Navegar da história para o código

A pessoa usuária deve conseguir navegar de uma cena até:

- Arquivo.
- Símbolo.
- Linha de código.
- Teste relacionado.
- Contrato.
- Migration.
- Documento.
- Commit.

A navegação deve manter contexto da história e permitir retorno ao player visual.

### RF-025 — Navegar do código para histórias

A interface e a CLI devem permitir descobrir quais histórias referenciam um arquivo ou símbolo.

```bash
storycode impact src/services/retrieval.py
storycode impact src/services/retrieval.py::retrieve
```

O resultado deve informar:

- Histórias afetadas.
- Cenas afetadas.
- Tipo de relação.
- Estado de verificação.
- Testes e invariantes potencialmente relacionados.

### RF-026 — Exibir impacto de mudança

O sistema deve permitir comparar um conjunto de arquivos modificados com o índice atual e identificar histórias potencialmente afetadas.

```bash
storycode impact --git-diff
storycode impact --base main
```

O relatório deve diferenciar:

- História diretamente afetada.
- História indiretamente afetada.
- Evidência modificada.
- Evidência removida.
- Relação inferida que requer revisão humana.

---

## 10. Verificação e documentação viva

### RF-027 — Verificar histórias

O sistema deve verificar se uma história permanece vinculada às evidências atuais.

```bash
storycode verify
storycode verify rag-answer
```

A verificação deve identificar no mínimo:

- Arquivos removidos.
- Símbolos removidos ou renomeados.
- Linhas ou trechos alterados.
- Testes removidos ou renomeados.
- Contratos incompatíveis.
- Gatilhos que não podem mais ser detectados.
- Relações do fluxo que deixaram de ser encontradas.
- Evidências sem acesso ou impossíveis de analisar.

### RF-028 — Classificar resultado de verificação

A verificação de cada evidência deve resultar em um estado explícito:

| Estado | Significado |
|---|---|
| `verified` | Evidência atual corresponde ao esperado |
| `changed` | Evidência existe, mas seu conteúdo mudou |
| `missing` | Evidência não foi encontrada |
| `ambiguous` | Mais de uma evidência possível foi encontrada |
| `unavailable` | Evidência não pôde ser analisada |
| `inferred` | Relação estimada, sem comprovação determinística |

O status agregado da história deve ser calculado com base no estado de suas evidências.

### RF-029 — Mostrar drift visualmente

A interface deve tornar evidente quando uma história estiver desatualizada.

Ela deve informar:

- Qual cena está afetada.
- Qual evidência mudou.
- Em qual commit ou indexação a alteração foi detectada.
- Qual é o impacto conhecido.
- Ação recomendada: revisar, atualizar, remover ou arquivar.

### RF-030 — Gerar relatório de verificação

O sistema deve gerar relatório legível em terminal e arquivo.

```bash
storycode verify --format markdown
storycode verify --format json
```

O relatório deve incluir resumo, histórias afetadas, evidências com problema e sugestões de revisão.

---

## 11. Git e decisões

### RF-031 — Relacionar evidências ao histórico Git

Quando Git estiver disponível, o sistema deve permitir visualizar:

- Último commit que alterou um arquivo ou símbolo.
- Commits relacionados a uma história.
- Autores e datas.
- Mensagens de commit.
- Alterações relevantes para uma cena.

O sistema deve tratar Git como evidência contextual, não como fonte definitiva de intenção.

### RF-032 — Indexar ADRs e documentos

O sistema deve reconhecer documentos de arquitetura e decisão, principalmente arquivos Markdown em diretórios configurados.

O sistema deve permitir associar um ADR ou documento a:

- Uma história.
- Uma cena.
- Um componente.
- Um invariante.
- Uma integração externa.

### RF-033 — Exibir linha do tempo

A interface deve permitir visualizar uma linha do tempo de mudanças para uma história, incluindo:

- Criação da história.
- Mudanças em suas evidências.
- Alterações de status.
- Commits relacionados.
- ADRs associados.
- Arquivamento ou substituição da história.

---

## 12. Importação e exportação

### RF-034 — Exportar história

O sistema deve exportar histórias em:

- Markdown.
- JSON.
- Mermaid.
- HTML estático.

```bash
storycode export rag-answer --format markdown
storycode export rag-answer --format mermaid
storycode export rag-answer --format html
```

A exportação deve conter narrativa, fluxo, atores, invariantes e evidências.

### RF-035 — Gerar diagrama Mermaid

O sistema deve gerar Mermaid para:

- Fluxo de sequência.
- Fluxo de decisão.
- Mapa simplificado de componentes.

A exportação deve identificar elementos inferidos ou com evidência quebrada.

### RF-036 — Importar histórias

O sistema deve permitir importar histórias a partir de:

- Arquivos YAML compatíveis.
- JSON compatível.
- Diretório configurado de histórias.
- Templates internos.

Importações inválidas devem retornar erros com linha, campo e motivo.

---

## 13. Integração com IA opcional

### RF-037 — Manter funcionamento sem IA

Todas as funcionalidades essenciais do MVP devem funcionar sem provider de IA configurado.

A ausência de IA não pode impedir:

- Indexação.
- Descoberta básica.
- Criação manual.
- Reprodução visual.
- Evidências.
- Verificação.
- Exportação.

### RF-038 — Usar IA apenas como assistência

Quando habilitada, a IA pode:

- Sugerir título para história.
- Sugerir intenção.
- Reescrever narrativa em linguagem mais clara.
- Sugerir nomes de cenas.
- Agrupar cenas técnicas em capítulos.
- Sugerir perguntas para onboarding.
- Resumir mudanças detectadas.

A IA não pode criar automaticamente uma evidência com status `verified`.

Toda informação gerada por IA sem referência determinística deve ser marcada como `inferred` ou `suggested`.

### RF-039 — Suportar provider configurável

O sistema deve permitir configurar providers compatíveis com:

- Ollama.
- APIs OpenAI-compatible.
- Outros providers via plugin futuro.

A configuração deve permitir endpoint, modelo, credenciais por variável de ambiente e modo de privacidade.

---

## 14. Interface de linha de comando

### RF-040 — Disponibilizar comandos principais

O MVP deve disponibilizar os seguintes comandos:

```bash
storycode init
storycode status
storycode index
storycode discover
storycode serve
storycode story create <id>
storycode story list
storycode story show <id>
storycode tell <id>
storycode verify [id]
storycode impact <path-or-symbol>
storycode export <id>
```

Cada comando deve fornecer:

- Ajuda com `--help`.
- Mensagens de erro acionáveis.
- Código de saída diferente de zero em falhas.
- Saída legível para humanos.
- Opção JSON quando aplicável.

### RF-041 — Narrar no terminal

O comando `tell` deve apresentar uma versão textual resumida da história.

```bash
storycode tell answer-with-rag
```

A saída deve incluir:

- Título.
- Intenção.
- Gatilho.
- Atores.
- Cenas em ordem.
- Evidências principais.
- Invariantes.
- Estado de verificação.

---

## 15. Critérios de aceite do MVP

O MVP será considerado funcional quando for possível executar o seguinte cenário sem configuração de serviços externos:

1. Clonar ou abrir um projeto Python local contendo uma aplicação FastAPI simples.
2. Executar `storycode init`.
3. Executar `storycode index`.
4. Detectar ao menos uma rota HTTP e seu handler.
5. Executar `storycode discover --type http`.
6. Gerar uma história em rascunho baseada em uma rota detectada.
7. Editar a história para incluir intenção, narrativa e evidências.
8. Executar `storycode serve`.
9. Abrir a história em um browser local.
10. Navegar cena a cena pelo fluxo.
11. Abrir uma evidência e visualizar o trecho de código associado.
12. Alterar ou remover um símbolo referenciado.
13. Executar `storycode verify`.
14. Visualizar a história como `stale` ou `broken`.
15. Exportar a história em Markdown e Mermaid.

---

## 16. Fora do escopo inicial

Os seguintes itens não fazem parte do MVP:

- Análise completa de todas as linguagens e frameworks.
- Execução remota, SaaS multi-tenant ou autenticação de usuários.
- Colaboração em tempo real.
- Edição simultânea de histórias.
- Integração obrigatória com GitHub, GitLab, Jira, Linear ou Notion.
- Captura obrigatória de traces em produção.
- Geração de vídeo, GIF ou apresentação automática.
- Análise de segurança, SAST ou qualidade de código como objetivo principal.
- Substituição de ferramentas de observabilidade, APM ou documentação de API.
- Alteração automática de código com base em uma história.