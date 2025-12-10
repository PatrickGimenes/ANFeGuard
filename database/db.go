package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Conectar() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, pass, name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[ERRO] Erro ao conectar ao banco: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("[ERRO] Erro ao testar conexão com banco: %v", err)
	}

	DB = db
	log.Println("[INFO] Banco conectado com sucesso.")
}
