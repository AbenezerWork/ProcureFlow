package apidocs

import (
	"embed"
	"fmt"
	"net/http"
)

var (
	//go:embed openapi.yaml
	assets embed.FS
	spec   = mustReadSpec()
)

func mustReadSpec() []byte {
	data, err := assets.ReadFile("openapi.yaml")
	if err != nil {
		panic(fmt.Sprintf("read embedded openapi spec: %v", err))
	}

	return data
}

func SpecHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	})
}
