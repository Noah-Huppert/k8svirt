package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"fmt"
	"encoding/json"
	"strings"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type K8sConfig struct {
	Path string `json:"path"`
	Server string `json:"server"`
}

var logger *log.Logger
const K8SVIRT_CONFIG_FILE_NAME string = "k8svirt-config.json"
var appDef []K8sConfig

func RouteIndexHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("    [%v] %v => Index page", r.Method, r.URL)

	index, _ := json.MarshalIndent(map[string][]K8sConfig{"endpoints": appDef}, "",  "     ")
	fmt.Fprintf(w, "%s", index)
}

func main() {
	logger = log.New(os.Stdout, "[k8svirt] ", 0)
	flag.Parse()

	// Map application
	if len(flag.Args()) < 1 {
		fmt.Println("usage: k8svirt [application dir]")
		os.Exit(1)
	}

	appDir, _ := filepath.Abs(flag.Arg(0))

	logger.Printf("Mapping \"%v\"", appDir)

	filepath.Walk(appDir, walkFunc)

	mux := http.NewServeMux()

	mux.HandleFunc("/", RouteIndexHandler)

	for _, config := range appDef {
		serverUrl, _ := url.Parse(config.Server)
		mux.HandleFunc(config.Path, func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("    [%v] %v => %v", r.Method, r.URL, config.Server)
			httputil.NewSingleHostReverseProxy(serverUrl).ServeHTTP(w, r)
		})
	}

	logger.Printf("Serving on :8080")
	http.ListenAndServe(":8080", mux)
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() && info.Name() == "vendor" {
		return filepath.SkipDir	
	}

	if !info.IsDir() && info.Name() == K8SVIRT_CONFIG_FILE_NAME {
		pathArr := strings.Split(path, "/")
		name := strings.Join(pathArr[len(pathArr) - 2:], "/")

		logger.Printf("    Found \"%v\"", name)

		handle, err := os.Open(path)

		if err == nil {
			jsonParser := json.NewDecoder(handle)

			var config K8sConfig

			if err = jsonParser.Decode(&config); err == nil {
				logger.Printf("        path => %v", config.Path)
				logger.Printf("        server => %v", config.Server)
				appDef = append(appDef, config)
			} else {
				logger.Printf("        Skipping")
				logger.Printf("        Json decoder error: %v", err.Error())
			}
		} else {
			logger.Printf("        Skipping")
			logger.Printf("        File error: %v", err.Error())	
		}
	}

	return nil
}
