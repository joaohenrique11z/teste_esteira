// =============================================================================
// vulnerable.c — Fixture de teste para validação do Semgrep (stack: cpp)
//
// ATENÇÃO: Este código é PROPOSITALMENTE INSEGURO.
// Ele existe apenas para garantir que o scanner Semgrep detecte
// vulnerabilidades em projetos C/C++ durante o auto-teste da esteira.
// NÃO use este código em produção.
// =============================================================================

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Regra esperada: cpp-command-injection (uso de system() com input do usuário)
void execute_command(const char *user_input) {
    char cmd[256];
    sprintf(cmd, "echo %s", user_input);
    system(cmd);  // Semgrep deve detectar command injection aqui
}

// Regra esperada: cpp-unsafe-string-functions (uso de strcpy sem bounds check)
void copy_string(const char *src) {
    char dest[64];
    strcpy(dest, src);  // Semgrep deve detectar buffer overflow aqui
    printf("Copiado: %s\n", dest);
}

int main() {
    execute_command("hello");
    copy_string("teste");
    return 0;
}
