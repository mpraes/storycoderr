# Requisitos Não Funcionais — StoryCode

## 1. Objetivo

O StoryCode deve ser uma ferramenta local-first, leve e simples de instalar, capaz de indexar e apresentar histórias de código no navegador sem exigir infraestrutura externa.

O produto deve funcionar de forma consistente em Windows e Linux, priorizando execução via CLI, arquivos locais e um único processo executável sempre que possível.

---

## 2. Princípios de qualidade

### RNF-001 — Local-first

O StoryCode deve funcionar integralmente no computador da pessoa usuária.

A execução padrão não deve exigir:

- Conta de usuário.
- Serviço SaaS.
- Banco de dados remoto.
- Container Docker.
- Kubernetes.
- Node.js instalado globalmente.
- Python instalado globalmente.
- Chave de API.
- Conexão com a internet.

O sistema pode oferecer integrações opcionais com IA, Git hosting ou observabilidade, mas elas devem permanecer desabilitadas por padrão.

### RNF-002 — Privacidade por padrão

O StoryCode não deve transmitir código-fonte, metadados do repositório, histórias, evidências, commits ou arquivos de configuração para serviços externos sem ação explícita da pessoa usuária.

Quando uma integração externa for habilitada, o sistema deve informar:

- Qual provider será utilizado.
- Quais dados poderão ser enviados.
- Qual configuração ativou a integração.
- Como desabilitá-la.

### RNF-003 — Baixo atrito

A experiência mínima desejada deve ser:

```bash
storycode init
storycode index
storycode serve
```

A instalação deve ser possível a partir de um binário adequado à plataforma, sem etapas de compilação ou dependências de runtime.

O sistema deve mostrar mensagens claras de progresso, sucesso, erro e próxima ação recomendada.

### RNF-004 — Um binário distribuível

A distribuição oficial do MVP deve disponibilizar binários autocontidos para:

- Windows x86_64.
- Linux x86_64.
- Linux ARM64, quando viável.

O binário deve conter CLI, servidor HTTP local, assets estáticos da interface web e mecanismos de migração do índice local.

A pessoa usuária não deve precisar instalar Go, Node.js, Python, Java ou Docker para utilizar as funções principais.

---

## 3. Compatibilidade

### RNF-005 — Sistemas operacionais suportados

O MVP deve suportar:

| Plataforma | Arquitetura | Nível de suporte |
|---|---|---|
| Windows 10 e 11 | x86_64 | Obrigatório |
| Linux moderno | x86_64 | Obrigatório |
| Linux moderno | ARM64 | Desejável |
| WSL2 | x86_64 e ARM64 | Desejável |
| macOS | Apple Silicon e Intel | Futuro, não obrigatório no MVP |

O sistema deve detectar e informar claramente quando estiver em plataforma não suportada.

### RNF-006 — Shells suportados

Os comandos documentados devem funcionar em:

- PowerShell 7+.
- Windows Command Prompt.
- Bash.
- Zsh.
- Fish, quando não houver sintaxe específica de shell.

A documentação não deve depender de comandos exclusivos de Unix sem fornecer alternativa para Windows.

### RNF-007 — Caminhos e sistema de arquivos

O sistema deve tratar caminhos de arquivo de maneira independente de plataforma.

Ele deve:

- Usar APIs nativas de manipulação de caminhos.
- Suportar separadores `\` e `/`.
- Suportar caminhos com espaços.
- Suportar caracteres Unicode em caminhos e conteúdo.
- Evitar assumir que nomes de arquivo diferenciam maiúsculas de minúsculas.
- Detectar e informar colisões em sistemas case-insensitive.
- Lidar corretamente com line endings `LF` e `CRLF`.
- Ignorar links simbólicos por padrão ou processá-los somente com opção explícita.
- Evitar loops causados por links simbólicos.

### RNF-008 — Permissões

O sistema deve funcionar em diretórios onde a pessoa usuária possua apenas permissões de leitura para o código-fonte.

Por padrão, o StoryCode deve escrever seus dados apenas em:

```text
<repository>/.storycode/
```

Quando não houver permissão de escrita no repositório, o sistema deve oferecer armazenamento alternativo no diretório local de dados do usuário, informando claramente onde os dados serão persistidos.

---

## 4. Instalação e atualização

### RNF-009 — Métodos de instalação

O projeto deve fornecer, no mínimo:

- Download direto de binário compactado.
- Checksums SHA-256.
- Instruções de instalação para Windows e Linux.
- Comando de verificação de versão.

```bash
storycode --version
```

Métodos adicionais, como `winget`, Scoop, Chocolatey, Homebrew, AUR, `apt`, `dnf` ou `npm`, podem ser adicionados posteriormente, mas não devem ser dependência do MVP.

### RNF-010 — Atualização previsível

O sistema deve permitir atualização manual segura por substituição do binário.

Se um comando de autoatualização for disponibilizado futuramente, ele deve:

- Exigir confirmação explícita.
- Mostrar versão atual e versão alvo.
- Validar checksum ou assinatura do artefato.
- Permitir desabilitar verificação automática de atualizações.
- Nunca atualizar silenciosamente.

### RNF-011 — Sem privilégios administrativos

A instalação e a execução devem funcionar sem privilégios de administrador ou root.

O sistema não deve:

- Alterar variáveis globais do sistema automaticamente.
- Criar serviços de inicialização.
- Instalar drivers.
- Abrir portas externas.
- Modificar firewall.
- Solicitar elevação de privilégio.

---

## 5. Desempenho e consumo

### RNF-012 — Inicialização da CLI

Em um computador de desenvolvimento comum, comandos informativos devem iniciar em até 1 segundo, exceto quando houver leitura inevitável de disco.

Exemplos:

```bash
storycode --version
storycode --help
storycode status
storycode story list
```

### RNF-013 — Servidor local

O comando `storycode serve` deve iniciar o servidor HTTP local e disponibilizar a interface web em até 3 segundos para um índice já existente.

O servidor deve escutar somente em loopback por padrão:

```text
127.0.0.1
localhost
```

Ele não deve expor a interface em `0.0.0.0` sem uma opção explícita.

### RNF-014 — Indexação inicial

Para um repositório Python de até 10.000 arquivos de código e documentação, em hardware comum de desenvolvimento, o sistema deve:

- Exibir progresso contínuo.
- Permitir cancelamento seguro.
- Persistir progresso suficiente para retomar ou reiniciar sem corromper o índice.
- Concluir a primeira indexação em tempo aceitável para uso interativo.

A meta inicial recomendada é concluir em até 5 minutos em SSD e processador moderno, sem execução de testes e sem uso de IA.

### RNF-015 — Indexação incremental

Após uma indexação concluída, o sistema deve processar apenas arquivos modificados, criados ou removidos.

Para mudanças locais em até 50 arquivos, a meta recomendada é concluir a reindexação incremental em até 15 segundos, em repositórios de tamanho moderado.

### RNF-016 — Consumo de memória

Durante comandos de consulta, visualização ou verificação, o StoryCode deve manter consumo moderado de memória e evitar carregar todo o repositório simultaneamente.

Para o MVP, a meta recomendada é:

| Operação | Meta de memória |
|---|---:|
| CLI informativa | Até 100 MB |
| Servidor local com índice carregado | Até 300 MB |
| Indexação de repositório moderado | Até 1 GB |
| Processos temporários | Devem ser liberados ao final da operação |

O sistema deve processar arquivos em streaming ou lotes quando isso reduzir pressão de memória.

### RNF-017 — Consumo de disco

O índice local deve ser proporcional ao conteúdo analisado e não deve duplicar desnecessariamente o repositório.

O StoryCode não deve armazenar o conteúdo completo de todos os arquivos por padrão quando apenas metadados, hashes, símbolos e referências forem suficientes.

Caches devem ter:

- Local conhecido.
- Tamanho observável.
- Comando de limpeza.
- Política de invalidação clara.

```bash
storycode cache status
storycode cache clear
```

---

## 6. Confiabilidade e integridade

### RNF-018 — Operações idempotentes

Os comandos abaixo devem ser seguros para repetição:

```bash
storycode init
storycode index
storycode discover
storycode verify
storycode serve
```

Reexecutar um comando não deve duplicar histórias, símbolos, evidências ou entradas de índice.

### RNF-019 — Segurança contra corrupção

O sistema deve manter consistência do índice mesmo em caso de:

- Cancelamento manual.
- Queda do processo.
- Falta de energia.
- Falta de espaço em disco.
- Arquivos mudando durante a indexação.
- Erro de parsing em arquivo individual.

A escrita do índice deve usar transações ou estratégia equivalente de escrita atômica.

O sistema deve preservar o último índice válido até que uma nova indexação seja concluída com sucesso.

### RNF-020 — Tolerância a erros de análise

Um arquivo inválido, ilegível ou incompatível não deve interromper toda a indexação.

O sistema deve:

- Registrar o arquivo com erro.
- Informar o motivo.
- Continuar com os demais arquivos, quando seguro.
- Exibir aviso no status e na interface.
- Permitir reprocessamento após correção.

### RNF-021 — Recuperação do índice

O sistema deve fornecer mecanismos para diagnosticar e reconstruir dados locais.

```bash
storycode doctor
storycode index --full
storycode cache clear
```

O comando `doctor` deve validar:

- Arquivos de configuração.
- Versão do índice.
- Integridade da base local.
- Permissões de leitura e escrita.
- Disponibilidade de Git, quando utilizado.
- Disponibilidade da porta do servidor local.
- Espaço em disco quando insuficiente.

### RNF-022 — Migrações compatíveis

Mudanças no schema do índice devem ser versionadas.

Ao detectar uma versão antiga, o sistema deve:

- Informar que existe migração necessária.
- Fazer backup ou manter possibilidade de reconstrução.
- Executar migração de maneira atômica.
- Solicitar confirmação apenas quando a operação puder ser destrutiva ou demorada.
- Oferecer opção de reconstrução completa.

---

## 7. Segurança

### RNF-023 — Execução não invasiva

O StoryCode deve analisar código sem executá-lo.

Por padrão, ele não deve:

- Importar módulos do projeto em runtime.
- Executar scripts.
- Rodar migrations.
- Iniciar aplicações.
- Rodar testes automaticamente.
- Fazer requisições HTTP.
- Conectar-se a bancos de dados, filas ou APIs externas.
- Avaliar conteúdo de configuração como código.

A análise deve ser baseada em leitura estática, parsing e metadados.

### RNF-024 — Proteção contra conteúdo malicioso

O sistema deve tratar repositórios como conteúdo não confiável.

Ele deve:

- Limitar tamanho de arquivos processados.
- Evitar seguir links simbólicos sem autorização.
- Não executar hooks Git.
- Não interpretar Markdown, HTML ou código como instrução operacional.
- Sanitizar conteúdo antes de renderizá-lo na interface web.
- Escapar conteúdo de código e documentação exibido no browser.
- Mitigar XSS em títulos, narrativas, comentários e arquivos Markdown.

### RNF-025 — Servidor HTTP local seguro

O servidor local deve:

- Escutar em loopback por padrão.
- Usar uma porta configurável.
- Escolher uma porta livre quando a padrão estiver ocupada, se configurado.
- Informar a URL exata no terminal.
- Aplicar headers de segurança apropriados.
- Não disponibilizar operações de escrita sem proteção contra solicitações cruzadas.
- Rejeitar origens externas não autorizadas.
- Encerrar conexões ao finalizar o processo.

Quando exposto explicitamente na rede, o sistema deve emitir aviso de segurança claro e requerer uma opção de consentimento explícito.

### RNF-026 — Segredos e credenciais

O StoryCode não deve indexar, exibir ou exportar segredos conhecidos intencionalmente.

O sistema deve oferecer regras de exclusão para arquivos como:

```text
.env
.env.*
*.pem
*.key
id_rsa
credentials.json
secrets.*
```

A detecção de possíveis segredos deve priorizar ocultação na interface e exportações, exibindo um placeholder em vez do valor bruto.

Credenciais de integrações opcionais devem ser obtidas preferencialmente por variáveis de ambiente ou mecanismos seguros da plataforma, nunca gravadas em texto aberto no repositório.

---

## 8. Usabilidade e experiência de desenvolvimento

### RNF-027 — Mensagens acionáveis

Mensagens de erro devem explicar:

- O que falhou.
- Onde falhou.
- Por que pode ter falhado.
- O que a pessoa usuária pode fazer em seguida.

Exemplo esperado:

```text
Erro: não foi possível escrever em .storycode/index.

Causa provável: o diretório do repositório é somente leitura.

Próximas ações:
  1. Execute o comando em um clone com permissão de escrita.
  2. Use: storycode config set storage.mode user
  3. Consulte: storycode doctor
```

### RNF-028 — Progresso transparente

Operações que excedam 1 segundo devem exibir progresso.

A indexação deve informar, quando disponível:

- Fase atual.
- Arquivos processados e total estimado.
- Linguagem ou analisador em execução.
- Quantidade de símbolos e relações encontradas.
- Avisos encontrados.
- Tempo decorrido.
- Possibilidade de cancelamento.

A saída deve ser legível em terminais simples e aprimorada em terminais com suporte a cores.

### RNF-029 — Acessibilidade da interface web

A interface web deve seguir boas práticas de acessibilidade.

Ela deve oferecer:

- Navegação por teclado.
- Foco visível.
- Textos alternativos ou descrições para elementos visuais.
- Contraste suficiente.
- Não depender exclusivamente de cor para comunicar estado.
- Preferência de redução de movimento.
- Controle para reduzir ou desabilitar animações.
- Suporte a leitores de tela nos fluxos principais.

### RNF-030 — Responsividade visual

A interface deve funcionar em telas a partir de 1280 px de largura no MVP.

Ela deve permanecer utilizável em resoluções menores, priorizando:

- Leitura da narrativa.
- Navegação de cenas.
- Painel de evidências.
- Busca de histórias.

Visualizações grandes devem adaptar-se sem exigir que o usuário arraste horizontalmente toda a página.

### RNF-031 — Internacionalização

O MVP deve usar textos de interface centralizados e preparados para tradução.

O idioma inicial pode ser inglês, mas o sistema deve permitir adicionar traduções sem alterar componentes de interface.

A documentação oficial deve incluir ao menos inglês. Traduções em português podem ser mantidas pela comunidade ou como idioma adicional.

---

## 9. Observabilidade local

### RNF-032 — Logs locais

O sistema deve registrar logs locais estruturados para diagnóstico, sem incluir conteúdo sensível desnecessário.

Os logs devem permitir níveis:

```text
error
warn
info
debug
trace
```

A pessoa usuária deve poder controlar o nível por flag ou variável de ambiente:

```bash
storycode index --log-level debug
STORYCODE_LOG_LEVEL=debug storycode index
```

### RNF-033 — Sem telemetria por padrão

O StoryCode não deve enviar telemetria, métricas, logs, identificação de repositório ou informações de uso para servidores externos por padrão.

Caso telemetria seja adicionada no futuro, ela deve ser:

- Explicitamente opt-in.
- Documentada.
- Anonimizada quando possível.
- Fácil de desabilitar.
- Visível na configuração.

### RNF-034 — Diagnóstico compartilhável

O comando `storycode doctor` deve poder gerar um relatório seguro para suporte.

```bash
storycode doctor --report
```

O relatório não deve incluir:

- Conteúdo do código.
- Tokens.
- Senhas.
- Valores de variáveis de ambiente.
- Caminhos completos do usuário, salvo autorização explícita.
- Dados de histórias ou evidências, salvo opção explícita.

---

## 10. Manutenibilidade e extensibilidade

### RNF-035 — Arquitetura modular

O projeto deve separar, no mínimo:

- CLI.
- Configuração.
- Descoberta de repositório.
- Análise estática.
- Armazenamento do índice.
- Modelo de histórias.
- Verificação de evidências.
- API HTTP local.
- Interface web.
- Exportadores.
- Integrações opcionais.

Uma falha em integração opcional não deve impedir funções centrais.

### RNF-036 — Plugins e analisadores

O núcleo deve permitir adicionar suporte a novas linguagens, frameworks, fontes de evidência e exportadores sem modificar a lógica principal de histórias.

A primeira versão pode manter plugins compilados no binário, mas suas interfaces devem ser definidas desde o início.

### RNF-037 — Estabilidade de formatos

Os formatos de configuração, história e exportação devem possuir:

- Campo de versão.
- Schema documentado.
- Regras de compatibilidade.
- Estratégia de depreciação.
- Migrações explícitas quando necessárias.

O sistema não deve alterar silenciosamente arquivos de histórias criados manualmente.

### RNF-038 — Qualidade de código

O projeto deve possuir:

- Formatação automática.
- Linting.
- Testes unitários.
- Testes de integração para a CLI.
- Testes de compatibilidade de fixtures.
- Testes end-to-end da interface web.
- Análise de dependências.
- Revisão de pull request antes de merge na branch principal.

---

## 11. Open source e distribuição

### RNF-039 — Licença

O projeto deve adotar uma licença open source permissiva, preferencialmente Apache License 2.0 ou MIT.

A escolha deve ser registrada em `LICENSE` e explicada em `README.md`.

### RNF-040 — Repositório público compreensível

O repositório deve conter:

```text
README.md
LICENSE
CONTRIBUTING.md
CODE_OF_CONDUCT.md
SECURITY.md
CHANGELOG.md
docs/
examples/
```

O `README.md` deve explicar propósito, instalação, primeiros comandos, limitações atuais, arquitetura e como contribuir.

### RNF-041 — Builds reproduzíveis

Os releases devem ser gerados por pipeline automatizado e versionado.

Cada release deve publicar:

- Binários por plataforma.
- Checksums SHA-256.
- Notas de versão.
- Versão semântica.
- Informações de compatibilidade.
- Artefatos da interface web incorporados ou versionados junto ao binário.

### RNF-042 — Dependências auditáveis

O projeto deve manter dependências mínimas e auditáveis.

Toda dependência adicionada deve ter justificativa técnica, licença compatível e manutenção ativa.

Dependências que exigem serviços externos não devem ser obrigatórias para o núcleo do produto.

---

## 12. CHANGELOG

### RNF-043 — Manter CHANGELOG.md

O projeto deve manter um arquivo `CHANGELOG.md` na raiz do repositório.

O changelog deve seguir o formato [Keep a Changelog](https://keepachangelog.com/) e usar [Semantic Versioning](https://semver.org/).

Cada release deve documentar, quando aplicável:

- `Added`: funcionalidades novas.
- `Changed`: mudanças em comportamento existente.
- `Deprecated`: funcionalidades em processo de remoção.
- `Removed`: funcionalidades removidas.
- `Fixed`: correções.
- `Security`: mudanças relacionadas à segurança.

### RNF-044 — Registrar mudanças incompatíveis

Mudanças que alterem comportamento, formato de histórias, configuração, comandos CLI, API local ou formato do índice devem ser classificadas como incompatíveis quando apropriado.

Mudanças incompatíveis devem:

- Ser explicadas claramente no `CHANGELOG.md`.
- Incluir instruções de migração.
- Indicar versões afetadas.
- Ser refletidas no versionamento semântico.
- Mostrar aviso em runtime quando a migração for necessária.

### RNF-045 — Estrutura inicial do changelog

O arquivo inicial deve seguir esta estrutura:

```md
# Changelog

Todas as mudanças relevantes deste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/),
e este projeto segue [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Estrutura inicial do projeto StoryCode.
- CLI local-first para inicialização, indexação e visualização de histórias.
- Modelo versionado de histórias e evidências.
- Interface web local para reprodução visual de jornadas.

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [0.1.0] - YYYY-MM-DD

### Added
- Primeira versão pública do StoryCode.
```

---

## 13. Critérios de aceite

Os requisitos não funcionais do MVP serão considerados atendidos quando:

1. Uma pessoa usuária no Windows 10/11 ou Linux x86_64 conseguir baixar um binário e executar `storycode --version` sem instalar runtime adicional.
2. O binário conseguir inicializar, indexar e servir um repositório local sem internet, Docker, conta ou chave de API.
3. A interface web abrir somente em `localhost` por padrão.
4. O sistema não executar código analisado, testes, scripts ou hooks Git sem solicitação explícita.
5. Uma indexação interrompida não corromper o último índice válido.
6. Arquivos inválidos ou não suportados gerarem avisos sem interromper todo o processo.
7. O sistema tratar corretamente caminhos Windows, Linux, Unicode, espaços e line endings diferentes.
8. Segredos conhecidos forem excluídos ou mascarados por padrão.
9. A interface puder ser usada por teclado e oferecer opção de reduzir animações.
10. O projeto possuir `CHANGELOG.md`, releases versionados e checksums para os binários distribuídos.