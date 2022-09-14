package brokerserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/plancheck"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/policyvalidator"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/schedulerutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cred_helper"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/lager"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)
var _ domain.ServiceBroker = &Broker{}
type Broker struct {
	logger          lager.Logger
	conf            *config.Config
	bindingdb       db.BindingDB
	policydb        db.PolicyDB
	policyValidator *policyvalidator.PolicyValidator
	schedulerUtil   *schedulerutil.SchedulerUtil
	catalog         []domain.Service
	PlanChecker     plancheck.PlanChecker
	cfClient        cf.CFClient
	credentials     cred_helper.Credentials
}

var (
	emptyJSONObject                  = regexp.MustCompile(`^\s*{\s*}\s*$`)
	errorCreatingServiceBinding      = errors.New("error creating service binding")
	errorUpdatingServiceInstance     = errors.New("error updating service instance")
	errorDeleteSchedulesForUnbinding = errors.New("failed to delete schedules for unbinding")
	errorBindingDoesNotExist         = errors.New("service binding does not exist")
	errorDeletePolicyForUnbinding    = errors.New("failed to delete policy for unbinding")
	errorDeleteServiceBinding        = errors.New("error deleting service binding")
	errorCredentialNotDeleted        = errors.New("failed to delete custom metrics credential for unbinding")
)

func NewBroker(logger lager.Logger, conf *config.Config, bindingdb db.BindingDB, policydb db.PolicyDB, catalog []domain.Service, cfClient cf.CFClient, credentials cred_helper.Credentials) *Broker {
	broker := &Broker{
		logger:          logger,
		conf:            conf,
		bindingdb:       bindingdb,
		policydb:        policydb,
		catalog:         catalog,
		policyValidator: policyvalidator.NewPolicyValidator(conf.PolicySchemaPath, conf.ScalingRules.CPU.LowerThreshold, conf.ScalingRules.CPU.UpperThreshold),
		schedulerUtil:   schedulerutil.NewSchedulerUtil(conf, logger),
		PlanChecker:     plancheck.NewPlanChecker(conf.PlanCheck, logger),
		cfClient:        cfClient,
		credentials:     credentials,
	}
	return broker
}

// Services gets the catalog of services offered by the service broker
// GET /v2/catalog
func (b *Broker) Services(_ context.Context) ([]domain.Service, error) {
	return b.catalog, nil
}

// Provision creates a new service instance
// PUT /v2/service_instances/{instance_id}
func (b *Broker) Provision(_ context.Context, instanceID string, details domain.ProvisionDetails, _ bool) (domain.ProvisionedServiceSpec, error) {
	result := domain.ProvisionedServiceSpec{}

	logger := b.logger.Session("provision", lager.Data{"instanceID": instanceID, "provisionDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	if instanceID == "" || details.OrganizationGUID == "" || details.SpaceGUID == "" || details.ServiceID == "" || details.PlanID == "" {
		err := errors.New("failed to create service instance when trying to get mandatory data")
		logger.Error("check-for-mandatory-data", err)
		return result, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "check-for-mandatory-data")
	}

	parameters, err := parseInstanceParameters(details.RawParameters)
	if err != nil {
		return result, err
	}

	var policyJson json.RawMessage
	if parameters.DefaultPolicy != nil {
		policyJson = *parameters.DefaultPolicy
	}

	policyStr, policyGuidStr, err := b.getPolicyFromJsonRawMessage(policyJson, instanceID, details.PlanID)
	if err != nil {
		return result, err
	}

	err = b.bindingdb.CreateServiceInstance(models.ServiceInstance{ServiceInstanceId: instanceID, OrgId: details.OrganizationGUID, SpaceId: details.SpaceGUID, DefaultPolicy: policyStr, DefaultPolicyGuid: policyGuidStr})
	switch {
	case err == nil:
		result.DashboardURL = GetDashboardURL(b.conf, instanceID)
	case errors.Is(err, db.ErrAlreadyExists):
		logger.Error("failed to create service instance: service instance already exists", err, lager.Data{"instanceID": instanceID, "orgGuid": details.OrganizationGUID, "spaceGuid": details.SpaceGUID})
		result.DashboardURL = GetDashboardURL(b.conf, instanceID)
		result.AlreadyExists = true
		err = nil
	case errors.Is(err, db.ErrConflict):
		logger.Error("failed to create service instance: conflicting service instance exists", err, lager.Data{"instanceID": instanceID, "orgGuid": details.OrganizationGUID, "spaceGuid": details.SpaceGUID})
		err = apiresponses.ErrInstanceAlreadyExists
	default:
		logger.Error("failed to create service instance", err, lager.Data{"instanceID": instanceID, "orgGuid": details.OrganizationGUID, "spaceGuid": details.SpaceGUID})
		err = apiresponses.NewFailureResponse(errors.New("error creating service instance"), http.StatusInternalServerError, "failed to create service instance")
	}
	return result, err
}

func (b *Broker) getPolicyFromJsonRawMessage(policyJson json.RawMessage, instanceID string, planID string) (string, string, error) {
	var (
		policyGuidStr string
		err           error
	)
	policyStr := string(policyJson)
	if policyStr != "" {
		policyStr, err = b.validateAndCheckPolicy(policyStr, instanceID, planID)
		if err != nil {
			return "", "", err
		}

		policyGuid, err := uuid.NewV4()
		if err != nil {
			b.logger.Error("get-default-policy-create-guid", err, lager.Data{"instanceID": instanceID})
			return "", "", apiresponses.NewFailureResponse(errors.New("error generating policy guid"), http.StatusInternalServerError, "get-default-policy-create-guid")
		}
		policyGuidStr = policyGuid.String()
	}
	return policyStr, policyGuidStr, nil
}

func (b *Broker) validateAndCheckPolicy(policyStr string, instanceID string, planID string) (string, error) {
	errResults, valid, validatedPolicy := b.policyValidator.ValidatePolicy(policyStr)
	logger := b.logger.Session("validate-and-check-policy", lager.Data{"instanceID": instanceID, "policy": policyStr, "planID": planID, "errResults": errResults})

	if !valid {
		logger.Info("got-invalid-default-policy")
		resultsJson, err := json.Marshal(errResults)
		if err != nil {
			logger.Error("failed-marshalling-errors", err)
		}
		return "", apiresponses.NewFailureResponse(fmt.Errorf("invalid policy provided: %s", string(resultsJson)), http.StatusBadRequest, "failed-to-validate-policy")
	}
	policyStr = validatedPolicy

	if err := b.planDefinitionExceeded(policyStr, planID, instanceID); err != nil {
		return "", err
	}
	return policyStr, nil
}

// Deprovision deletes an existing service instance
// DELETE /v2/service_instances/{instance_id}
func (b *Broker) Deprovision(_ context.Context, instanceID string, details domain.DeprovisionDetails, _ bool) (domain.DeprovisionServiceSpec, error) {
	result := domain.DeprovisionServiceSpec{}

	logger := b.logger.Session("deprovision", lager.Data{"instanceID": instanceID, "deprovisionDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	serviceInstanceDeletionError := errors.New("error deleting service instance")
	// fetch and delete service bindings
	bindingIds, err := b.bindingdb.GetBindingIdsByInstanceId(instanceID)
	if err != nil {
		logger.Error("list-bindings-of-service-instance-to-be-deleted", err)
		return result, apiresponses.NewFailureResponse(serviceInstanceDeletionError, http.StatusInternalServerError, "list-bindings-of-service-instance-to-be-deleted")
	}

	for _, bindingId := range bindingIds {
		err = b.deleteBinding(bindingId, instanceID)
		wrappedError := fmt.Errorf("service binding deletion failed: %w", err)
		if err != nil && (errors.Is(err, errorDeleteServiceBinding) ||
			errors.Is(err, errorDeletePolicyForUnbinding) ||
			errors.Is(err, errorDeleteSchedulesForUnbinding) ||
			errors.Is(err, errorCredentialNotDeleted)) {
			logger.Error("delete-bindings-of-service-instance-to-be-deleted", wrappedError)
			return result, apiresponses.NewFailureResponse(serviceInstanceDeletionError, http.StatusInternalServerError, "delete-bindings-of-service-instance-to-be-deleted")
		}
	}

	err = b.bindingdb.DeleteServiceInstance(instanceID)
	if err != nil {
		if errors.Is(err, db.ErrDoesNotExist) {
			logger.Error("failed to delete service instance: service instance does not exist", err)
			return result, apiresponses.ErrInstanceDoesNotExist
		}
		logger.Error("delete-service-instance", err)
		return result, apiresponses.NewFailureResponse(serviceInstanceDeletionError, http.StatusInternalServerError, "delete-service-instance")
	}

	return result, nil
}

// GetInstance fetches information about a service instance
// GET /v2/service_instances/{instance_id}
func (b *Broker) GetInstance(_ context.Context, instanceID string, details domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	logger := b.logger.Session("get-instance", lager.Data{"instanceID": instanceID, "fetchInstanceDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	err := errors.New("error: get-instance is not implemented and this call should not have been allowed as instances_retrievable should be set to false")
	logger.Error("get-instance-is-not-implemented", err)
	return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "get-instance-is-not-implemented")
}

// Update modifies an existing service instance
// PATCH /v2/service_instances/{instance_id}
func (b *Broker) Update(_ context.Context, instanceID string, details domain.UpdateDetails, _ bool) (domain.UpdateServiceSpec, error) {
	logger := b.logger.Session("update", lager.Data{"instanceID": instanceID, "updateDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	result := domain.UpdateServiceSpec{}

	serviceInstance, err := b.getServiceInstance(instanceID)
	if err != nil {
		return result, err
	}

	servicePlan, servicePlanIsNew, err := b.getExistingOrUpdatedServicePlan(instanceID, details)
	if err != nil {
		return result, err
	}

	parameters, err := parseInstanceParameters(details.RawParameters)
	if err != nil {
		return result, err
	}

	// determine new default policy if any
	defaultPolicy, defaultPolicyGuid, defaultPolicyIsNew, err := b.determineDefaultPolicy(parameters, serviceInstance, servicePlan)
	if err != nil {
		return result, err
	}

	if !(servicePlanIsNew || defaultPolicyIsNew) {
		logger.Info("no-changes-requested")
		return result, nil
	}

	logger.Info("update-service-instance", lager.Data{"instanceID": instanceID, "serviceId": details.ServiceID, "planId": details.PlanID, "defaultPolicy": defaultPolicy})

	allBoundApps, err := b.bindingdb.GetAppIdsByInstanceId(serviceInstance.ServiceInstanceId)
	if err != nil {
		logger.Error("failed to retrieve bound apps", err, lager.Data{"instanceID": instanceID})
		return result, apiresponses.NewFailureResponse(errorUpdatingServiceInstance, http.StatusInternalServerError, "failed to retrieve bound apps")
	}

	if servicePlanIsNew {
		if err := b.checkScalingPoliciesUnderNewPlan(allBoundApps, servicePlan, instanceID); err != nil {
			return result, err
		}
	}

	if defaultPolicyIsNew {
		if err := b.applyDefaultPolicyUpdate(allBoundApps, serviceInstance, defaultPolicy, defaultPolicyGuid); err != nil {
			return result, err
		}

		// persist the changes to the default policy
		// NOTE: As the plan is not persisted, we do not need to this if we are only performing a plan change!
		updatedServiceInstance := models.ServiceInstance{
			ServiceInstanceId: serviceInstance.ServiceInstanceId,
			OrgId:             serviceInstance.OrgId,
			SpaceId:           serviceInstance.SpaceId,
			DefaultPolicy:     defaultPolicy,
			DefaultPolicyGuid: defaultPolicyGuid,
		}

		err = b.bindingdb.UpdateServiceInstance(updatedServiceInstance)
		if err != nil {
			logger.Error("failed to update service instance", err, lager.Data{"instanceID": instanceID})
			return result, apiresponses.NewFailureResponse(errorUpdatingServiceInstance, http.StatusInternalServerError, "update service instance")
		}
	}

	return result, nil
}

func (b *Broker) applyDefaultPolicyUpdate(allBoundApps []string, serviceInstance *models.ServiceInstance, defaultPolicy string, defaultPolicyGuid string) error {
	if defaultPolicy == "" {
		// default policy was present and will now be removed
		if err := b.removeDefaultPolicyFromApps(serviceInstance); err != nil {
			return err
		}
	} else {
		// a new default policy needs to be set
		if err := b.setDefaultPolicyOnApps(defaultPolicy, defaultPolicyGuid, allBoundApps, serviceInstance); err != nil {
			return err
		}
	}
	return nil
}

func parseInstanceParameters(rawParameters json.RawMessage) (*models.InstanceParameters, error) {
	parameters := &models.InstanceParameters{}
	if rawParameters != nil {
		err := json.Unmarshal(rawParameters, parameters)
		if err != nil {
			return nil, apiresponses.ErrRawParamsInvalid
		}
	}
	return parameters, nil
}

func (b *Broker) getServiceInstance(instanceID string) (*models.ServiceInstance, error) {
	serviceInstance, err := b.bindingdb.GetServiceInstance(instanceID)
	if err != nil {
		if errors.Is(err, db.ErrDoesNotExist) {
			b.logger.Error("failed to find service instance to update", err, lager.Data{"instanceID": instanceID})
			return nil, apiresponses.ErrInstanceDoesNotExist
		} else {
			b.logger.Error("failed to retrieve service instance", err, lager.Data{"instanceID": instanceID})
			return nil, apiresponses.NewFailureResponse(errors.New("failed to retrieve service instance"), http.StatusInternalServerError, "retrieving-instance-for-update")
		}
	}
	return serviceInstance, nil
}

func (b *Broker) setDefaultPolicyOnApps(updatedDefaultPolicy string, updatedDefaultPolicyGuid string, allBoundApps []string, serviceInstance *models.ServiceInstance) error {
	instanceID := serviceInstance.ServiceInstanceId
	b.logger.Info("update-service-instance-set-or-update", lager.Data{"instanceID": instanceID, "updatedDefaultPolicy": updatedDefaultPolicy, "updatedDefaultPolicyGuid": updatedDefaultPolicyGuid, "allBoundApps": allBoundApps, "serviceInstance": serviceInstance})

	updatedAppIds, err := b.policydb.SetOrUpdateDefaultAppPolicy(allBoundApps, serviceInstance.DefaultPolicyGuid, updatedDefaultPolicy, updatedDefaultPolicyGuid)
	if err != nil {
		b.logger.Error("failed to set default policies", err, lager.Data{"instanceID": instanceID})
		return apiresponses.NewFailureResponse(errors.New("failed to set default policy"), http.StatusInternalServerError, "updating-default-policy")
	}

	// there is synchronization between policy and schedule, so errors creating schedules should not break
	// the whole update process
	for _, appId := range updatedAppIds {
		if err = b.schedulerUtil.CreateOrUpdateSchedule(appId, updatedDefaultPolicy, updatedDefaultPolicyGuid); err != nil {
			b.logger.Error("failed to create/update schedules", err, lager.Data{"appId": appId, "policyGuid": updatedDefaultPolicyGuid, "policy": updatedDefaultPolicy})
		}
	}
	return nil
}

func (b *Broker) removeDefaultPolicyFromApps(serviceInstance *models.ServiceInstance) error {
	updatedAppIds, err := b.policydb.DeletePoliciesByPolicyGuid(serviceInstance.DefaultPolicyGuid)
	if err != nil {
		b.logger.Error("failed to delete default policies", err, lager.Data{"instanceID": serviceInstance.ServiceInstanceId})
		return apiresponses.NewFailureResponse(errors.New("failed to delete default policy"), http.StatusInternalServerError, "deleting-default-policy")
	}
	// there is synchronization between policy and schedule, so errors deleting schedules should not break
	// the whole update process
	for _, appId := range updatedAppIds {
		if err = b.schedulerUtil.DeleteSchedule(appId); err != nil {
			b.logger.Error("failed to delete schedules", err, lager.Data{"appId": appId})
		}
	}
	return nil
}

func (b *Broker) checkScalingPoliciesUnderNewPlan(allBoundApps []string, servicePlan string, instanceID string) error {
	var existingPolicy *models.ScalingPolicy
	var existingPolicyByteArray []byte
	var err error
	for _, appId := range allBoundApps {
		existingPolicy, err = b.policydb.GetAppPolicy(appId)
		if err != nil {
			b.logger.Error("failed to retrieve policy from db", err, lager.Data{"appId": appId})
			return apiresponses.NewFailureResponse(errorUpdatingServiceInstance, http.StatusInternalServerError, "failed to retrieve policy from db")
		}
		existingPolicyByteArray, err = json.Marshal(existingPolicy)
		if err != nil {
			b.logger.Error("failed to marshal policy from db", err, lager.Data{"appId": appId})
			return apiresponses.NewFailureResponse(errorUpdatingServiceInstance, http.StatusInternalServerError, "failed to marshal policy from db")
		}
		existingPolicyStr := string(existingPolicyByteArray)
		if err := b.planDefinitionExceeded(existingPolicyStr, servicePlan, instanceID); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) determineDefaultPolicy(parameters *models.InstanceParameters, serviceInstance *models.ServiceInstance, planID string) (string, string, bool, error) {
	defaultPolicy := serviceInstance.DefaultPolicy
	defaultPolicyGuid := serviceInstance.DefaultPolicyGuid
	defaultPolicyIsNew := false
	var err error

	if parameters.DefaultPolicy == nil {
		return defaultPolicy, defaultPolicyGuid, false, nil
	}

	newDefaultPolicy := string(*parameters.DefaultPolicy)
	if emptyJSONObject.MatchString(newDefaultPolicy) {
		// accept an empty json object "{}" as a default policy update to specify the removal of the default policy
		if defaultPolicy != "" {
			defaultPolicy = ""
			defaultPolicyGuid = ""
			defaultPolicyIsNew = true
		}
	} else {
		if newDefaultPolicy != defaultPolicy {
			newDefaultPolicy, err = b.validateAndCheckPolicy(newDefaultPolicy, serviceInstance.ServiceInstanceId, planID)
			if err != nil {
				return "", "", false, err
			}

			policyGuid, err := uuid.NewV4()
			if err != nil {
				b.logger.Error("determine-default-policy-create-guid", err, lager.Data{"instanceID": serviceInstance.ServiceInstanceId})
				return "", "", false, apiresponses.NewFailureResponse(errors.New("failed to create policy guid"), http.StatusInternalServerError, "determine-default-policy-create-guidz")
			}
			defaultPolicy = newDefaultPolicy
			defaultPolicyGuid = policyGuid.String()
			defaultPolicyIsNew = true
		}
	}

	return defaultPolicy, defaultPolicyGuid, defaultPolicyIsNew, nil
}

// LastOperation fetches last operation state for a service instance
// GET /v2/service_instances/{instance_id}/last_operation
func (b *Broker) LastOperation(_ context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	logger := b.logger.Session("last-operation", lager.Data{"instanceID": instanceID, "pollDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	err := errors.New("error: last-operation is not implemented and this endpoint should not have been called as all broker operations are synchronous")
	logger.Error("last-operation-is-not-implemented", err)
	return domain.LastOperation{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "last-operation-is-not-implemented")
}

// Bind creates a new service binding
// PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *Broker) Bind(_ context.Context, instanceID string, bindingID string, details domain.BindDetails, _ bool) (domain.Binding, error) {
	logger := b.logger.Session("bind", lager.Data{"instanceID": instanceID, "bindingID": bindingID, "bindDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	result := domain.Binding{}
	appGUID := details.AppGUID

	if appGUID == "" {
		err := errors.New("error: service must be bound to an application - service key creation is not supported")
		logger.Error("check-required-app-guid", err)
		return result, apiresponses.NewFailureResponseBuilder(err, http.StatusUnprocessableEntity, "check-required-app-guid").WithErrorKey("RequiresApp").Build()
	}

	var policyJson json.RawMessage
	if details.RawParameters != nil {
		policyJson = details.RawParameters
	}

	policyStr, policyGuidStr, err := b.getPolicyFromJsonRawMessage(policyJson, instanceID, details.PlanID)
	if err != nil {
		logger.Error("get-default-policy", err)
		return result, err
	}

	// fallback to default policy if no policy was provided
	if policyStr == "" {
		if serviceInstance, err := b.bindingdb.GetServiceInstance(instanceID); err != nil {
			logger.Error("get-service-instance", err)
			return result, apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "get-service-instance")
		} else {
			policyStr = serviceInstance.DefaultPolicy
			policyGuidStr = serviceInstance.DefaultPolicyGuid
		}
	}

	if err := b.handleExistingBindingsResiliently(instanceID, appGUID, logger); err != nil {
		return result, err
	}

	// create binding in DB
	err = b.bindingdb.CreateServiceBinding(bindingID, instanceID, appGUID)
	if err != nil {
		logger.Error("create-service-binding", err)
		if errors.Is(err, db.ErrAlreadyExists) {
			return result, apiresponses.NewFailureResponse(errors.New("error: an autoscaler service instance is already bound to the application and multiple bindings are not supported"), http.StatusConflict, "create-service-binding")
		}
		return result, apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "create-service-binding")
	}

	// create credentials
	cred, err := b.credentials.Create(appGUID, nil)
	if err != nil {
		//revert binding creating
		logger.Error("create-credentials", err)
		err = b.bindingdb.DeleteServiceBindingByAppId(appGUID)
		if err != nil {
			logger.Error("revert-binding-creation-due-to-credentials-creation-failure", err)
		}
		return result, apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "revert-binding-creation-due-to-credentials-creation-failure")
	}

	// attach policy to appGUID
	if err := b.attachPolicyToApp(appGUID, policyStr, policyGuidStr, logger); err != nil {
		return result, err
	}

	result.Credentials = models.Credentials{
		CustomMetrics: models.CustomMetricsCredentials{
			Credential: cred,
			URL:        b.conf.MetricsForwarder.MetricsForwarderUrl,
			MtlsUrl:    b.conf.MetricsForwarder.MetricsForwarderMtlsUrl,
		},
	}

	return result, nil
}

func (b *Broker) attachPolicyToApp(appGUID string, policyStr string, policyGuidStr string, logger lager.Logger) error {
	logger = logger.Session("saving-policy-json", lager.Data{"policy": policyStr})
	if policyStr == "" {
		logger.Info("no-policy-json-provided")
	} else {
		logger.Info("saving-policy-json")
		if err := b.policydb.SaveAppPolicy(appGUID, policyStr, policyGuidStr); err != nil {
			logger.Error("save-appGUID-policy", err)
			//failed to save policy, so revert creating binding and custom metrics credential
			err = b.credentials.Delete(appGUID)
			if err != nil {
				logger.Error("revert-custom-metrics-credential-due-to-failed-to-save-policy", err)
			}
			err = b.bindingdb.DeleteServiceBindingByAppId(appGUID)
			if err != nil {
				logger.Error("revert-binding-due-to-failed-to-save-policy", err)
			}
			return apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "save-appGUID-policy")
		}

		logger.Info("creating/updating schedules")
		if err := b.schedulerUtil.CreateOrUpdateSchedule(appGUID, policyStr, policyGuidStr); err != nil {
			//while there is synchronization between policy and schedule, so creating schedule error does not break
			//the whole creating binding process
			logger.Error("failed to create/update schedules", err)
		}
	}
	return nil
}

func (b *Broker) handleExistingBindingsResiliently(instanceID string, appGUID string, logger lager.Logger) error {
	// fetch and all service bindings for the service instance
	logger = logger.Session("handle-existing-bindings-resiliently")
	bindingIds, err := b.bindingdb.GetBindingIdsByInstanceId(instanceID)
	if err != nil {
		logger.Error("get-existing-service-bindings-before-binding", err)
		return apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "get-existing-service-bindings-before-binding")
	}

	for _, existingBindingId := range bindingIds {
		// get the service binding for the appGUID
		fetchedAppID, err := b.bindingdb.GetAppIdByBindingId(existingBindingId)
		if err != nil {
			logger.Error("get-existing-service-binding-before-binding", err, lager.Data{"existingBindingID": existingBindingId})
			return apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "get-existing-service-binding-before-binding")
		}

		//select the binding-id for the appGUID
		if fetchedAppID == appGUID {
			err = b.deleteBinding(existingBindingId, instanceID)
			wrappedError := fmt.Errorf("failed to bind service: %w", err)
			if err != nil && (errors.Is(err, errorDeleteServiceBinding) ||
				errors.Is(err, errorDeletePolicyForUnbinding) ||
				errors.Is(err, errorDeleteSchedulesForUnbinding) ||
				errors.Is(err, errorCredentialNotDeleted)) {
				logger.Error("delete-existing-service-binding-before-binding", wrappedError, lager.Data{"existingBindingID": existingBindingId})
				return apiresponses.NewFailureResponse(errorCreatingServiceBinding, http.StatusInternalServerError, "delete-existing-service-binding-before-binding")
			}
		}
	}
	return nil
}

// Unbind deletes an existing service binding
// DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *Broker) Unbind(_ context.Context, instanceID string, bindingID string, details domain.UnbindDetails, _ bool) (domain.UnbindSpec, error) {
	logger := b.logger.Session("unbind", lager.Data{"instanceID": instanceID, "bindingID": bindingID, "unbindDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	result := domain.UnbindSpec{}

	err := b.deleteBinding(bindingID, instanceID)

	if err != nil {
		logger.Error("delete-binding", fmt.Errorf("failed to unbind service: %w", err))

		if errors.Is(err, errorBindingDoesNotExist) {
			return result, apiresponses.ErrBindingDoesNotExist
		}
		return result, apiresponses.NewFailureResponse(errorDeleteServiceBinding, http.StatusInternalServerError, "delete-binding")
	}
	return result, nil
}

// GetBinding fetches an existing service binding
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (b *Broker) GetBinding(_ context.Context, instanceID string, bindingID string, details domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	logger := b.logger.Session("get-binding", lager.Data{"instanceID": instanceID, "bindingID": bindingID, "fetchBindingDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	err := errors.New("error: get-instance is not implemented and this call should not have been allowed as bindings_retrievable should be set to false")
	logger.Error("get-binding-is-not-implemented", err)
	return domain.GetBindingSpec{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "get-binding-is-not-implemented")
}

// LastBindingOperation fetches last operation state for a service binding
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
func (b *Broker) LastBindingOperation(_ context.Context, instanceID string, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	logger := b.logger.Session("last-binding-operation", lager.Data{"instanceID": instanceID, "bindingID": bindingID, "pollDetails": details})
	logger.Info("begin")
	defer logger.Info("end")

	err := errors.New("error: last-binding-operation is not implemented and this endpoint should not have been called as all broker operations are synchronous")
	logger.Error("last-binding-operation-is-not-implemented", err)
	return domain.LastOperation{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "last-binding-operation-is-not-implemented")
}

func (b *Broker) planDefinitionExceeded(policyStr string, planID string, instanceID string) error {
	policy := models.ScalingPolicy{}
	err := json.Unmarshal([]byte(policyStr), &policy)
	if err != nil {
		b.logger.Error("failed to unmarshal policy", err, lager.Data{"instanceID": instanceID, "policyStr": policyStr})
		return apiresponses.NewFailureResponse(errors.New("error reading policy"), http.StatusBadRequest, "failed to unmarshal policy")
	}
	ok, checkResult, err := b.PlanChecker.CheckPlan(policy, planID)
	if err != nil {
		b.logger.Error("failed to check policy for plan adherence", err, lager.Data{"instanceID": instanceID, "policyStr": policyStr})
		return apiresponses.NewFailureResponse(errors.New("error validating policy"), http.StatusInternalServerError, "failed to check policy for plan adherence")
	}
	if !ok {
		b.logger.Error("policy did not adhere to plan", fmt.Errorf(checkResult), lager.Data{"instanceID": instanceID, "policyStr": policyStr})
		return apiresponses.NewFailureResponse(fmt.Errorf("error: policy did not adhere to plan: %s", checkResult), http.StatusBadRequest, "policy did not adhere to plan")
	}
	return nil
}

func (b *Broker) getExistingOrUpdatedServicePlan(instanceID string, updateDetails domain.UpdateDetails) (string, bool, error) {
	existingServicePlan := updateDetails.PreviousValues.PlanID
	updateToPlan := updateDetails.PlanID

	servicePlan := existingServicePlan
	servicePlanIsNew := false

	var brokerErr error
	if updateToPlan != "" {
		servicePlanIsNew = servicePlan != updateToPlan
		servicePlan = updateToPlan
		if existingServicePlan != updateToPlan {
			isPlanUpdatable, err := b.PlanChecker.IsPlanUpdatable(existingServicePlan)
			if err != nil {
				b.logger.Error("checking-broker-plan-updatable", err, lager.Data{"instanceID": instanceID, "existingServicePlan": existingServicePlan, "newServicePlan": updateToPlan})
				brokerErr = apiresponses.NewFailureResponse(errors.New("error checking if the broker plan is updatable"), http.StatusInternalServerError, "checking-broker-plan-updatable")
			} else if !isPlanUpdatable {
				b.logger.Info("plan-not-updatable", lager.Data{"instanceID": instanceID, "existingServicePlan": existingServicePlan, "newServicePlan": updateToPlan})
				brokerErr = apiresponses.ErrPlanChangeNotSupported
			}
		}
	}

	return servicePlan, servicePlanIsNew, brokerErr
}

func GetDashboardURL(conf *config.Config, instanceID string) string {
	result := ""
	if conf.DashboardRedirectURI != "" {
		result = fmt.Sprintf("%s/manage/%s", conf.DashboardRedirectURI, instanceID)
	}

	return result
}

func (b *Broker) deleteBinding(bindingId string, serviceInstanceId string) error {
	appId, err := b.bindingdb.GetAppIdByBindingId(bindingId)
	if errors.Is(err, sql.ErrNoRows) {
		b.logger.Info("binding does not exist", nil, lager.Data{"instanceId": serviceInstanceId, "bindingId": bindingId})
		return errorBindingDoesNotExist
	}
	if err != nil {
		b.logger.Error("failed to get appId by bindingId", err, lager.Data{"instanceId": serviceInstanceId, "bindingId": bindingId})
		return errorDeleteServiceBinding
	}
	b.logger.Info("deleting policy json", lager.Data{"appId": appId})
	err = b.policydb.DeletePolicy(appId)
	if err != nil {
		b.logger.Error("failed to delete policy for unbinding", err, lager.Data{"appId": appId})
		return errorDeletePolicyForUnbinding
	}

	b.logger.Info("deleting schedules", lager.Data{"appId": appId})
	err = b.schedulerUtil.DeleteSchedule(appId)
	if err != nil {
		b.logger.Info("failed to delete schedules for unbinding", lager.Data{"appId": appId})
		return errorDeleteSchedulesForUnbinding
	}
	err = b.bindingdb.DeleteServiceBinding(bindingId)
	if err != nil {
		b.logger.Error("failed to delete binding", err, lager.Data{"bindingId": bindingId, "appId": appId})
		if errors.Is(err, db.ErrDoesNotExist) {
			return errorBindingDoesNotExist
		}

		return errorDeleteServiceBinding
	}

	err = b.credentials.Delete(appId)
	if err != nil {
		b.logger.Error("failed to delete custom metrics credential for unbinding", err, lager.Data{"appId": appId})
		return errorCredentialNotDeleted
	}

	return nil
}
