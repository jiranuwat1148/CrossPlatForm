package main

import (
	"embed"
	"log"

	"Backup_image/repository"
	"Backup_image/service"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	db, dbErr := repository.NewDbConnection()
	if dbErr != nil {
		log.Printf("database unavailable: %v", dbErr)
	}
	backupRepository := repository.NewBackupRepository(db)
	imageRepository := repository.NewImageRepository(db)
	app := NewApp(
		service.NewBackupService(backupRepository),
		service.NewImageService(imageRepository),
		dbErr,
	)

	err := wails.Run(&options.App{
		Title:  "Backup Image",
		Width:  1100,
		Height: 750,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
