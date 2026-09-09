# Esteira DevSecOps — Central de Segurança Reutilizável

[![Pipeline Base - GitHub Actions](https://github.com/joaohenrique11z/teste_esteira/actions/workflows/pipeline.yml/badge.svg)](https://github.com/joaohenrique11z/teste_esteira/actions/workflows/pipeline.yml)

---

## 1. Sobre o Projeto

Este repositório é a **Central DevSecOps** da equipe — uma esteira de segurança automatizada construída como um **Reusable Workflow** do GitHub Actions. O objetivo é fornecer um pipeline padronizado de análise de segurança que qualquer projeto (corporativo ou acadêmico) pode consumir sem duplicar código ou configurações.

**Princípios:**

- **Centralização**: Um único repositório contém toda a lógica de segurança. Correções e novas ferramentas são propagadas automaticamente para todos os projetos consumidores.
- **Shadow Mode**: Nenhum scanner bloqueia a pipeline do projeto consumidor. Todos os findings são reportados como `::warning::` e exportados para análise posterior — a build **nunca quebra** por causa de uma vulnerabilidade detectada.
- **Multi-Stack**: Suporte nativo para **7 ecossistemas** — `node`, `dotnet`, `python`, `java`, `cpp`, `go` e `rust`.

---

## 2. Arquitetura e Estrutura

### Como funciona

A esteira opera com dois arquivos de workflow distintos:

```
┌─────────────────────────────────────────────────────────────────┐
│  REPOSITÓRIO CENTRAL (joaohenrique11z/teste_esteira)           │
│                                                                 │
│  .github/workflows/pipeline.yml   ← Reusable Workflow          │
│    • Trigger: workflow_call (consumidores) + workflow_dispatch   │
│    • Inputs: project_name, stack_type, run_dast, dast_target_url│
│    • Secrets: DEFECTDOJO_URL, DEFECTDOJO_API_KEY                │
│    • Jobs: setup stack → Gitleaks → Semgrep → Trivy → ZAP      │
│                                                                 │
│  .github/workflows/ci.yml         ← Auto-teste interno         │
│    • Trigger: push/pull_request em main e develop               │
│    • Chama pipeline.yml via workflow_call para validar a esteira│
└─────────────────────────────────────────────────────────────────┘
         ▲
         │  uses: joaohenrique11z/teste_esteira/...pipeline.yml@v1
         │
┌────────┴──────────────────────────────────────┐
│  PROJETO CONSUMIDOR (qualquer repositório)     │
│                                                │
│  .github/workflows/ci.yml                      │
│    • Trigger: push/pull_request                │
│    • Job "security" chama a esteira central    │
│    • Passa project_name, stack_type, etc.      │
└────────────────────────────────────────────────┘
```

### `pipeline.yml` — A Esteira Central

Este é o coração da solução. Definido como um **Reusable Workflow** (`on: workflow_call`), ele recebe parâmetros do projeto consumidor e executa toda a cadeia de scanners de segurança. Também possui trigger `workflow_dispatch` para execução manual diretamente pela interface do GitHub Actions.

**O que ele faz em ordem:**

1. **Checkout duplo** — clona o código do projeto consumidor (`app-code`) e as ferramentas de segurança desta central (`security-tools`).
2. **Setup da stack** — instala o toolchain correto com base no `stack_type` informado (Node 20, .NET 8, Python 3.12, Java 17, C/C++ build-essential, Go stable, Rust stable).
3. **Scanners de segurança** — executa Gitleaks, Semgrep, Trivy e (opcionalmente) OWASP ZAP.
4. **Upload de artefatos** — empacota todos os relatórios JSON em um artefato `devsecops-reports`.

### `ci.yml` — Arquivo do Projeto Consumidor

Cada projeto que deseja consumir a esteira precisa criar um arquivo `.github/workflows/ci.yml` no seu próprio repositório. Este arquivo **não contém lógica de segurança** — ele apenas chama o `pipeline.yml` desta central via `uses:`, passando os parâmetros relevantes.

---

## 3. Ferramentas Integradas de Segurança (Scanners)

Todos os scanners rodam em **Shadow Mode** — utilizam `|| true`, `exit-code: '0'` ou `fail_action: false` para garantir que findings **nunca bloqueiam** a pipeline do consumidor.

### 🔑 Gitleaks — Análise de Segredos e Vazamentos

| Aspecto | Detalhe |
|---------|---------|
| **Tipo** | Secret Scanning |
| **Execução** | Container Docker (`ghcr.io/gitleaks/gitleaks:latest`) |
| **Config** | Regras customizadas em `security/gitleaks/.gitleaks.toml` |
| **Saída** | `reports/gitleaks/gitleaks-results.json` |

Detecta credenciais, tokens de API, chaves privadas e outros segredos que possam ter sido commitados acidentalmente no repositório.

### 🔍 Semgrep — SAST (Análise Estática de Código)

| Aspecto | Detalhe |
|---------|---------|
| **Tipo** | Static Application Security Testing (SAST) |
| **Execução** | Container Docker (`semgrep/semgrep`) |
| **Config** | Regras OWASP Top 10 + regras customizadas em `security/semgrep/rules.yaml` |
| **Saída** | `reports/semgrep/semgrep-results.json` |

Analisa o código-fonte em busca de vulnerabilidades como SQL Injection, XSS, uso de funções inseguras, segredos hardcoded e más práticas de segurança. As regras do Semgrep Registry são selecionadas automaticamente com base no `stack_type`:

| Stack | Registry Semgrep |
|-------|-----------------|
| `node` | `p/nodejs` |
| `python` | `p/python` |
| `java` | `p/java` |
| `dotnet` | `p/csharp` |
| `cpp` | `p/c` |
| `go` | `p/golang` |
| `rust` | `p/rust` |

Além do registry específico, **todas as stacks** recebem as regras `p/owasp-top-ten` e as regras customizadas da central.

### 📦 Trivy — Análise de Vulnerabilidades (SCA)

| Aspecto | Detalhe |
|---------|---------|
| **Tipo** | Software Composition Analysis (SCA) / Filesystem Scan |
| **Execução** | GitHub Action (`aquasecurity/trivy-action@master`) |
| **Escopo** | Scan de filesystem no diretório `app-code` |
| **Severidade** | `HIGH` e `CRITICAL` |
| **Saída** | `reports/trivy/trivy-fs-results.json` |

Verifica dependências do projeto (ex: `package-lock.json`, `go.sum`, `Cargo.lock`, `requirements.txt`) em busca de CVEs conhecidos. Também analisa o filesystem em busca de configurações inseguras.

### 🌐 OWASP ZAP — DAST (Teste Dinâmico de Segurança)

| Aspecto | Detalhe |
|---------|---------|
| **Tipo** | Dynamic Application Security Testing (DAST) |
| **Execução** | GitHub Action (`zaproxy/action-baseline@v0.14.0`) |
| **Condição** | Executa **somente** se `run_dast: true` **e** `dast_target_url` estiver preenchido |
| **Saída** | `reports/zap/report_json.json` |

Realiza varredura ativa contra uma aplicação web em execução, testando vulnerabilidades como XSS, SQL Injection, headers inseguros e mais.

---

## ⚠️ Observação Importante sobre o OWASP ZAP (DAST)

O **OWASP ZAP é uma ferramenta de análise dinâmica (DAST)**. Diferente dos demais scanners que analisam código-fonte ou dependências de forma estática, o ZAP precisa **interagir com uma aplicação web ou API que esteja rodando ativamente** em um endereço HTTP.

**Isso significa que:**

- O parâmetro `dast_target_url` deve apontar para um servidor web real e acessível (ex: `http://localhost:3000`, `https://staging.meu-app.com`).
- O ZAP envia requisições HTTP reais (GET, POST, etc.) ao alvo para testar vulnerabilidades — ele não analisa arquivos.
- **Se o seu projeto não sobe um servidor web** (por exemplo: scripts CLI, bibliotecas, pacotes, ferramentas de linha de comando, notebooks, etc.), a etapa do ZAP **não fará sentido** e pode ser desativada com `run_dast: false`.

**Exemplos práticos:**

| Tipo de Projeto | Usar ZAP? | Configuração |
|----------------|-----------|--------------|
| API REST (Express, Flask, Spring) | ✅ Sim | `run_dast: true` + `dast_target_url: 'http://...'` |
| Aplicação Web (React + Backend) | ✅ Sim | `run_dast: true` + `dast_target_url: 'https://staging...'` |
| Script Python / CLI | ❌ Não | `run_dast: false` |
| Biblioteca / SDK | ❌ Não | `run_dast: false` |
| Job de processamento de dados | ❌ Não | `run_dast: false` |

> **Dica**: mesmo com `run_dast: false`, os demais scanners (Gitleaks, Semgrep, Trivy) continuam executando normalmente e cobrindo o projeto.

---

## 5. Como Usar (Exemplo Prático)

### Passo a passo

1. No repositório do seu projeto, crie o arquivo `.github/workflows/ci.yml`.
2. Configure os **secrets** `DEFECTDOJO_URL` e `DEFECTDOJO_API_KEY` em **Settings → Secrets and variables → Actions** do seu repositório.
3. Cole o template abaixo, ajustando os parâmetros para o seu projeto.

### Template do `ci.yml`

```yaml
name: CI Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'meu-projeto'          # Nome usado no DefectDojo
      stack_type: 'node'                    # Opções: node, dotnet, python, java, cpp, go, rust
      run_dast: true                        # true = executa OWASP ZAP
      dast_target_url: 'https://staging.meu-projeto.com'  # Obrigatório se run_dast: true
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

### Parâmetros

| Parâmetro | Tipo | Obrigatório | Descrição |
|-----------|------|-------------|-----------|
| `project_name` | `string` | ✅ Sim | Nome do projeto, usado para identificar o produto/engagement no DefectDojo |
| `stack_type` | `string` | ✅ Sim | Stack tecnológica: `node`, `dotnet`, `python`, `java`, `cpp`, `go`, `rust` |
| `run_dast` | `boolean` | Não (default: `false`) | Habilita o scan dinâmico com OWASP ZAP |
| `dast_target_url` | `string` | Só se `run_dast: true` | URL da aplicação web/API alvo para o ZAP |

### Secrets

| Secret | Descrição |
|--------|-----------|
| `DEFECTDOJO_URL` | URL base da instância do DefectDojo (ex: `https://defectdojo.exemplo.com`) |
| `DEFECTDOJO_API_KEY` | Token de API para autenticação na API REST do DefectDojo |

### Exemplo sem DAST (projeto CLI/biblioteca)

```yaml
jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'minha-lib-python'
      stack_type: 'python'
      run_dast: false
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

---

## 6. Geração de Artefatos (Relatórios)

Ao final de cada execução, a pipeline compacta e exporta **todos os relatórios JSON** gerados pelos scanners como um artefato do GitHub Actions chamado `devsecops-reports`.

**Estrutura do artefato:**

```
devsecops-reports/
├── gitleaks/
│   └── gitleaks-results.json       # Resultados do Gitleaks (secrets)
├── semgrep/
│   └── semgrep-results.json        # Resultados do Semgrep (SAST)
├── trivy/
│   └── trivy-fs-results.json       # Resultados do Trivy (SCA)
└── zap/
    └── report_json.json            # Resultados do OWASP ZAP (DAST) — só se run_dast=true
```

**Como acessar:**

1. Vá em **Actions** no seu repositório do GitHub.
2. Clique na execução do workflow desejada.
3. Role até **Artifacts** e faça o download de `devsecops-reports`.

Esses relatórios JSON são compatíveis com a API de importação do **DefectDojo** (`/api/v2/import-scan/`), permitindo a centralização e gestão de vulnerabilidades em um painel unificado. Os scripts de importação estão disponíveis em [`scripts/`](scripts/) nesta central.

---

## Stacks Suportadas

Para informações detalhadas sobre a cobertura de cada scanner por stack, incluindo regras Semgrep customizadas e notas específicas por linguagem, consulte [docs/suporte-linguagens.md](docs/suporte-linguagens.md).

| Stack | `stack_type` | Versão |
|-------|-------------|--------|
| Node.js | `node` | 20 |
| .NET | `dotnet` | 8.0.x |
| Python | `python` | 3.12 |
| Java | `java` | 17 (Temurin) |
| C/C++ | `cpp` | build-essential + cmake |
| Go | `go` | stable |
| Rust | `rust` | stable |

---

## Equipe

- Alexia Josielly Duarte da Silva Alves
- João Henrique Lopes de Araújo Freire
- Pedro Henrique Borges Silva
- Raphaela Samille Ramalho de Oliveira
- Thiago Farias Leal
