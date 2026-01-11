package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"

	aiData "github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	articlesData "github.com/Xiaoxinkeji/WX/internal/features/articles/data"
	"github.com/Xiaoxinkeji/WX/internal/ui"
)

func main() {
	dbPath := os.Getenv("WX_DB_PATH")
	if dbPath == "" {
		dbPath = "wx.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	repo, err := articlesData.NewSQLiteRepository(db)
	if err != nil {
		log.Fatalf("articles repo: %v", err)
	}

	prompts := aiData.NewDefaultPromptRepository()

	if err := ui.Run(ui.Config{
		ArticlesRepo: repo,
		Prompts:      prompts,
	}); err != nil {
		log.Fatalf("ui: %v", err)
	}
}
