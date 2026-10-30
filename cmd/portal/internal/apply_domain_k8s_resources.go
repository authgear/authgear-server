package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/lib/pq"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
)

type ApplyDomainK8SResourcesOptions struct {
	DatabaseURL         string
	DatabaseSchema      string
	AppID               string
	KubeConfigPath      string
	Namespace           string
	IngressTemplateFile string
	DryRun              bool
	Prune               bool
}

func ApplyDomainK8SResources(ctx context.Context, opts ApplyDomainK8SResourcesOptions) error {
	db := openDB(opts.DatabaseURL, opts.DatabaseSchema)

	domains, err := selectVerifiedDomains(ctx, db, opts.AppID)
	if err != nil {
		return err
	}

	kube := &service.Kubernetes{
		KubernetesConfig: &portalconfig.KubernetesConfig{
			KubeConfig:   opts.KubeConfigPath,
			AppNamespace: opts.Namespace,
		},
		AppConfig: &portalconfig.AppConfig{
			Kubernetes: portalconfig.AppKubernetesConfig{
				IngressTemplateFile: opts.IngressTemplateFile,
			},
		},
	}

	err = kube.ApplyResourcesForDomains(ctx, domains, service.ApplyResourcesForDomainsOptions{
		DryRun: opts.DryRun,
		Prune:  opts.Prune,
		Output: os.Stdout,
	})
	fmt.Printf("processed %d domains\n", len(domains))
	return err
}

// selectVerifiedDomains returns the verified domains, which are the ones with k8s resources.
func selectVerifiedDomains(ctx context.Context, queryer Queryer, appID string) ([]service.DomainResourcesTarget, error) {
	builder := newSQLBuilder().
		Select("id", "app_id", "domain", "is_custom").
		From(pq.QuoteIdentifier("_portal_domain")).
		OrderBy("app_id", "domain")
	if appID != "" {
		builder = builder.Where("app_id = ?", appID)
	}

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := queryer.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []service.DomainResourcesTarget
	for rows.Next() {
		var id string
		var d service.DomainResourcesTarget
		if err := rows.Scan(&id, &d.AppID, &d.Domain, &d.IsCustom); err != nil {
			return nil, err
		}
		d.DomainID = service.DomainModelID(true, id)
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
