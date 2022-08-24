package tasks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/web3"

	web3util "github.com/authgear/authgear-server/pkg/util/web3"
)

const WatchNFTCollections = "WatchNFTCollections"

type WatchNFTCollectionsParam struct {
	AppID string
}

func (p *WatchNFTCollectionsParam) TaskName() string {
	return WatchNFTCollections
}

func ConfigureWatchNFTCollectionsTask(registry task.Registry, t task.Task) {
	registry.Register(WatchNFTCollections, t)
}

type NFTService interface {
	WatchNFTCollection(contractID web3.ContractID) (*model.WatchColletionResponse, error)
}

type WatchNFTCollectionsLogger struct{ *log.Logger }

func NewWatchNFTCollectionsLogger(lf *log.Factory) WatchNFTCollectionsLogger {
	return WatchNFTCollectionsLogger{lf.New("watch-nft-collections")}
}

type WatchNFTCollectionsTask struct {
	NFTService   NFTService
	ConfigSource *configsource.ConfigSource
	Logger       WatchNFTCollectionsLogger
}

func (t *WatchNFTCollectionsTask) Run(ctx context.Context, param task.Param) (err error) {
	taskParam := param.(*WatchNFTCollectionsParam)

	appCtx, err := t.ConfigSource.ContextResolver.ResolveContext(taskParam.AppID)
	if err != nil {
		t.Logger.WithError(err).Error("failed to load app context")
		return err
	}

	web3Config := appCtx.Config.AppConfig.Web3
	if web3Config == nil || web3Config.NFT == nil {
		t.Logger.Info("no NFTConfig found, skipping task")
		return
	}

	if len(web3Config.NFT.Collections) == 0 {
		t.Logger.Info("no collections found in config, skipping task")
		return
	}
	for _, collection := range web3Config.NFT.Collections {
		contractID, err := web3util.ParseContractID(collection)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"collection": collection,
			}).Error("failed to parse collection to contractID")
			return err
		}

		_, err = t.NFTService.WatchNFTCollection(*contractID)
		if err != nil {
			t.Logger.WithError(err).WithFields(logrus.Fields{
				"collection": collection,
			}).Error("failed to watch collection")
			return err
		}
	}

	return
}
