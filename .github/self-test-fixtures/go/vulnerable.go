// =============================================================================
// vulnerable.go — Fixture de teste para validação do Semgrep (stack: go)
//
// ATENÇÃO: Este código é PROPOSITALMENTE INSEGURO.
// Ele existe apenas para garantir que o scanner Semgrep detecte
// vulnerabilidades em projetos Go durante o auto-teste da esteira.
// NÃO use este código em produção.
// =============================================================================

package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"net/http"
)

// Regra esperada: go-sql-string-concat (SQL injection via concatenação)
func getUserByName(db *sql.DB, name string) {
	query := "SELECT * FROM users WHERE name = '" + name + "'"
	db.Query(query) // Semgrep deve detectar SQL injection aqui
}

// Regra esperada: go-weak-hash (uso de MD5 para hashing)
func hashPassword(password string) string {
	h := md5.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil)) // Semgrep deve detectar weak hash aqui
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	fmt.Fprintf(w, "Hello, %s!", name) // potencial XSS
}

func main() {
	fmt.Println(hashPassword("teste123"))
}
