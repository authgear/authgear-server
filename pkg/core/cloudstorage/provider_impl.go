package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/http/httpsigning"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

type providerImpl struct {
	storage      Storage
	appID        string
	secret       []byte
	timeProvider coreTime.Provider
}

var _ Provider = &providerImpl{}

func NewProvider(appID string, storage Storage, secret string, timeProvider coreTime.Provider) Provider {
	return &providerImpl{
		appID:        appID,
		storage:      storage,
		secret:       []byte(secret),
		timeProvider: timeProvider,
	}
}

func (p *providerImpl) AssetNameToAssetID(assetName string) string {
	// This is the final name in the storage.
	// It must not start with a leading slash because
	// /a/b is treated as <empty> / a / b by Azure Storage.
	return fmt.Sprintf("%s/%s", p.appID, assetName)
}

func (p *providerImpl) AssetIDToAssetName(assetID string) string {
	return strings.TrimPrefix(assetID, fmt.Sprintf("%s/", p.appID))
}

func (p *providerImpl) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	contentLength := r.ContentLength()
	if contentLength <= 0 || contentLength > MaxContentLength {
		return nil, ErrTooLargeAsset
	}

	assetName, err := r.DeriveAssetName()
	if err != nil {
		return nil, err
	}

	r.SetCacheControl()

	r.RemoveEmptyHeaders()

	assetID := p.AssetNameToAssetID(assetName)

	// Check duplicate
	err = p.checkDuplicate(assetID)
	if err != nil {
		return nil, err
	}

	httpHeader := r.HTTPHeader()
	httpRequest, err := p.storage.PresignPutObject(assetID, r.Access, httpHeader)
	if err != nil {
		return nil, err
	}

	resp := NewPresignUploadResponse(httpRequest, assetName)
	return &resp, nil
}

func (p *providerImpl) checkDuplicate(assetID string) error {
	u, err := p.storage.PresignHeadObject(assetID)
	if err != nil {
		return err
	}
	resp, err := http.Head(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil
	}
	return ErrDuplicateAsset
}

func (p *providerImpl) Sign(scheme string, host string, r *SignRequest) error {
	now := p.timeProvider.NowUTC()
	for i, assetItem := range r.Assets {
		u := &url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   fmt.Sprintf("/_asset/get/%s", assetItem.AssetName),
		}
		httpRequest, _ := http.NewRequest("GET", u.String(), nil)
		httpsigning.Sign(p.secret, httpRequest, now, int(PresignGetExpires.Seconds()))
		r.Assets[i].URL = httpRequest.URL.String()
	}
	return nil
}

func (p *providerImpl) Verify(r *http.Request) error {
	now := p.timeProvider.NowUTC()
	copiedReq := &http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Proto:         r.Proto,
		ProtoMajor:    r.ProtoMajor,
		ProtoMinor:    r.ProtoMinor,
		Header:        r.Header,
		ContentLength: r.ContentLength,
		Host:          r.Host,
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    r.RequestURI,
	}
	if copiedReq.Method == "HEAD" {
		copiedReq.Method = "GET"
	}
	return httpsigning.Verify(p.secret, copiedReq, now)
}

func (p *providerImpl) PresignGetRequest(assetName string) (*url.URL, error) {
	assetID := p.AssetNameToAssetID(assetName)
	return p.storage.PresignGetObject(assetID)
}

func (p *providerImpl) List(r *ListObjectsRequest) (*ListObjectsResponse, error) {
	if r.Prefix != "" {
		r.Prefix = p.AssetNameToAssetID(r.Prefix)
	}
	// 1000 is the greatest common page size.
	r.PageSize = 1000

	resp, err := p.storage.ListObjects(r)
	if err != nil {
		return nil, err
	}

	for i, assetItem := range resp.Assets {
		resp.Assets[i].AssetName = p.AssetIDToAssetName(assetItem.AssetName)
	}

	return resp, nil
}

func (p *providerImpl) Delete(name string) error {
	assetID := p.AssetNameToAssetID(name)
	err := p.storage.DeleteObject(assetID)
	if err != nil {
		return err
	}
	return nil
}

func (p *providerImpl) ProprietaryToStandard(header http.Header) http.Header {
	return p.storage.ProprietaryToStandard(header)
}

func (p *providerImpl) AccessType(header http.Header) AccessType {
	return p.storage.AccessType(header)
}
