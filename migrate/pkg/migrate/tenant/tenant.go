package tenant

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pq "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	nurl "net/url"
)

type App struct {
	Name        string
	DatabaseURL string
}

func GetMigrateApps(hostnameOverride string, db sq.BaseRunner, appFilterKey string, appFilterValue string) ([]App, error) {
	// get apps for migration
	configSchema := "app_config"
	sqlizer := sq.Select("app.name", "config.config #> '{app_config,database_url}'").
		From(configSchema + ".app").
		Join(configSchema + ".config ON app.config_id = config.id")

	if appFilterKey != "" {
		fmt.Printf("appFilterKey %v \n", appFilterKey)
		appFilterKeyColumn := pq.QuoteIdentifier(appFilterKey)
		sqlizer = sqlizer.Where(fmt.Sprintf(`"app".%s = $1`, appFilterKeyColumn), appFilterValue)
	}

	apps := []App{}
	rows, err := sqlizer.RunWith(db).Query()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	for rows.Next() {
		a := App{}
		err = rows.Scan(&a.Name, &a.DatabaseURL)
		if err != nil {
			fmt.Printf("err: %v \n", err)
			return nil, err
		}
		// Remove double quotes
		a.DatabaseURL = a.DatabaseURL[1 : len(a.DatabaseURL)-1]
		// Override host if needed
		if hostnameOverride != "" {
			u, err := nurl.Parse(a.DatabaseURL)
			if err != nil {
				return nil, err
			}
			port := u.Port()
			if port != "" {
				u.Host = fmt.Sprintf("%s:%s", hostnameOverride, port)
			} else {
				u.Host = hostnameOverride
			}
			a.DatabaseURL = u.String()
		}
		apps = append(apps, a)
	}

	return apps, nil
}
