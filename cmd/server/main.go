package main

import (
	"log"

	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("koneksi database gagal: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("seeder gagal: %v", err)
	}
	aiSvc := &ai.Service{
		Chat:   ai.NewClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel),
		Search: ai.NewSearchClient(cfg.SearchEndpoint, cfg.OpenAIKey),
	}
	r := server.NewRouter(cfg, db, aiSvc)
	log.Fatal(r.Run(":" + cfg.AppPort))
}
