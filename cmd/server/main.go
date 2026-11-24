package main

import (
	"app/internal/config"
	"app/internal/database"
	"app/internal/features/ai"
	"app/internal/server"
)

var runServer = func(addr string, run func(...string) error) error {
	return run(addr)
}

func main() {
	cfg := config.LoadConfig()
	db := database.ConnectOrFallback(cfg)
	_ = database.Migrate(db)
	_ = database.Seed(db, cfg)
	aiSvc := &ai.Service{
		Chat:   ai.NewClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel),
		Search: ai.NewSearchClient(cfg.SearchEndpoint, cfg.OpenAIKey),
	}
	r := server.NewRouter(cfg, db, aiSvc)
	_ = runServer(":"+cfg.AppPort, r.Run)
}
