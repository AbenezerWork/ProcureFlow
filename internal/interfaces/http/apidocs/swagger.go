package apidocs

import (
	"html/template"
	"net/http"
)

var swaggerTemplate = template.Must(template.New("swagger-ui").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ProcureFlow API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin="anonymous"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: {{ .SpecURL }},
        dom_id: '#swagger-ui',
        deepLinking: true,
        persistAuthorization: true
      });
    };
  </script>
</body>
</html>
`))

type swaggerViewData struct {
	SpecURL string
}

func SwaggerUIHandler(specURL string) http.Handler {
	data := swaggerViewData{
		SpecURL: specURL,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := swaggerTemplate.Execute(w, data); err != nil {
			http.Error(w, "render swagger ui", http.StatusInternalServerError)
		}
	})
}
