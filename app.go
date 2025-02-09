package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v2"
)

// App struct
type App struct {
	ctx        context.Context
	ImgDir     string
	Categories Data
	OutputDir  string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

type Data struct {
	Categories []string `yaml:"categories"`
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SetImgDir() string {
	dialogOps := runtime.OpenDialogOptions{
		Title:                "Set directory for images to classify",
		ShowHiddenFiles:      false,
		CanCreateDirectories: false,
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, dialogOps)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error setting dir %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
	}
	a.ImgDir = dir
	return dir
}

func startHTTPServer(dir string, port int) {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        fs,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		fmt.Printf("Serving %s on http://localhost:%d\n", dir, port)
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error starting HTTP server:", err)
		}
	}()
}

func (a *App) GetCategories() []string {
	yamlFile, err := os.ReadFile("data.yaml")
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error reading config for categories %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
		return []string{}
	}

	// Unmarshal YAML into struct
	var data Data
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error reading yaml %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
		return []string{}
	}
	return data.Categories
}

func (a *App) SetCategory(category string) {
	var data Data

	yamlFile, err := os.ReadFile("data.yaml")
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
		return
	}

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
		return
	}

	data.Categories = append(data.Categories, category)

	// Marshal struct back to YAML
	updatedYAML, err := yaml.Marshal(&data)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
		return
	}

	// Write back to file
	err = os.WriteFile("data.yaml", updatedYAML, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (a *App) GetImageList() []string {
	if len(a.ImgDir) == 0 {
		return []string{}
	}
	var images []string

	files, err := os.ReadDir(a.ImgDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return images
	}

	for _, file := range files {
		// Check if the file is an image (basic filter)
		if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".png" || filepath.Ext(file.Name()) == ".jpeg" {
			images = append(images, filepath.Join(file.Name()))
		}
	}
	startHTTPServer(a.ImgDir, 5618)

	return images
}

// Copy function
func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func (a *App) ClassifyImage(catgory, filename string) error {
	if a.OutputDir == "" {
		wdir, err := os.Getwd()
		if err != nil {

			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Title:   "Error",
				Message: fmt.Sprintf("Error setting workdir %s", err.Error()),
				Type:    runtime.ErrorDialog,
			})
			return err
		}
		a.OutputDir = filepath.Join(wdir, "output")
		log.Println("outout dir", a.OutputDir)
	}

	src := filepath.Join(a.ImgDir, filename)
	dest := filepath.Join(a.OutputDir, catgory, filename)
	ensureDirExists(filepath.Join(a.OutputDir, catgory))
	if os.Getenv("IMG_CLASSIFY_MOVE_FILE") == "true" {
		err := os.Rename(src, dest)
		if err == nil {
			log.Printf("from %s moved successfully to %s", src, dest)
		}
	}

	err := copyFile(src, dest)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "Error",
			Message: fmt.Sprintf("Error copying file %s", err.Error()),
			Type:    runtime.ErrorDialog,
		})
	} else {
		log.Printf("from %s copied successfully to %s", src, dest)
	}
	return nil
}

func ensureDirExists(dirPath string) error {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}
