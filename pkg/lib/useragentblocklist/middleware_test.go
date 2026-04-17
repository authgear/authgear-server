package useragentblocklist

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/blocklist"
)

func TestMiddleware(t *testing.T) {
	Convey("Middleware", t, func() {
		list, err := blocklist.New(`
			/(?i)\bGooglebot\b/
			/(?i)\bBytespider\b/
			/(?i)\bCCBot\b/
			/(?i)\bChatGPT-User\b/
			/(?i)\bClaude-User\b/
			/(?i)\bPerplexity-User\b/
			/(?i)\banthropic-ai\b/
			/(?i)\bcohere-ai\b/
			/(?i)\bomgili\b/
			/(?i)\bomgilibot\b/
			/(?i)\bAhrefsBot\b/
			/(?i)\bSemrushBot\b/
			/(?i)\bMJ12bot\b/
			/(?i)\bDotBot\b/
			/(?i)\brogerbot\b/
			/(?i)\bBLEXBot\b/
			/(?i)\bBarkrowler\b/
			/(?i)\bPetalBot\b/
		`)
		So(err, ShouldBeNil)

		middleware := NewMiddleware(list)

		Convey("blocks matching user agents", func() {
			cases := []struct {
				name string
				ua   string
			}{
				{name: "googlebot", ua: "Googlebot/2.1"},
				{name: "bytespider", ua: "Mozilla/5.0 (compatible; Bytespider; +https://zhanzhang.toutiao.com/)"},
				{name: "ccbot", ua: "CCBot/3.0 (http://commoncrawl.org/faq/)"},
				{name: "chatgpt-user", ua: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot"},
				{name: "claude-user", ua: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; Claude-User/1.0; +https://anthropic.com/"},
				{name: "perplexity-user", ua: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexity-user"},
				{name: "anthropic-ai", ua: "anthropic-ai"},
				{name: "cohere-ai", ua: "cohere-ai"},
				{name: "omgili", ua: "omgili"},
				{name: "omgilibot", ua: "omgilibot"},
				{name: "ahrefsbot", ua: "Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)"},
				{name: "semrushbot", ua: "Mozilla/5.0 (compatible; SemrushBot-SI/0.97; +http://www.semrush.com/bot.html)"},
				{name: "mj12bot", ua: "Mozilla/5.0 (compatible; MJ12bot/v1.4.8; http://mj12bot.com/)"},
				{name: "dotbot", ua: "Mozilla/5.0 (compatible; DotBot/1.1; http://www.opensiteexplorer.org/dotbot, help@moz.com)"},
				{name: "rogerbot", ua: "rogerbot/1.2 (https://moz.com/help/guides/moz-procedures/what-is-rogerbot, rogerbot-crawler+aardwolf-production-crawler-01@moz.com)"},
				{name: "blexbot", ua: "Mozilla/5.0 (compatible; BLEXBot/1.0; +http://webmeup-crawler.com/)"},
				{name: "barkrowler", ua: "Mozilla/5.0 (compatible; Barkrowler/0.9; +https://babbar.tech/crawler)"},
				{name: "petalbot", ua: "Mozilla/5.0 (compatible;PetalBot;+https://webmaster.petalsearch.com/site/petalbot)"},
			}

			for _, c := range cases {
				Convey(c.name, func() {
					rr := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.Header.Set("User-Agent", c.ua)

					nextCalled := false
					handler := middleware.Handle(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
						nextCalled = true
					}))
					handler.ServeHTTP(rr, req)

					So(nextCalled, ShouldBeFalse)
					So(rr.Code, ShouldEqual, http.StatusForbidden)
					So(strings.Join(rr.Header().Values("Vary"), ","), ShouldContainSubstring, "User-Agent")
					So(rr.Header().Get("Cache-Control"), ShouldEqual, "no-store")
					So(rr.Header().Get("Pragma"), ShouldEqual, "no-cache")
					So(rr.Header().Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
					So(rr.Body.String(), ShouldEqual, "Your User-Agent is not allowed to access this resource")
				})
			}
		})

		Convey("allows normal browser user agents", func() {
			cases := []struct {
				name string
				ua   string
			}{
				{
					name: "chrome desktop",
					ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
				},
				{
					name: "firefox desktop",
					ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
				},
				{
					name: "safari desktop",
					ua:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
				},
				{
					name: "edge desktop",
					ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.2478.51",
				},
				{
					name: "chrome mobile",
					ua:   "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36",
				},
				{
					name: "safari mobile",
					ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
				},
				{
					name: "edge mobile",
					ua:   "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36 EdgA/124.0.2478.51",
				},
				{
					name: "curl",
					ua:   "curl/8.7.1",
				},
				{
					name: "backend service",
					ua:   "authgear-backend/1.0",
				},
			}

			for _, c := range cases {
				Convey(c.name, func() {
					rr := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					req.Header.Set("User-Agent", c.ua)

					nextCalled := false
					handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						nextCalled = true
						w.Header().Add("Vary", "Origin")
						w.WriteHeader(http.StatusOK)
					}))
					handler.ServeHTTP(rr, req)

					So(nextCalled, ShouldBeTrue)
					So(rr.Code, ShouldEqual, http.StatusOK)
					So(strings.Join(rr.Header().Values("Vary"), ","), ShouldContainSubstring, "User-Agent")
					So(strings.Join(rr.Header().Values("Vary"), ","), ShouldContainSubstring, "Origin")
					So(rr.Header().Get("Cache-Control"), ShouldEqual, "")
					So(rr.Header().Get("Pragma"), ShouldEqual, "")
				})
			}
		})
	})
}
