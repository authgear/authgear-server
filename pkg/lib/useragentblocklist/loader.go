package useragentblocklist

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func Load(resources *resource.Manager) (*blocklist.Blocklist, error) {
	result, err := resources.Read(context.Background(), UserAgentBlockListTXT, resource.EffectiveResource{})
	if err != nil {
		return nil, err
	}

	list, ok := result.(*blocklist.Blocklist)
	if !ok {
		return nil, fmt.Errorf("unexpected bot user agent blocklist type %T", result)
	}

	return list, nil
}

func MustLoad(resources *resource.Manager) *blocklist.Blocklist {
	list, err := Load(resources)
	if err != nil {
		panic(err)
	}
	return list
}
