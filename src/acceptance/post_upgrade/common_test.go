package post_upgrade_test

import (
	"code.cloudfoundry.org/app-autoscaler/src/acceptance/helpers"
	"strings"
)

func GetAppInfo(org, space, appType string) (fullAppName string, appGuid string) {
	apps := helpers.GetApps(cfg, org, space, "autoscaler-")
	for _, app := range apps {
		if strings.Contains(app, appType) {
			return app, helpers.GetAppGuid(cfg, app)
		}
	}
	return "", ""
}
