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
	"sort"
)

type K8sConfig struct {
	Path string `json:"path"`
	Server string `json:"server"`
}

type K8sConfigs []K8sConfig

func (confs K8sConfigs) Len() int {
	return len(confs)
}

// Determines if i should be before j
func (confs K8sConfigs) Less(i, j int) bool {
	ic := len(strings.Split(confs[i].Path, "/"))
	jc := len(strings.Split(confs[j].Path, "/"))

	return ic < jc
}

func (confs K8sConfigs) Swap(i, j int) {
	confs[i], confs[j] = confs[j], confs[i]
}

var logger *log.Logger
const LOGGER_PREFIX string = "[k8svirt] "
const K8SVIRT_CONFIG_FILE_NAME string = "k8svirt-config.json"
var appDef K8sConfigs

func RouteIndexHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("    [%v] %v => Index page", r.Method, r.URL)

	index, _ := json.MarshalIndent(map[string][]K8sConfig{"endpoints": appDef}, "",  "     ")
	fmt.Fprintf(w, "%s", index)
}

func main() {
	logger = log.New(os.Stdout, LOGGER_PREFIX, 0)
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

	sort.Sort(appDef)

	logger.Printf("")
	logger.Printf("Configuring routes")
	for _, config := range appDef {
		logger.Printf("    %v => %v", config.Path, config.Server)
		serverUrl, _ := url.Parse(config.Server)

		mux.HandleFunc(config.Path, func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("    [%v] %v => %v", r.Method, r.URL, config.Server)

			proxy := httputil.NewSingleHostReverseProxy(serverUrl)
			proxy.ErrorLog = logger
			proxy.ServeHTTP(w, r)
		})
	}

	logger.Printf("")
	logger.Printf("Serving on :8080")
	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		logger.Fatalf("Error starting server (error: %v)", err)
	}
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
