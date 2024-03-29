package authenticationflow

import (
	"context"
	"net/http"
)

type CookieGetter interface {
	GetCookies(ctx context.Context, deps *Dependencies, flows Flows) ([]*http.Cookie, error)
}

func CollectCookies(ctx context.Context, deps *Dependencies, flows Flows) (cookies []*http.Cookie, err error) {
	err = TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			if n, ok := nodeSimple.(CookieGetter); ok {
				c, err := n.GetCookies(ctx, deps, flows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if i, ok := intent.(CookieGetter); ok {
				c, err := i.GetCookies(ctx, deps, flows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
	}, flows.Nearest)
	if err != nil {
		return
	}

	return
}
