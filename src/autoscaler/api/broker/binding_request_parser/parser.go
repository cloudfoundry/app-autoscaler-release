package bindingrequestparser

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/policyvalidator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"code.cloudfoundry.org/lager/v3"
)

type BindRequestParser = interface {
	Parse(instanceID string, details domain.BindDetails) (models.AppScalingConfig, error)
}

type bindRequestParser struct {
	policyValidator *policyvalidator.PolicyValidator
}

var _ BindRequestParser = &bindRequestParser{
	policyValidator: policyvalidator.NewPolicyValidator(
		"🚸 This is a dummy and never executed!", 0, 0, 0, 0, 0, 0, 0, 0,
	)}

func (brp *bindRequestParser) Parse(instanceID string, details domain.BindDetails) (models.AppScalingConfig, error) {
	var scalingPolicyRaw json.RawMessage
	if details.RawParameters != nil {
		scalingPolicyRaw = details.RawParameters
	}

	// This just gets used for legacy-reasons. The actually parsing happens in the step
	// afterwards. But it still does not validate against the schema, which is done here.
	_, err := brp.getPolicyFromJsonRawMessage(scalingPolicyRaw, instanceID, details.PlanID)
	if err != nil {
		err := fmt.Errorf("validation-error against the json-schema:\n\t%w", err)
		return models.AppScalingConfig{}, err
	}

	scalingPolicy, err := models.ScalingPolicyFromRawJSON(scalingPolicyRaw)
	if err != nil {
		err := fmt.Errorf("could not parse scaling policy from request:\n\t%w", err)
		return models.AppScalingConfig{}, err
		// // ⚠️ I need to be run on the receiver-side.
		// return nil, apiresponses.NewFailureResponseBuilder(
		//	ErrInvalidConfigurations, http.StatusBadRequest, actionReadScalingPolicy).
		//	WithErrorKey(actionReadScalingPolicy).
		//	Build()
	}

	// // 🚧 To-do: Check if exactly one is provided. We don't want to accept both to be present.
	// requestAppGuid := details.BindResource.AppGuid
	// paramsAppGuid := bindingConfig.Configuration.AppGUID
	var appGUID string
	if details.BindResource != nil && details.BindResource.AppGuid != "" {
		appGUID = details.BindResource.AppGuid
	} else if details.AppGUID != "" {
		// 👎 Access to `details.AppGUID` has been deprecated, see:
		// <https://github.com/openservicebrokerapi/servicebroker/blob/v2.17/spec.md#request-creating-a-service-binding>
		appGUID = details.AppGUID
	} else {
		// 🚧 To-do: Implement feature: service-key-creation; Use appID from `bindingConfig`!
	}

	if appGUID == "" {
		err := errors.New("error: service must be bound to an application - service key creation is not supported")
		logger.Error("check-required-app-guid", err)
		return result, apiresponses.NewFailureResponseBuilder(
			err, http.StatusUnprocessableEntity, "check-required-app-guid").
			WithErrorKey("RequiresApp").Build()
	}

	// 💡🚧 To-do: We should fail during startup if this does not work. Because then the
	// configuration of the service is corrupted.
	var defaultCustomMetricsCredentialType *models.CustomMetricsBindingAuthScheme
	defaultCustomMetricsCredentialType, err = models.ParseCustomMetricsBindingAuthScheme(
		b.conf.DefaultCustomMetricsCredentialType)
	if err != nil {
		programmingError := &models.InvalidArgumentError{
			Param: "default-credential-type",
			Value: b.conf.DefaultCustomMetricsCredentialType,
			Msg:   "error parsing default credential type",
		}
		logger.Error("parse-default-credential-type", programmingError,
			lager.Data{
				"default-credential-type": b.conf.DefaultCustomMetricsCredentialType,
			})
		return result, apiresponses.NewFailureResponse(err, http.StatusInternalServerError,
			"parse-default-credential-type")
	}
	// 🏚️ Subsequently we assume that this credential-type-configuration is part of the
	// scaling-policy and check it accordingly. However this is legacy and not in line with the
	// current terminology of “PolicyDefinition”, “ScalingPolicy”, “BindingConfig” and
	// “AppScalingConfig”.
	customMetricsBindingAuthScheme, err := getOrDefaultCredentialType(scalingPolicyRaw,
		defaultCustomMetricsCredentialType, logger)
	if err != nil {
		return result, err
	}

	// To-do: 🚧 Factor everything that is involved in this creation out into an own
	// helper-function. Consider a function analogous to `getScalingPolicyFromRequest` that is
	// defined within this file.
	appScalingConfig := models.NewAppScalingConfig(
		*models.NewBindingConfig(models.GUID(appGUID), customMetricsBindingAuthScheme),
		*scalingPolicy)

	return models.AppScalingConfig{}, models.ErrUnimplemented
}

func (brp *bindRequestParser) getPolicyFromJsonRawMessage(policyJson json.RawMessage, instanceID string, planID string) (*models.ScalingPolicy, error) {
	if isEmptyPolicy := len(policyJson) <= 0; isEmptyPolicy { // no nil-check needed: `len(nil) == 0`
		return nil, nil
	}

	policy, errResults := brp.policyValidator.ParseAndValidatePolicy(policyJson)
	if errResults != nil {
		// 🚫 The subsequent log-message is a strong assumption about the context of the caller. But
		// how can we actually know here that we operate on a default-policy? In fact, when we are
		// in the call-stack of `Bind` then we are *not* called with a default-policy.
		resultsJson, err := json.Marshal(errResults)
		if err != nil {
			return nil, &models.InvalidArgumentError{
				Param: "errResults",
				Value: errResults,
				Msg:   "Failed to json-marshal validation results; This should never happen."}
		}
		return policy, apiresponses.NewFailureResponse(fmt.Errorf("invalid policy provided: %s", string(resultsJson)), http.StatusBadRequest, "failed-to-validate-policy")
	}

	return policy, nil
}

func (brp *bindRequestParser) getScalingPolicyFromRequest(
	scalingPolicyRaw json.RawMessage, logger lager.Logger,
) (*models.ScalingPolicy, error) {
}
