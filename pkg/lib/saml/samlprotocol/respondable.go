package samlprotocol

import "github.com/beevik/etree"

type Respondable interface {
	Element() *etree.Element
}
