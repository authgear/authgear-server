package testrunner

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/beevik/etree"
	"golang.org/x/net/html"
)

func readResponseBodyPreserve(response *http.Response) ([]byte, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

func validateHTTPHTML(t *testing.T, step Step, httpOutput *HTTPOutput, response *http.Response) bool {
	body, err := readResponseBodyPreserve(response)
	if err != nil {
		t.Errorf("failed to read response body: %v", err)
		return false
	}

	ok := true
	bodyString := string(body)

	for _, expectedText := range httpOutput.HTMLTextContains {
		if !bytes.Contains(body, []byte(expectedText)) {
			t.Errorf("html text not found in '%s': %q", step.Name, expectedText)
			ok = false
		}
	}

	if len(httpOutput.HTMLXPathExists) == 0 {
		return ok
	}

	doc, err := parseHTMLDocument(bodyString)
	if err != nil {
		t.Errorf("failed to parse response body as html in '%s': %v", step.Name, err)
		return false
	}

	for _, xpath := range httpOutput.HTMLXPathExists {
		if doc.FindElement(xpath) == nil {
			t.Errorf("html xpath not found in '%s': %s", step.Name, xpath)
			ok = false
		}
	}

	return ok
}

func parseHTMLDocument(body string) (*etree.Document, error) {
	root, err := html.Parse(bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	doc := etree.NewDocument()
	appendHTMLNodes(doc, root)
	return doc, nil
}

type etreeChildWriter interface {
	AddChild(token etree.Token)
}

func appendHTMLNodes(parent etreeChildWriter, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			el := etree.NewElement(child.Data)
			for _, attr := range child.Attr {
				el.CreateAttr(attr.Key, attr.Val)
			}
			parent.AddChild(el)
			appendHTMLNodes(el, child)
		case html.TextNode:
			parent.AddChild(etree.NewText(child.Data))
		}
	}
}
