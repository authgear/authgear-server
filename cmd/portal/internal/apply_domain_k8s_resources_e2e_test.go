//go:build e2e

package internal

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/authgear/authgear-server/pkg/portal/service"
)

// TestE2EApplyDomainK8SResources runs the SQL of the command against the database of the e2e environment,
// which has the real schema from the portal migrations. It is run by the go_test step in
// e2e/tests/portal/domain_apply_k8s_resources.test.yaml.
func TestE2EApplyDomainK8SResources(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	dbSchema := os.Getenv("DATABASE_SCHEMA")
	if dbURL == "" || dbSchema == "" {
		t.Fatal("DATABASE_URL and DATABASE_SCHEMA must be set, see e2e/.env")
	}
	db := openDB(dbURL, dbSchema)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	Convey("apply-k8s-resources against the e2e database", t, func() {
		ctx := context.Background()
		kube := newTestKubernetes(t)
		var out bytes.Buffer

		// The e2e database is shared with other tests, so every case uses its own apps and domains.
		suffix := randomE2ESuffix(t)
		appA := "apply-k8s-a-" + suffix
		appB := "apply-k8s-b-" + suffix
		Reset(func() {
			for _, table := range []string{"_portal_domain", "_portal_pending_domain"} {
				_, err := db.Exec(`DELETE FROM `+pq.QuoteIdentifier(table)+` WHERE app_id = $1 OR app_id = $2`, appA, appB)
				if err != nil {
					t.Error(err)
				}
			}
		})

		Convey("creates a verified custom domain, removes the pending one and applies its resources", func() {
			domain := "auth." + suffix + ".com"
			insertE2EDomain(t, db, "_portal_pending_domain", appA, domain, suffix+".com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: appA, Domain: domain}, &out)
			So(err, ShouldBeNil)

			var id, apexDomain, nonce string
			var isCustom bool
			var createdAt time.Time
			err = db.QueryRow(
				`SELECT id, apex_domain, verification_nonce, is_custom, created_at FROM _portal_domain WHERE app_id = $1 AND domain = $2`,
				appA, domain,
			).Scan(&id, &apexDomain, &nonce, &isCustom, &createdAt)
			So(err, ShouldBeNil)
			So(apexDomain, ShouldEqual, suffix+".com")
			So(nonce, ShouldNotBeEmpty)
			So(isCustom, ShouldBeTrue)
			So(createdAt.IsZero(), ShouldBeFalse)

			So(countE2EDomains(t, db, "_portal_pending_domain", appA), ShouldEqual, 0)

			ingress, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-"+domain, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(ingress.GetLabels()[service.LabelDomainID], ShouldEqual, service.DomainModelID(true, id))
		})

		Convey("uses the existing domain of the app", func() {
			domain := "auth." + suffix + ".com"
			id := insertE2EDomain(t, db, "_portal_domain", appA, domain, suffix+".com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: appA, Domain: domain}, &out)
			So(err, ShouldBeNil)
			So(out.String(), ShouldStartWith, "domain exists: ")
			So(countE2EDomains(t, db, "_portal_domain", appA), ShouldEqual, 1)

			ingress, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-"+domain, metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(ingress.GetLabels()[service.LabelDomainID], ShouldEqual, service.DomainModelID(true, id))
		})

		Convey("refuses a domain whose apex domain is used, which the table does not allow either", func() {
			insertE2EDomain(t, db, "_portal_domain", appA, "auth."+suffix+".com", suffix+".com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: appB, Domain: "login." + suffix + ".com"}, &out)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "conflicts with existing domains")
			So(countE2EDomains(t, db, "_portal_domain", appB), ShouldEqual, 0)

			// The check mirrors UNIQUE (apex_domain) of _portal_domain. If the constraint changes, so must the check.
			_, err = db.Exec(
				`INSERT INTO _portal_domain (id, app_id, created_at, domain, apex_domain, verification_nonce, is_custom) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				"apply-k8s-dup-"+suffix, appB, time.Now().UTC(), "login."+suffix+".com", suffix+".com", "nonce", true,
			)
			So(err, ShouldNotBeNil)
		})

		Convey("in a dry run, does not keep the domain", func() {
			domain := "auth." + suffix + ".com"
			insertE2EDomain(t, db, "_portal_pending_domain", appA, domain, suffix+".com", true)

			domains, err := ensureCustomDomain(ctx, db, ApplyDomainK8SResourcesOptions{AppID: appA, Domain: domain, DryRun: true}, &out)
			So(err, ShouldBeNil)
			So(domains, ShouldHaveLength, 1)
			So(countE2EDomains(t, db, "_portal_domain", appA), ShouldEqual, 0)
			So(countE2EDomains(t, db, "_portal_pending_domain", appA), ShouldEqual, 1)
		})

		Convey("with an app ID, applies the resources of the verified domains of that app", func() {
			insertE2EDomain(t, db, "_portal_domain", appA, appA+".authgear.cloud", appA+".authgear.cloud", false)
			insertE2EDomain(t, db, "_portal_domain", appA, "auth."+suffix+".com", suffix+".com", true)
			insertE2EDomain(t, db, "_portal_domain", appB, "auth."+suffix+"-b.com", suffix+"-b.com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: appA}, &out)
			So(err, ShouldBeNil)
			So(out.String(), ShouldContainSubstring, "processed 2 domains\n")
			So(listTestResourceNames(t, kube, ingressGVR), ShouldResemble, []string{"traefik-auth." + suffix + ".com"})
		})
	})
}

func randomE2ESuffix(t *testing.T) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func insertE2EDomain(t *testing.T, db *sql.DB, table string, appID string, domain string, apexDomain string, isCustom bool) (id string) {
	id = "apply-k8s-" + randomE2ESuffix(t)
	_, err := db.Exec(
		`INSERT INTO `+pq.QuoteIdentifier(table)+` (id, app_id, created_at, domain, apex_domain, verification_nonce, is_custom) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, appID, time.Now().UTC(), domain, apexDomain, "nonce", isCustom,
	)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func countE2EDomains(t *testing.T, db *sql.DB, table string, appID string) int {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM `+pq.QuoteIdentifier(table)+` WHERE app_id = $1`, appID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}
