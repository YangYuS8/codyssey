package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/YangYuS8/codyssey/backend/internal/http/router"
)

// openAPI minimal structure for paths
type openAPIDoc struct {
    Paths map[string]map[string]any `yaml:"paths"`
}

// Replace ":param" with "{param}" for comparison with OpenAPI style
var ginParamRe = regexp.MustCompile(`:([a-zA-Z0-9_]+)`) // :id -> {id}

func normalizeGinPath(p string) string {
    return ginParamRe.ReplaceAllString(p, `{$1}`)
}

func main() {
    var openapiPath string
    var includeMetrics bool
    // 默认假设在 backend 目录下执行:  go run ./cmd/routecheck
    // 因此默认 OpenAPI 路径为 ../docs/openapi.yaml （repo 根的 docs）
    flag.StringVar(&openapiPath, "openapi", "../docs/openapi.yaml", "path to openapi.yaml (default assumes running from backend dir)")
    flag.BoolVar(&includeMetrics, "include-metrics", false, "include /metrics endpoint in diff output")
    flag.Parse()

    // Resolve path relative to executable directory (assumes run via go run from cmd/routecheck)
    if !filepath.IsAbs(openapiPath) {
        openapiPath = filepath.Clean(openapiPath)
    }

    data, err := os.ReadFile(openapiPath)
    if err != nil {
        log.Fatalf("read openapi file: %v", err)
    }
    var doc openAPIDoc
    if err := yaml.Unmarshal(data, &doc); err != nil {
        log.Fatalf("unmarshal openapi yaml: %v", err)
    }

    // Build a router with nil dependencies (only need route registrations that don't depend on repos).
    r := router.Setup(router.Dependencies{Env: "development", Version: "routecheck"})

    type route struct{ Method, Path string }
    routes := r.Routes()
    implemented := make(map[string]map[string]struct{}) // path -> method set (lower)
    for _, rt := range routes {
        p := normalizeGinPath(rt.Path)
        m := strings.ToLower(rt.Method)
        if !includeMetrics && p == "/metrics" { // usually excluded
            continue
        }
        // Ignore internal gin irrelevant handlers (favicon etc) - none currently
        if implemented[p] == nil { implemented[p] = make(map[string]struct{}) }
        implemented[p][m] = struct{}{}
    }

    documented := make(map[string]map[string]struct{})
    for p, methods := range doc.Paths {
        if !includeMetrics && p == "/metrics" { continue }
        for method := range methods { // method keys are lower-case in OpenAPI
            lm := strings.ToLower(method)
            if documented[p] == nil { documented[p] = make(map[string]struct{}) }
            documented[p][lm] = struct{}{}
        }
    }

    type mismatch struct { Path string `json:"path"`; Method string `json:"method"` }
    var implementedOnly []mismatch
    var documentedOnly []mismatch
    var methodMismatches []struct{ Path string `json:"path"`; Implemented []string `json:"implemented"`; Documented []string `json:"documented"` }

    // Compare
    visited := make(map[string]struct{})
    for p, implMethods := range implemented {
        visited[p] = struct{}{}
        docMethods, ok := documented[p]
        if !ok {
            for m := range implMethods { implementedOnly = append(implementedOnly, mismatch{Path: p, Method: strings.ToUpper(m)}) }
            continue
        }
        // Compute method sets
        var implList, docList []string
        for m := range implMethods { implList = append(implList, strings.ToUpper(m)) }
        for m := range docMethods { docList = append(docList, strings.ToUpper(m)) }
        sort.Strings(implList); sort.Strings(docList)
        // If method sets differ, record mismatch details
        if len(implList) != len(docList) || !equalStringSlices(implList, docList) {
            // find implementedOnly & documentedOnly at method level
            implSet := make(map[string]struct{}, len(implList))
            docSet := make(map[string]struct{}, len(docList))
            for _, v := range implList { implSet[v] = struct{}{} }
            for _, v := range docList { docSet[v] = struct{}{} }
            for _, v := range implList { if _, ok := docSet[v]; !ok { implementedOnly = append(implementedOnly, mismatch{Path: p, Method: v}) } }
            for _, v := range docList { if _, ok := implSet[v]; !ok { documentedOnly = append(documentedOnly, mismatch{Path: p, Method: v}) } }
            methodMismatches = append(methodMismatches, struct{ Path string `json:"path"`; Implemented []string `json:"implemented"`; Documented []string `json:"documented"` }{Path: p, Implemented: implList, Documented: docList})
        }
    }
    for p, docMethods := range documented {
        if _, ok := visited[p]; ok { continue }
        for m := range docMethods { documentedOnly = append(documentedOnly, mismatch{Path: p, Method: strings.ToUpper(m)}) }
    }

    sort.Slice(implementedOnly, func(i, j int) bool { if implementedOnly[i].Path == implementedOnly[j].Path { return implementedOnly[i].Method < implementedOnly[j].Method }; return implementedOnly[i].Path < implementedOnly[j].Path })
    sort.Slice(documentedOnly, func(i, j int) bool { if documentedOnly[i].Path == documentedOnly[j].Path { return documentedOnly[i].Method < documentedOnly[j].Method }; return documentedOnly[i].Path < documentedOnly[j].Path })
    sort.Slice(methodMismatches, func(i, j int) bool { return methodMismatches[i].Path < methodMismatches[j].Path })

    output := map[string]any{
        "implemented_only": implementedOnly,
        "documented_only": documentedOnly,
        "method_mismatches": methodMismatches,
    }
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    if err := enc.Encode(output); err != nil { log.Fatalf("encode output: %v", err) }
}

func equalStringSlices(a, b []string) bool {
    if len(a) != len(b) { return false }
    for i := range a { if a[i] != b[i] { return false } }
    return true
}
// end
