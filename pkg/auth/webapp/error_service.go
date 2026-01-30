package webapp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

// queryKeyError is "q_error" so that it is not persisent across pages.
const queryKeyError = "q_error"

type ErrorState struct {
	Form       url.Values          `json:"form,omitempty"`
	Error      *apierrors.APIError `json:"error,omitempty"`
	TrackingID string              `json:"tracking_id,omitempty"`
}

type ErrorService struct {
	AppID       config.AppID
	Cookie      ErrorTokenCookieDef
	RedisHandle *appredis.Handle
	Cookies     CookieManager
}

func (c *ErrorService) HasError(ctx context.Context, r *http.Request) bool {
	_, ok := c.GetRecoverableError(ctx, r)
	if ok {
		return true
	}

	_, ok = c.GetNonRecoverableError(r)
	if ok {
		return true
	}

	return false
}

func (c *ErrorService) PopError(ctx context.Context, w http.ResponseWriter, r *http.Request) (*ErrorState, bool) {
	errorState, ok := c.GetDelRecoverableError(ctx, w, r)
	if ok {
		return errorState, true
	}

	errorState, ok = c.GetNonRecoverableError(r)
	if ok {
		return errorState, true
	}

	return nil, false
}

func (c *ErrorService) GetRecoverableError(ctx context.Context, r *http.Request) (*ErrorState, bool) {
	cookie, cookieErr := c.Cookies.GetCookie(r, c.Cookie.Def)
	if cookieErr != nil {
		return nil, false
	}
	if cookie.Value == "" {
		return nil, false
	}

	token := cookie.Value
	tokenHash := crypto.SHA256String(token)
	redisKey := redisKeyWebError(c.AppID, tokenHash)

	var redisValue string
	var err error
	err = c.RedisHandle.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		redisValue, err = conn.Get(ctx, redisKey).Result()
		return err
	})
	if err != nil {
		return nil, false
	}

	dataBytes := []byte(redisValue)
	var errorState ErrorState
	err = json.Unmarshal(dataBytes, &errorState)
	if err != nil {
		return nil, false
	}

	return &errorState, true
}

func (c *ErrorService) GetDelRecoverableError(ctx context.Context, w http.ResponseWriter, r *http.Request) (*ErrorState, bool) {
	cookie, cookieErr := c.Cookies.GetCookie(r, c.Cookie.Def)
	if cookieErr != nil {
		return nil, false
	}
	if cookie.Value == "" {
		return nil, false
	}

	token := cookie.Value
	tokenHash := crypto.SHA256String(token)
	redisKey := redisKeyWebError(c.AppID, tokenHash)

	var redisValue string
	var err error
	err = c.RedisHandle.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		redisValue, err = conn.Get(ctx, redisKey).Result()
		if err != nil {
			return err
		}

		_, err = conn.Del(ctx, redisKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, false
	}

	dataBytes := []byte(redisValue)
	var errorState ErrorState
	err = json.Unmarshal(dataBytes, &errorState)
	if err != nil {
		return nil, false
	}

	cookie = c.Cookies.ClearCookie(c.Cookie.Def)
	httputil.UpdateCookie(w, cookie)
	return &errorState, true
}

func (c *ErrorService) GetNonRecoverableError(r *http.Request) (*ErrorState, bool) {
	q := r.URL.Query()
	value := q.Get(queryKeyError)
	if value == "" {
		return nil, false
	}

	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, false
	}

	var errorState ErrorState
	err = json.Unmarshal(data, &errorState)
	if err != nil {
		return nil, false
	}

	return &errorState, true
}

// SetRecoverableError stores the error in cookie and retains the form.
func (c *ErrorService) SetRecoverableError(ctx context.Context, r *http.Request, value *apierrors.APIError) (*http.Cookie, error) {
	token, tokenHash := newErrorToken()

	redisKey := redisKeyWebError(c.AppID, tokenHash)

	apiError := apierrors.AsAPIErrorWithContext(ctx, value)
	dataBytes, err := json.Marshal(&ErrorState{
		Form:       r.Form,
		Error:      apiError,
		TrackingID: apiError.TrackingID,
	})
	if err != nil {
		return nil, err
	}

	redisValue := string(dataBytes)
	err = c.RedisHandle.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, redisKey, redisValue, duration.WebError).Result()
		return err
	})
	if err != nil {
		return nil, err
	}

	cookie := c.Cookies.ValueCookie(c.Cookie.Def, token)
	return cookie, nil
}

// SetNonRecoverableError does NOT retain form.
func (c *ErrorService) SetNonRecoverableError(ctx context.Context, result *Result, value *apierrors.APIError) error {
	apiError := apierrors.AsAPIErrorWithContext(ctx, value)
	dataBytes, err := json.Marshal(&ErrorState{
		Error:      apiError,
		TrackingID: apiError.TrackingID,
	})
	if err != nil {
		return err
	}

	queryValue := base64.RawURLEncoding.EncodeToString(dataBytes)

	redirectURI, err := url.Parse(result.RedirectURI)
	if err != nil {
		return err
	}

	q := redirectURI.Query()
	q.Set(queryKeyError, queryValue)
	redirectURI.RawQuery = q.Encode()

	result.RedirectURI = redirectURI.String()
	return nil
}

const (
	errorTokenAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func newErrorToken() (token string, tokenHash string) {
	token = rand.StringWithAlphabet(32, errorTokenAlphabet, rand.SecureRand)
	tokenHash = crypto.SHA256String(token)
	return
}

func redisKeyWebError(appID config.AppID, errorTokenHash string) string {
	return fmt.Sprintf("app:%s:web-error:%s", appID, errorTokenHash)
}
