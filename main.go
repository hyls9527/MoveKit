package main

import (
	"embed"
	"os"

	winsyscall "github.com/movekit/movekit/internal/syscall"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

const currentVersion = "0.1.0"

func main() {
	// Enforce single instance via named mutex.
	handle, err := winsyscall.EnsureSingleInstance("MoveKit_SingleInstance")
	if err != nil {
		println("MoveKit 已在运行，请勿重复打开")
		os.Exit(1)
	}
	defer winsyscall.CloseHandle(handle)

	app := NewApp()
	err = wails.Run(&options.App{
		Title:  "MoveKit - C盘瘦身工具",
		Width:  520,
		Height: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 245, B: 245, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
