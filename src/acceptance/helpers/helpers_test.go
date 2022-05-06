package helpers_test

import (
	"acceptance/config"
	"acceptance/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServicePlans_urlIsCorrect(t *testing.T) {
	url := helpers.ServicePlansUrl(&config.Config{ServiceName: "autoscaler", ServiceBroker: "autoscaler"}, "GUID_UUID")
	assert.Equal(t, url, "/v3/service_plans?available=true&fields%5Bservice_offering.service_broker%5D=name%2Cguid&include=service_offering&per_page=5000&service_broker_names=autoscaler&service_offering_names=autoscaler&space_guids=GUID_UUID")
}
