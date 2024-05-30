package auth

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
)

func (a *Auth) BasicAuth(r *http.Request, appID string) error {
	username, password, parseOK := r.BasicAuth()

	if !parseOK {
		return ErrorAuthNotFound
	}

	a.logger.Info(fmt.Sprintf("credentials in authenticator: %p", &a.credentials))
	valid, err := a.credentials.Validate(r.Context(), appID, models.Credential{Username: username, Password: password})
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("credentials are not valid")
	}
	return nil
}
