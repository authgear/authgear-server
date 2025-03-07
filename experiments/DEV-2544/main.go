package main

import (
	"net/http"
	"time"
)

var indexHTML = `
<!DOCTYPE html>
<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>index</title>
	</head>
	<body style="background-color: green;">
		<p>
			This page has a green background.
		</p>

		<button id="use-iframe" type="button">Use iframe to redirect</button>
		<p>
			It is expected that when you use iframe to redirect,
			this page stays unchanged until the iframe has fully loaded.
			By that moment, the iframe will trigger top navigation and redirect.
		</p>

		<form action="/redirect" method="POST">
			<button id="use-form-post" type="submit">POST URL and then be redirected</button>
		</form>
		<p>
			However, as of 2025-03-06 with Chrome Desktop 133, using form post
			will have a similar effect. The browser will still display the current page while
			the form post request is loading.
		</p>
		<script>
			document.getElementById("use-iframe").addEventListener("click", (e) => {
				const newIframe = document.createElement("iframe");
				newIframe.setAttribute("src", "/redirect");
				newIframe.setAttribute("sandbox", "allow-scripts allow-top-navigation");
				newIframe.setAttribute("width", "0");
				newIframe.setAttribute("height", "0");
				newIframe.setAttribute("style", "visibility: hidden;");
				document.body.appendChild(newIframe);
			});
		</script>
	</body>
</html>
`

var redirectHTML = `
<!DOCTYPE html>
<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>redirect</title>
	</head>
	<body style="background-color: red;">
		<script>
			window.location.href = "customuri://host/path";
		</script>
	</body>
</html>
`

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(indexHTML))
	})
	http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Location", "customuri://host/path")
		w.WriteHeader(http.StatusSeeOther)
		w.Write([]byte(redirectHTML))
	})
	http.ListenAndServe(":3000", nil)
}
