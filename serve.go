package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	// "github.com/InkProject/ink.go"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var watcher *fsnotify.Watcher
var conn *websocket.Conn

func buildWatchList() (files []string, dirs []string) {
	dirs = []string{
		filepath.Join(rootPath, "source"),
	}
	files = []string{
		filepath.Join(rootPath, "config.yml"),
		filepath.Join(themePath),
	}

	// Add files and directories defined in theme's config.yml to watcher
	for _, themeCopiedPath := range themeConfig.Copy {
		if themeCopiedPath != "" {
			fullPath := filepath.Join(themePath, themeCopiedPath)
			s, err := os.Stat(fullPath)
			if s == nil || err != nil {
				continue
			}

			if s.IsDir() {
				dirs = append(dirs, fullPath)
			} else {
				files = append(files, fullPath)
			}
		}
	}
	return files, dirs
}

// Add files and dirs to watcher
func configureWatcher(watcher *fsnotify.Watcher, files []string, dirs []string) error {
	for _, source := range dirs {
		filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
			if f != nil && f.IsDir() {
				if err := watcher.Add(path); err != nil {
					slog.Warn(err.Error())
				}
			}
			return nil
		})
	}
	for _, source := range files {
		if err := watcher.Add(source); err != nil {
			slog.Warn(err.Error())
		}
	}
	return nil
}

func Watch() {
	// Listen watched file change event
	if watcher != nil {
		watcher.Close()
	}
	watcher, _ = fsnotify.NewWatcher()
	files, dirs := buildWatchList()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					// Handle when file change
					slog.Info(event.Name)
					ParseGlobalConfigWrap(rootPath, true)

					newFiles, newDirs := buildWatchList()
					// If file list changed, reconfigure watcher
					if !reflect.DeepEqual(files, newFiles) || !reflect.DeepEqual(dirs, newDirs) {
						configureWatcher(watcher, newFiles, newDirs)
						files = newFiles
						dirs = newDirs
					}

					Build()
					if conn != nil {
						if err := conn.WriteMessage(websocket.TextMessage, []byte("change")); err != nil {
							slog.Warn(err.Error())
						}
					}
				}
			case err := <-watcher.Errors:
				slog.Warn(err.Error())
			}
		}
	}()
	configureWatcher(watcher, files, dirs)
}

func Serve() {
	// editorWeb := ink.New()
	//
	// editorWeb.Get("/articles", ApiListArticle)
	// editorWeb.Get("/articles/:id", ApiGetArticle)
	// editorWeb.Post("/articles", ApiCreateArticle)
	// editorWeb.Put("/articles/:id", ApiSaveArticle)
	// editorWeb.Delete("/articles/:id", ApiRemoveArticle)
	// editorWeb.Get("/config", ApiGetConfig)
	// editorWeb.Put("/config", ApiSaveConfig)
	// editorWeb.Post("/upload", ApiUploadFile)
	// editorWeb.Use(ink.Cors)
	// editorWeb.Get("*", ink.Static(filepath.Join("editor/assets")))

	// Log("Access http://localhost:" + globalConfig.Build.Port + "/ to open editor")
	// go editorWeb.Listen(":2333")

	router := http.NewServeMux()

	router.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("create a todo")
	})

	// router.HandleFunc("GET /live", func(w http.ResponseWriter, r *http.Request) {
	// 	var upgrader = websocket.Upgrader{
	// 		ReadBufferSize:  1024,
	// 		WriteBufferSize: 1024,
	// 	}
	// 	if c, err := upgrader.Upgrade(ctx.Res, ctx.Req, nil); err != nil {
	// 		slog.Warn(err.Error())
	// 	} else {
	// 		conn = c
	// 	}
	// 	ctx.Stop()
	// })

	router.HandleFunc("PATCH /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Println("update a todo by id", id)
	})

	router.HandleFunc("DELETE /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Println("delete a todo by id", id)
	})

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":8000", router)

	// previewWeb := ink.New()
	// previewWeb.Get("/live", Websocket)
	// previewWeb.Get("*", ink.Static(filepath.Join(rootPath, globalConfig.Build.Output)))

}
