# Suporte a Linguagens — Esteira DevSecOps Open Source

Este documento detalha todas as stacks/linguagens suportadas pela esteira DevSecOps central e a cobertura de cada scanner por stack.

## Stacks Suportadas

| Stack | `stack_type` | Toolchain Setup | Versão |
|-------|-------------|-----------------|--------|
| Node.js | `node` | `actions/setup-node@v4` | 20 |
| .NET | `dotnet` | `actions/setup-dotnet@v4` | 8.0.x |
| Python | `python` | `actions/setup-python@v5` | 3.12 |
| Java | `java` | `actions/setup-java@v4` (Temurin) | 17 |
| C/C++ | `cpp` | `apt-get install build-essential cmake` | Sistema |
| Go | `go` | `actions/setup-go@v5` | stable |
| Rust | `rust` | `dtolnay/rust-toolchain@stable` | stable |

## Cobertura de Scanners por Stack

| Scanner | node | dotnet | python | java | cpp | go | rust |
|---------|:----:|:------:|:------:|:----:|:---:|:--:|:----:|
| **Gitleaks** (Secrets) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Semgrep** (SAST) — OWASP Top 10 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Semgrep** (SAST) — Registry específico | `p/nodejs` | `p/csharp` | `p/python` | `p/java` | `p/c` | `p/golang` | `p/rust` |
| **Semgrep** (SAST) — Regras customizadas | ✅ | — | — | — | ✅ | ✅ | — |
| **Trivy** (SCA/Container) | ✅ | ✅ | ✅ | ✅ | ✅¹ | ✅² | ✅³ |
| **OWASP ZAP** (DAST) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

> **Notas:**
> 1. ¹ C/C++: Trivy analisa vulnerabilidades de SO e dependências de sistema. Não possui scanner específico para `CMakeLists.txt`/`Makefile`, mas cobre imagens Docker e filesystem.
> 2. ² Go: Trivy detecta dependências automaticamente via `go.mod`/`go.sum`.
> 3. ³ Rust: Trivy detecta dependências automaticamente via `Cargo.lock`.

## Regras Semgrep Customizadas

### Node.js / JavaScript / TypeScript (existentes)

| ID da Regra | Detecção | CWE | Severidade |
|-------------|----------|-----|------------|
| `nodejs-dangerous-code-execution` | `eval()`, `new Function()` | CWE-95 | ERROR |
| `nodejs-weak-hash-md5-sha1` | `crypto.createHash("md5"/"sha1")` | CWE-327 | ERROR |
| `nodejs-hardcoded-jwt-secret` | JWT secret hardcoded | CWE-798 | ERROR |
| `nodejs-sql-string-concat` | SQL injection por concatenação | CWE-89 | ERROR |
| `nodejs-sensitive-data-in-log` | Dados sensíveis em logs | CWE-532 | ERROR |
| `nodejs-cors-wildcard` | CORS `origin: "*"` | CWE-942 | WARNING |
| `nodejs-debug-enabled` | Debug habilitado | — | WARNING |
| `express-api-route-without-auth` | Rota API sem auth | CWE-285 | WARNING |

### C/C++ (novas)

| ID da Regra | Detecção | CWE | Severidade |
|-------------|----------|-----|------------|
| `cpp-unsafe-string-functions` | `strcpy`, `strcat`, `sprintf`, `gets` — buffer overflow | CWE-120/CWE-121 | ERROR |
| `cpp-malloc-without-null-check` | `malloc` sem verificação NULL | CWE-476 | WARNING |
| `cpp-command-injection` | `system()`, `popen()`, `exec*()` | CWE-78 | ERROR |
| `cpp-integer-overflow-risk` | Cast de inteiro com risco de overflow | CWE-190 | WARNING |

### Go (novas)

| ID da Regra | Detecção | CWE | Severidade |
|-------------|----------|-----|------------|
| `go-command-injection` | `exec.Command` com input externo | CWE-78 | ERROR |
| `go-ignored-error` | Erros ignorados (`_ = err`) | CWE-754 | WARNING |
| `go-weak-hash` | `crypto/md5`, `crypto/sha1` | CWE-327 | ERROR |
| `go-sql-string-concat` | `fmt.Sprintf` em `db.Query` — SQL injection | CWE-89 | ERROR |

## Como Utilizar

### Exemplo de integração para cada stack

```yaml
# Node.js
jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'meu-projeto-node'
      stack_type: 'node'
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

```yaml
# C/C++
jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'meu-projeto-cpp'
      stack_type: 'cpp'
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

```yaml
# Go
jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'meu-projeto-go'
      stack_type: 'go'
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

```yaml
# Rust
jobs:
  security:
    uses: joaohenrique11z/teste_esteira/.github/workflows/pipeline.yml@v1
    with:
      project_name: 'meu-projeto-rust'
      stack_type: 'rust'
    secrets:
      DEFECTDOJO_URL: ${{ secrets.DEFECTDOJO_URL }}
      DEFECTDOJO_API_KEY: ${{ secrets.DEFECTDOJO_API_KEY }}
```

## Shadow Mode

Todas as stacks rodam em **Shadow Mode** por padrão:
- Os scanners **não bloqueiam** o build (usam `|| true` ou `exit-code: '0'`)
- Vulnerabilidades são reportadas como `::warning::` no log do GitHub Actions
- Resultados são enviados automaticamente para o **DefectDojo** para gestão centralizada
