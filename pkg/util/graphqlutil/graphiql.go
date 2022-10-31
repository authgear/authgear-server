package graphqlutil

import (
	htmltemplate "html/template"
	"net/http"
)

var graphiqlTemplate = htmltemplate.Must(htmltemplate.New("graphiql").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{ .Title }}</title>
	<style>
		body {
			padding: 0;
			margin: 0;
			min-height: 100vh;
		}
		#root {
			height: 100vh;
		}
	</style>
	<script
		crossorigin
		src="https://unpkg.com/react@17.0.2/umd/react.production.min.js"
	></script>
	<script
		crossorigin
		src="https://unpkg.com/react-dom@17.0.2/umd/react-dom.production.min.js"
	></script>
	<link rel="stylesheet" href="https://unpkg.com/graphiql@2.0.13/graphiql.min.css" />
</head>
<body>
	<div id="root">Loading...</div>
	<script
		crossorigin
		src="https://unpkg.com/graphiql@2.0.13/graphiql.min.js"
	></script>
	<script>
		ReactDOM.render(
			React.createElement(GraphiQL, {
				fetcher: GraphiQL.createFetcher({
					url: "",
				}),
				query: (new URLSearchParams(window.location.search)).get("query") || "",
			}),
			document.getElementById("root"),
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
	err := graphiqlTemplate.Execute(w, g)
	if err != nil {
		panic(err)
	}
}
