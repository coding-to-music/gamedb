package helpers

import (
	"golang.org/x/net/html"
)

func removeAttribute(n *html.Node, attribute string) {

	for k, v := range n.Attr {
		if v.Key == attribute {
			n.Attr = append(n.Attr[:k], n.Attr[k+1:]...)
			return
		}
	}
}

func getAttribute(n *html.Node, attribute string) string {

	for _, v := range n.Attr {
		if v.Key == attribute {
			return v.Val
		}
	}

	return ""
}

func setAttribute(n *html.Node, attribute string, value string) {

	i := -1
	for index, attr := range n.Attr {
		if attr.Key == attribute {
			i = index
			break
		}
	}
	if i == -1 {
		n.Attr = append(n.Attr, html.Attribute{Key: attribute, Val: value})
	} else {
		n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
	}
}
