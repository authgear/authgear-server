package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/util/kubeutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type MigrateK8SToDBOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	KubeConfigPath string
	Namespace      string
}

func MigrateK8SToDB(opts *MigrateK8SToDBOptions) error {
	kubeConfig, err := kubeutil.MakeKubeConfig(opts.KubeConfigPath)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	db := openDB(opts.DatabaseURL, opts.DatabaseSchema)

	ctx := context.Background()

	listOptions := metav1.ListOptions{
		LabelSelector: service.LabelAppID,
	}
	secretList, err := client.CoreV1().Secrets(opts.Namespace).List(ctx, listOptions)
	if err != nil {
		return err
	}

	for _, secret := range secretList.Items {
		ss := secret
		err := WithTx(ctx, db, func(tx *sql.Tx) error {
			err := migrateFromSecretToDatabase(ctx, tx, &ss)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateFromSecretToDatabase(ctx context.Context, tx *sql.Tx, secret *corev1.Secret) error {
	data := make(map[string]string)
	for key, val := range secret.Data {
		data[key] = base64.StdEncoding.EncodeToString(val)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	appID := secret.ObjectMeta.Labels[service.LabelAppID]

	builder := newSQLBuilder().
		Insert(pq.QuoteIdentifier("_portal_config_source")).
		Columns(
			"id",
			"app_id",
			"data",
			"created_at",
			"updated_at",
		).
		Values(
			uuid.New(),
			appID,
			dataJSON,
			time.Now().UTC(),
			time.Now().UTC(),
		).Suffix("ON CONFLICT (app_id) DO UPDATE SET data = ?", dataJSON)

	q, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}
