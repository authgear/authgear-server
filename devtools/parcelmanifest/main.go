package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	texttemplate "text/template"
)

var tpl = `package web_parcel

func init() {
	ParcelAssetMap = map[string]string{"authgear-modern.ts": "authgear.9b2eeef2.js", "authgear.ts": "authgear.b1305cf8.js", "build.html": "build.html", "tailwind.css": "tailwind.9a1182cc.css"}
}
`

func main() {
	pm, err := os.Open("resources/authgear/static/parcel-manifest.json")
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.Create("pkg/lib/web_parcel/parcel_gen.go")
	if err != nil {
		fmt.Println("create file: ", err)
		return
	}

	byteValue, _ := ioutil.ReadAll(pm)

	var m map[string]string
	json.Unmarshal([]byte(byteValue), &m)

	for k, v := range m {
		m[k] = strings.TrimPrefix(v, "/")
	}

	t := texttemplate.New("name")
	t, _ = t.Parse(tpl)
	_ = t.Execute(f, map[string]string{
		"AssetMap": fmt.Sprintf("%#v", m),
	})

	pm.Close()
	f.Close()
}
