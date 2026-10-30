package internal

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/restmapper"
	coretesting "k8s.io/client-go/testing"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
)

var ingressGVR = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}
var certificateGVR = schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}

// newTestDomainDB returns an in-memory database with the domain tables of
// cmd/portal/cmd/cmddatabase/migrations/portal (20201008183330-add_domains.sql, 20201009162506-custom_domain_flag.sql).
func newTestDomainDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Every connection to :memory: is a separate database.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Exec(`
CREATE TABLE _portal_pending_domain
(
    id                 text PRIMARY KEY,
    app_id             text      NOT NULL,
    created_at         timestamp NOT NULL,
    domain             text      NOT NULL,
    apex_domain        text      NOT NULL,
    verification_nonce text      NOT NULL,
    is_custom          boolean   NOT NULL,
    UNIQUE (app_id, apex_domain)
);
CREATE TABLE _portal_domain
(
    id                 text PRIMARY KEY,
    app_id             text      NOT NULL,
    created_at         timestamp NOT NULL,
    domain             text      NOT NULL,
    apex_domain        text      NOT NULL,
    verification_nonce text      NOT NULL,
    is_custom          boolean   NOT NULL,
    UNIQUE (apex_domain)
);`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertTestDomain(t *testing.T, db *sql.DB, table string, id string, appID string, domain string, apexDomain string, isCustom bool) {
	_, err := db.Exec(
		`INSERT INTO `+table+` (id, app_id, created_at, domain, apex_domain, verification_nonce, is_custom) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, appID, time.Now().UTC(), domain, apexDomain, "nonce", isCustom,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func countTestRows(t *testing.T, db *sql.DB, table string) int {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func newTestKubernetes(t *testing.T) *service.Kubernetes {
	discovery := &fakediscovery.FakeDiscovery{Fake: &coretesting.Fake{}}
	discovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "networking.k8s.io/v1",
			APIResources: []metav1.APIResource{{Name: "ingresses", Namespaced: true, Kind: "Ingress"}},
		},
		{
			GroupVersion: "cert-manager.io/v1",
			APIResources: []metav1.APIResource{{Name: "certificates", Namespaced: true, Kind: "Certificate"}},
		},
	}
	groupResources, err := restmapper.GetAPIGroupResources(discovery)
	if err != nil {
		t.Fatal(err)
	}

	return &service.Kubernetes{
		AppConfig: &portalconfig.AppConfig{
			Kubernetes: portalconfig.AppKubernetesConfig{IngressTemplateFile: "testdata/ingress.tpl.yaml"},
		},
		DiscoveryRESTMapper: restmapper.NewDiscoveryRESTMapper(groupResources),
		DynamicClient: fakedynamic.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
			ingressGVR:     "IngressList",
			certificateGVR: "CertificateList",
		}),
		Namespace: "test-namespace",
	}
}

func listTestResourceNames(t *testing.T, kube *service.Kubernetes, gvr schema.GroupVersionResource) []string {
	list, err := kube.DynamicClient.Resource(gvr).Namespace(kube.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, item := range list.Items {
		names = append(names, item.GetName())
	}
	return names
}

func TestApplyDomainK8SResources(t *testing.T) {
	Convey("applyDomainK8SResources", t, func() {
		ctx := context.Background()
		db := newTestDomainDB(t)
		kube := newTestKubernetes(t)
		var out bytes.Buffer

		Convey("without a domain, applies the resources of the verified domains of all apps", func() {
			insertTestDomain(t, db, "_portal_domain", "d1", "app1", "app1.authgear.cloud", "app1.authgear.cloud", false)
			insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)
			insertTestDomain(t, db, "_portal_domain", "d3", "app2", "auth.two.com", "two.com", true)
			insertTestDomain(t, db, "_portal_pending_domain", "p1", "app2", "auth.pending.com", "pending.com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{}, &out)
			So(err, ShouldBeNil)
			So(listTestResourceNames(t, kube, ingressGVR), ShouldResemble, []string{"traefik-auth.one.com", "traefik-auth.two.com"})
			So(listTestResourceNames(t, kube, certificateGVR), ShouldResemble, []string{"cert-auth.one.com", "cert-auth.two.com"})
			So(out.String(), ShouldContainSubstring, "processed 3 domains\n")

			ingress, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-auth.one.com", metav1.GetOptions{})
			So(err, ShouldBeNil)
			So(ingress.GetLabels(), ShouldResemble, map[string]string{
				service.LabelAppID:    "app1",
				service.LabelDomainID: service.DomainModelID(true, "d2"),
			})
		})

		Convey("with an app ID, applies the resources of the domains of that app only", func() {
			insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)
			insertTestDomain(t, db, "_portal_domain", "d3", "app2", "auth.two.com", "two.com", true)

			err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app2"}, &out)
			So(err, ShouldBeNil)
			So(listTestResourceNames(t, kube, ingressGVR), ShouldResemble, []string{"traefik-auth.two.com"})
		})

		Convey("with a domain", func() {
			Convey("creates the domain as a verified custom domain, removes it from the pending domains and applies its resources", func() {
				insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)
				insertTestDomain(t, db, "_portal_pending_domain", "p1", "app2", "auth.two.co.uk", "two.co.uk", true)
				insertTestDomain(t, db, "_portal_pending_domain", "p2", "app2", "other.com", "other.com", true)

				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app2", Domain: "auth.two.co.uk"}, &out)
				So(err, ShouldBeNil)
				So(out.String(), ShouldStartWith, "domain created: auth.two.co.uk (app: app2, apex domain: two.co.uk)\n")

				domains, err := selectVerifiedDomains(ctx, db, "app2", "")
				So(err, ShouldBeNil)
				So(domains, ShouldHaveLength, 1)
				So(domains[0].Domain, ShouldEqual, "auth.two.co.uk")
				So(domains[0].IsCustom, ShouldBeTrue)

				var apexDomain string
				err = db.QueryRow(`SELECT apex_domain FROM _portal_domain WHERE domain = 'auth.two.co.uk'`).Scan(&apexDomain)
				So(err, ShouldBeNil)
				So(apexDomain, ShouldEqual, "two.co.uk")

				// Only the pending entry of this domain is removed.
				So(countTestRows(t, db, "_portal_pending_domain"), ShouldEqual, 1)

				// Only the resources of this domain are applied.
				So(listTestResourceNames(t, kube, ingressGVR), ShouldResemble, []string{"traefik-auth.two.co.uk"})
				ingress, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-auth.two.co.uk", metav1.GetOptions{})
				So(err, ShouldBeNil)
				So(ingress.GetLabels()[service.LabelDomainID], ShouldEqual, domains[0].DomainID)
			})

			Convey("uses the existing domain of the app", func() {
				insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)

				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app1", Domain: "auth.one.com"}, &out)
				So(err, ShouldBeNil)
				So(out.String(), ShouldStartWith, "domain exists: auth.one.com (app: app1)\n")
				So(countTestRows(t, db, "_portal_domain"), ShouldEqual, 1)

				ingress, err := kube.DynamicClient.Resource(ingressGVR).Namespace(kube.Namespace).Get(ctx, "traefik-auth.one.com", metav1.GetOptions{})
				So(err, ShouldBeNil)
				So(ingress.GetLabels()[service.LabelDomainID], ShouldEqual, service.DomainModelID(true, "d2"))
			})

			Convey("uses the given apex domain", func() {
				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app1", Domain: "auth.localhost", ApexDomain: "auth.localhost"}, &out)
				So(err, ShouldBeNil)

				var apexDomain string
				err = db.QueryRow(`SELECT apex_domain FROM _portal_domain WHERE domain = 'auth.localhost'`).Scan(&apexDomain)
				So(err, ShouldBeNil)
				So(apexDomain, ShouldEqual, "auth.localhost")
			})

			Convey("fails when the apex domain cannot be derived", func() {
				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app1", Domain: "localhost"}, &out)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "specify it with --apex-domain")
				So(countTestRows(t, db, "_portal_domain"), ShouldEqual, 0)
			})

			Convey("fails when another app has the domain", func() {
				insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)

				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app2", Domain: "auth.one.com"}, &out)
				So(err, ShouldBeError, "domain auth.one.com (apex domain one.com) conflicts with existing domains: [auth.one.com (app: app1)]")
				So(countTestRows(t, db, "_portal_domain"), ShouldEqual, 1)
				So(listTestResourceNames(t, kube, ingressGVR), ShouldBeEmpty)
			})

			Convey("fails when a domain with the same apex domain exists, in another app or the same app", func() {
				insertTestDomain(t, db, "_portal_domain", "d2", "app1", "auth.one.com", "one.com", true)

				err := applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app2", Domain: "login.one.com"}, &out)
				So(err, ShouldBeError, "domain login.one.com (apex domain one.com) conflicts with existing domains: [auth.one.com (app: app1)]")

				err = applyDomainK8SResources(ctx, db, kube, ApplyDomainK8SResourcesOptions{AppID: "app1", Domain: "login.one.com"}, &out)
				So(err, ShouldBeError, "domain login.one.com (apex domain one.com) conflicts with existing domains: [auth.one.com (app: app1)]")

				So(countTestRows(t, db, "_portal_domain"), ShouldEqual, 1)
			})

			Convey("in a dry run, does not keep the domain", func() {
				insertTestDomain(t, db, "_portal_pending_domain", "p1", "app2", "auth.two.com", "two.com", true)

				domains, err := ensureCustomDomain(ctx, db, ApplyDomainK8SResourcesOptions{AppID: "app2", Domain: "auth.two.com", DryRun: true}, &out)
				So(err, ShouldBeNil)
				So(out.String(), ShouldEqual, "domain created: auth.two.com (app: app2, apex domain: two.com) (dry run, rolled back)\n")
				So(domains, ShouldHaveLength, 1)
				So(domains[0].Domain, ShouldEqual, "auth.two.com")

				So(countTestRows(t, db, "_portal_domain"), ShouldEqual, 0)
				So(countTestRows(t, db, "_portal_pending_domain"), ShouldEqual, 1)
			})
		})
	})
}
