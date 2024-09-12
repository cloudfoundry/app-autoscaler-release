package auth

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/lager/v3"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrXFCCHeaderNotFound = errors.New("mTLS authentication method not found")
var ErrorNoAppIDFound = errors.New("certificate does not contain an app id")
var ErrorAppIDWrong = errors.New("app id in certificate is not valid")

var ErrorAppNotBound = errors.New("application is not bound to the same service instance")

var ErrorUnmarshallingBody = errors.New("error unmarshalling custom metrics request body")
var ErrorReadingBody = errors.New("error reading custom metrics request body")

func (a *Auth) XFCCAuth(r *http.Request, bindingDB db.BindingDB, appID string, appToScale string) error {
	xfccHeader := r.Header.Get("X-Forwarded-Client-Cert")
	if xfccHeader == "" {
		return ErrXFCCHeaderNotFound
	}

	data, err := base64.StdEncoding.DecodeString(removeQuotes(xfccHeader))
	if err != nil {
		return fmt.Errorf("base64 parsing failed: %w", err)
	}

	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	certAppId := getAppId(cert)

	if len(certAppId) == 0 {
		return ErrorNoAppIDFound
	}

	// appID = custom metrics producer
	// certAppId = app id in certificate
	// Case 1 : custom metrics can only be published by the app itself
	// Case 2 : custom metrics can be published by any app bound to the same autoscaler instance
	// if the requester is not same as the scaling app
	if appID != certAppId {
		return ErrorAppIDWrong
	}
	// check for case 2 here
	/*
		TODO
		Read the parameter new boolean parameter from the http request body named as "allow_from": "bound_app or same_app"
		If it is set to true, then
		- check if the app is bound to the same autoscaler instance - How to check this? check from the database binding_db table -> app_id->binding_id->service_instance_id-> all bound apps
		- if it is bound, then allow the request i.e custom metrics to be published
		- if it is not bound, then return an error saying "app is not allowed to send custom metrics on as it not bound to the autoscaler service instance"
		If the parameter is not set, then follow the existing logic and allow the request to be published


	*/
	a.logger.Info("Checking custom metrics submission strategy")
	customMetricSubmissionStrategy := r.Header.Get("custom-metrics-submission-strategy")
	customMetricSubmissionStrategy = strings.ToLower(customMetricSubmissionStrategy)
	if customMetricSubmissionStrategy == "bound_app" {
		a.logger.Info("custom-metrics-submission-strategy-found", lager.Data{"strategy": customMetricSubmissionStrategy})
		// check if the app is bound to same autoscaler instance by check the binding id from the bindingdb
		// if the app is bound to the same autoscaler instance, then allow the request to the next handler i.e publish custom metrics
		isAppBound, err := bindingDB.IsAppBoundToSameAutoscaler(r.Context(), appID, appToScale)
		if err != nil {
			a.logger.Error("error-checking-app-bound-to-same-service", err, lager.Data{"app-id": appID})
			return err
		}
		if isAppBound == false {
			a.logger.Info("app-not-bound-to-same-service", lager.Data{"app-id": appID})
			return ErrorAppNotBound
		}
	} /*  no need to check as this is the default case
	else if customMetricSubmissionStrategy == "same_app" || customMetricSubmissionStrategy == "" { // default case
		// if the app is the same app, then allow the request to the next handler i.e 403
		a.logger.Info("custom-metrics-submission-strategy", lager.Data{"strategy": customMetricSubmissionStrategy})
		return ErrorAppIDWrong
	} */

	return nil
}

func getAppId(cert *x509.Certificate) string {
	var certAppId string
	for _, ou := range cert.Subject.OrganizationalUnit {
		if strings.Contains(ou, "app:") {
			certAppId = strings.Split(ou, ":")[1]
			break
		}
	}
	return certAppId
}

func removeQuotes(xfccHeader string) string {
	if xfccHeader[0] == '"' {
		xfccHeader = xfccHeader[1 : len(xfccHeader)-1]
	}
	return xfccHeader
}
