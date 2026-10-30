package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/lib/pq"
	"golang.org/x/net/publicsuffix"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
)

type ApplyDomainK8SResourcesOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	AppID          string
	// Domain, if set, is added to the app as a verified custom domain when it does not exist,
	// and only its resources are applied.
	Domain              string
	ApexDomain          string
	KubeConfigPath      string
	Namespace           string
	IngressTemplateFile string
	DryRun              bool
	Prune               bool
}

func ApplyDomainK8SResources(ctx context.Context, opts ApplyDomainK8SResourcesOptions) error {
	db := openDB(opts.DatabaseURL, opts.DatabaseSchema)

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

	return applyDomainK8SResources(ctx, db, kube, opts, os.Stdout)
}

func applyDomainK8SResources(ctx context.Context, db *sql.DB, kube *service.Kubernetes, opts ApplyDomainK8SResourcesOptions, out io.Writer) error {
	var domains []service.DomainResourcesTarget
	var err error
	if opts.Domain != "" {
		domains, err = ensureCustomDomain(ctx, db, opts, out)
	} else {
		domains, err = selectVerifiedDomains(ctx, db, opts.AppID, "")
	}
	if err != nil {
		return err
	}

	err = kube.ApplyResourcesForDomains(ctx, domains, service.ApplyResourcesForDomainsOptions{
		DryRun: opts.DryRun,
		Prune:  opts.Prune,
		Output: out,
	})
	fmt.Fprintf(out, "processed %d domains\n", len(domains))
	return err
}

// ensureCustomDomain adds opts.Domain to opts.AppID as a verified custom domain when the app does not have it yet,
// and returns it. In a dry run, the change is rolled back.
func ensureCustomDomain(ctx context.Context, db *sql.DB, opts ApplyDomainK8SResourcesOptions, out io.Writer) (domains []service.DomainResourcesTarget, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil || opts.DryRun {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
			return
		}
		err = tx.Commit()
	}()

	domains, err = selectVerifiedDomains(ctx, tx, opts.AppID, opts.Domain)
	if err != nil {
		return nil, err
	}
	if len(domains) > 0 {
		fmt.Fprintf(out, "domain exists: %s (app: %s)\n", opts.Domain, opts.AppID)
		return domains, nil
	}

	apexDomain := opts.ApexDomain
	if apexDomain == "" {
		apexDomain, err = publicsuffix.EffectiveTLDPlusOne(opts.Domain)
		if err != nil {
			return nil, fmt.Errorf("cannot derive the apex domain of %s, specify it with --apex-domain: %w", opts.Domain, err)
		}
	}

	// The domain may belong to another app, and _portal_domain allows one domain per apex domain.
	conflicts, err := selectConflictingDomains(ctx, tx, opts.Domain, apexDomain)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, fmt.Errorf("domain %s (apex domain %s) conflicts with existing domains: %v", opts.Domain, apexDomain, conflicts)
	}

	err = createCustomDomain(ctx, tx, opts.AppID, opts.Domain, apexDomain)
	if err != nil {
		return nil, err
	}
	// Verifying in the portal moves a pending domain to the verified ones, do the same.
	err = deletePendingDomain(ctx, tx, opts.AppID, opts.Domain)
	if err != nil {
		return nil, err
	}

	suffix := ""
	if opts.DryRun {
		suffix = " (dry run, rolled back)"
	}
	fmt.Fprintf(out, "domain created: %s (app: %s, apex domain: %s)%s\n", opts.Domain, opts.AppID, apexDomain, suffix)

	return selectVerifiedDomains(ctx, tx, opts.AppID, opts.Domain)
}

// selectConflictingDomains returns the verified domains, as "<domain> (app: <app-id>)",
// that are the domain or have the same apex domain.
func selectConflictingDomains(ctx context.Context, queryer Queryer, domain string, apexDomain string) ([]string, error) {
	builder := newSQLBuilder().
		Select("domain", "app_id").
		From(pq.QuoteIdentifier("_portal_domain")).
		Where("domain = ? OR apex_domain = ?", domain, apexDomain).
		OrderBy("domain")

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := queryer.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var d, appID string
		if err := rows.Scan(&d, &appID); err != nil {
			return nil, err
		}
		result = append(result, fmt.Sprintf("%s (app: %s)", d, appID))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func deletePendingDomain(ctx context.Context, tx *sql.Tx, appID string, domain string) error {
	builder := newSQLBuilder().
		Delete(pq.QuoteIdentifier("_portal_pending_domain")).
		Where("app_id = ? AND domain = ?", appID, domain)

	q, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// selectVerifiedDomains returns the verified domains, which are the ones with k8s resources,
// optionally of one app and one domain.
func selectVerifiedDomains(ctx context.Context, queryer Queryer, appID string, domain string) ([]service.DomainResourcesTarget, error) {
	builder := newSQLBuilder().
		Select("id", "app_id", "domain", "is_custom").
		From(pq.QuoteIdentifier("_portal_domain")).
		OrderBy("app_id", "domain")
	if appID != "" {
		builder = builder.Where("app_id = ?", appID)
	}
	if domain != "" {
		builder = builder.Where("domain = ?", domain)
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
