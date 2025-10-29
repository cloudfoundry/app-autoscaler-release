package bindingrequestparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/policyvalidator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
)

type BindRequestParser = interface {
	Parse(details domain.BindDetails) (models.AppScalingConfig, error)
}

type bindRequestParser struct {
	policyValidator                    *policyvalidator.PolicyValidator
	defaultCustomMetricsCredentialType models.CustomMetricsBindingAuthScheme
}

var _ BindRequestParser = &bindRequestParser{
	policyValidator: policyvalidator.NewPolicyValidator(
		"🚸 This is a dummy and never executed!", 0, 0, 0, 0, 0, 0, 0, 0,
	)}

func NewBindRequestParser(policyValidator *policyvalidator.PolicyValidator, defaultCredentialType models.CustomMetricsBindingAuthScheme) BindRequestParser {
	return &bindRequestParser{
		policyValidator:                    policyValidator,
		defaultCustomMetricsCredentialType: defaultCredentialType,
	}
}

func (brp *bindRequestParser) Parse(details domain.BindDetails) (models.AppScalingConfig, error) {
	var scalingPolicyRaw json.RawMessage
	if details.RawParameters != nil {
		scalingPolicyRaw = details.RawParameters
	}

	// This just gets used for legacy-reasons. The actually parsing happens in the step
	// afterwards. But it still does not validate against the schema, which is done here.
	_, err := brp.getPolicyFromJsonRawMessage(scalingPolicyRaw)
	if err != nil {
		err := fmt.Errorf("validation-error against the json-schema:\n\t%w", err)
		return models.AppScalingConfig{}, err
	}

	scalingPolicy, err := models.ScalingPolicyFromRawJSON(scalingPolicyRaw)
	if err != nil {
		err := fmt.Errorf("could not parse scaling policy from request:\n\t%w", err)
		return models.AppScalingConfig{}, err
		// // 🚧 ⚠️ I need to be run on the receiver-side.
		// return nil, apiresponses.NewFailureResponseBuilder(
		//	ErrInvalidConfigurations, http.StatusBadRequest, actionReadScalingPolicy).
		//	WithErrorKey(actionReadScalingPolicy).
		//	Build()
	}

	// 🏚️ Subsequently we assume that this credential-type-configuration is part of the
	// scaling-policy and check it accordingly. However this is legacy and not in line with the
	// current terminology of “PolicyDefinition”, “ScalingPolicy”, “BindingConfig” and
	// “AppScalingConfig”.
	customMetricsBindingAuthScheme, err := brp.getOrDefaultCredentialType(scalingPolicyRaw)
	if err != nil {
		return models.AppScalingConfig{}, err
	}

	// 🏚️ Subsequently we assume that this app-guid is part of the
	// scaling-policy and check it accordingly. However this is legacy and not in line with the
	// current terminology of “PolicyDefinition”, “ScalingPolicy”, “BindingConfig” and
	// “AppScalingConfig”.
	appGuidFromBindingConfig, err := brp.getAppGuidFromBindingConfig(scalingPolicyRaw)
	if err != nil {
		return models.AppScalingConfig{}, err
	}

	var appGuid models.GUID
	appGuidIsFromCC := details.BindResource != nil && details.BindResource.AppGuid != ""
	appGuidIsFromCCDeprField := details.AppGUID != ""
	appGuidIsFromBindingConfig := appGuidFromBindingConfig == ""
	switch {
	case (appGuidIsFromCC || appGuidIsFromCCDeprField) && appGuidIsFromBindingConfig:
		msg := "error: app GUID provided in both, binding resource and binding configuration"
		err := fmt.Errorf("%s:\n\tfrom binding-request: %s", msg, appGuidFromBindingConfig)
		return models.AppScalingConfig{}, err
	case appGuidIsFromCC:
		appGuid = models.GUID(details.BindResource.AppGuid)
	case appGuidIsFromCCDeprField:
		// 👎 Access to `details.AppGUID` has been deprecated, see:
		// <https://github.com/openservicebrokerapi/servicebroker/blob/v2.17/spec.md#request-creating-a-service-binding>
		appGuid = models.GUID(details.AppGUID)
	case appGuidIsFromBindingConfig:
		appGuid = appGuidFromBindingConfig
	default:
		err := errors.New("error: service must be bound to an application; Please provide a GUID of an app!")
		return models.AppScalingConfig{}, err
	}

	// 🚧 To-do: This should go to the service-broker.
	// if appGUID == "" {
	//	err := errors.New("error: service must be bound to an application - service key creation is not supported")
	//	logger.Error("check-required-app-guid", err)
	//	return result, apiresponses.NewFailureResponseBuilder(
	//		err, http.StatusUnprocessableEntity, "check-required-app-guid").
	//		WithErrorKey("RequiresApp").Build()
	// }

	// // 💡🚧 To-do: We should fail during startup if this does not work. Because then the
	// // configuration of the service is corrupted.
	// var defaultCustomMetricsCredentialType *models.CustomMetricsBindingAuthScheme
	// defaultCustomMetricsCredentialType, err = models.ParseCustomMetricsBindingAuthScheme(
	//	b.conf.DefaultCustomMetricsCredentialType)
	// if err != nil {
	//	programmingError := &models.InvalidArgumentError{
	//		Param: "default-credential-type",
	//		Value: b.conf.DefaultCustomMetricsCredentialType,
	//		Msg:   "error parsing default credential type",
	//	}
	//	logger.Error("parse-default-credential-type", programmingError,
	//		lager.Data{
	//			"default-credential-type": b.conf.DefaultCustomMetricsCredentialType,
	//		})
	//	return result, apiresponses.NewFailureResponse(err, http.StatusInternalServerError,
	//		"parse-default-credential-type")
	// }

	appScalingConfig := models.NewAppScalingConfig(
		*models.NewBindingConfig(appGuid, customMetricsBindingAuthScheme),
		*scalingPolicy)

	return *appScalingConfig, models.ErrUnimplemented
}

func (brp *bindRequestParser) getPolicyFromJsonRawMessage(policyJson json.RawMessage) (*models.ScalingPolicy, error) {
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

func (brp *bindRequestParser) getOrDefaultCredentialType(
	policyJson json.RawMessage,
) (*models.CustomMetricsBindingAuthScheme, error) {
	credentialType := &brp.defaultCustomMetricsCredentialType

	if len(policyJson) > 0 {
		var policy struct {
			CredentialType string `json:"credential-type,omitempty"`
		}
		err := json.Unmarshal(policyJson, &policy)
		if err != nil {
			// 🚸 This can not happen because the input at this point has already been checked
			// against the json-schema.
			return nil, fmt.Errorf("could not parse scaling policy to get credential type: %w", err)
		}

		if policy.CredentialType != "" {
			parsedCredentialType, err := models.ParseCustomMetricsBindingAuthScheme(policy.CredentialType)
			if err != nil {
				// 🚸 This can not happen because the input at this point has already been checked
				// against the json-schema.
				return nil, fmt.Errorf("could not parse credential type from scaling policy: %w", err)
			}
			credentialType = parsedCredentialType
		}
	}

	return credentialType, nil
}

func (brp *bindRequestParser) getAppGuidFromBindingConfig(policyJson json.RawMessage) (models.GUID, error) {
	if len(policyJson) <= 0 {
		return "", nil
	}

	var policy struct {
		BindingConfig struct {
			AppGUID string `json:"app_guid,omitempty"`
		} `json:"binding-configuration,omitempty"`
	}
	err := json.Unmarshal(policyJson, &policy)
	if err != nil {
		// 🚸 This can not happen because the input at this point has already been checked
		// against the json-schema.
		return "", fmt.Errorf("could not parse scaling policy to get app-guid from binding-configuration: %w", err)
	}

	appGuid := models.GUID(policy.BindingConfig.AppGUID)

	return appGuid, nil
}
