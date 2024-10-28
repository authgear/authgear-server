package graphqlutil

import (
	htmltemplate "html/template"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var graphiqlTemplate = htmltemplate.Must(htmltemplate.New("graphiql").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{ $.Title }}</title>
	<style nonce="{{ $.CSPNonce }}">
		body {
			height: 100%;
			margin: 0;
			width: 100%;
			overflow: hidden;
		}
		#graphiql {
			height: 100vh;
		}
	</style>
	<script
		nonce="{{ $.CSPNonce }}"
		crossorigin
		src="https://unpkg.com/react@18.2.0/umd/react.production.min.js"
	></script>
	<script
		nonce="{{ $.CSPNonce }}"
		crossorigin
		src="https://unpkg.com/react-dom@18.2.0/umd/react-dom.production.min.js"
	></script>
	<script
		nonce="{{ $.CSPNonce }}"
		crossorigin
		src="https://unpkg.com/graphiql@3.0.9/graphiql.min.js"
	></script>
	<link
		nonce="{{ $.CSPNonce }}"
		rel="stylesheet"
		href="https://unpkg.com/graphiql@3.0.9/graphiql.min.css" />
	<script
		nonce="{{ $.CSPNonce }}"
		crossorigin
		src="https://unpkg.com/@graphiql/plugin-explorer@1.0.2/dist/index.umd.js"
	></script>
	<link
		nonce="{{ $.CSPNonce }}"
		rel="stylesheet"
		href="https://unpkg.com/@graphiql/plugin-explorer@1.0.2/dist/style.css" />
</head>
<body>
	<div id="graphiql">Loading...</div>
	<script nonce="{{ $.CSPNonce }}">
		const root = ReactDOM.createRoot(document.getElementById("graphiql"));
		const fetcher = GraphiQL.createFetcher({
			url: "",
		});
		const explorerPlugin = GraphiQLPluginExplorer.explorerPlugin();
		root.render(
			React.createElement(GraphiQL, {
				fetcher,
				defaultEditorToolsVisibility: true,
				plugins: [explorerPlugin],
				query: (new URLSearchParams(window.location.search)).get("query") || "",
			})
		);
	</script>
</body>
</html>
`))

type GraphiQL struct {
	Title string
}

func (g *GraphiQL) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	nonce := httputil.GetCSPNonce(r.Context())
	data := map[string]interface{}{
		"Title":    g.Title,
		"CSPNonce": nonce,
	}

	err := graphiqlTemplate.Execute(w, data)
	if err != nil {
		panic(err)
	}
}
