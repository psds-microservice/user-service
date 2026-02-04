// gen_http_paths извлекает из .proto пути и методы google.api.http и генерирует Go-константы.
// Запуск: go run scripts/gen_http_paths.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	protoPath  = "pkg/user_service/user_service.proto"
	outPath    = "pkg/gen/user_service/http_paths.go"
	apiPrefix  = "/api/v1"
	basePrefix = "BasePathAPI = \"/api/v1\""
)

func main() {
	data, err := os.ReadFile(protoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read proto: %v\n", err)
		os.Exit(1)
	}
	content := string(data)

	// Найти все rpc Name (...) и следующий за ним option (google.api.http) с method: "path"
	rpcRe := regexp.MustCompile(`rpc\s+(\w+)\s*\([^)]+\)\s*returns\s*\([^)]+\)`)
	httpRe := regexp.MustCompile(`(get|post|put|patch|delete)\s*:\s*"([^"]+)"`)

	lines := strings.Split(content, "\n")
	var rpcName string
	var inOption bool
	var path, method string
	var entries []struct {
		rpcName string
		method  string
		path    string
	}

	for i, line := range lines {
		if m := rpcRe.FindStringSubmatch(line); len(m) > 0 {
			rpcName = m[1]
			inOption = false
			continue
		}
		if strings.Contains(line, "option (google.api.http)") {
			inOption = true
			continue
		}
		if inOption && rpcName != "" {
			if m := httpRe.FindStringSubmatch(line); len(m) > 0 {
				method = strings.ToUpper(m[1])
				path = m[2]
				if strings.HasPrefix(path, apiPrefix) {
					path = path[len(apiPrefix):]
					if path == "" {
						path = "/"
					}
				}
				entries = append(entries, struct {
					rpcName string
					method  string
					path    string
				}{rpcName, method, path})
				rpcName = ""
				inOption = false
			}
			if strings.TrimSpace(line) == "};" {
				inOption = false
			}
		}
		_ = i
	}

	if err := os.MkdirAll("pkg/gen/user_service", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	fmt.Fprintf(w, "// Code generated from %s (google.api.http). DO NOT EDIT.\n", protoPath)
	fmt.Fprintf(w, "// Run: go run scripts/gen_http_paths.go or make proto-http-paths\n\n")
	fmt.Fprintf(w, "package user_service\n\n")
	fmt.Fprintf(w, "const (\n")
	fmt.Fprintf(w, "\t// %s — базовый префикс API (все пути в proto под /api/v1)\n", basePrefix)
	fmt.Fprintf(w, "\tBasePathAPI = %q\n\n", apiPrefix)
	for _, e := range entries {
		name := e.rpcName
		fmt.Fprintf(w, "\t// %s\n", name)
		fmt.Fprintf(w, "\tPath%s   = %q\n", name, e.path)
		fmt.Fprintf(w, "\tMethod%s = %q\n\n", name, e.method)
	}
	fmt.Fprintf(w, ")\n")

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ %s (%d paths)\n", outPath, len(entries))
}
