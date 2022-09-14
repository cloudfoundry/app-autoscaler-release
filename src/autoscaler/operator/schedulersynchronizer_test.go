package operator_test

import (
	"net/http"
	"time"

	"github.com/cloudfoundry/app-autoscaler-release/operator"
	"github.com/cloudfoundry/app-autoscaler-release/routes"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ScheduleSynchronizer", func() {
	var (
		fakeSyncServer       *ghttp.Server
		buffer               *gbytes.Buffer
		scheduleSynchronizer *operator.ScheduleSynchronizer
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("schedule-synchoronizer-test")
		buffer = logger.Buffer()
		fclock := fakeclock.NewFakeClock(time.Now())
		fakeSyncServer = ghttp.NewServer()
		//nolint:staticcheck //TODO https://github.com/cloudfoundry/app-autoscaler-release/issues/549
		scheduleSynchronizer = operator.NewScheduleSynchronizer(cfhttp.NewClient(), fakeSyncServer.URL(), fclock, logger)

	})

	Describe("Sync", func() {
		JustBeforeEach(func() {
			scheduleSynchronizer.Operate()
		})

		Context("when sync server is available", func() {
			BeforeEach(func() {
				fakeSyncServer.RouteToHandler("PUT", routes.SyncActiveSchedulesPath, ghttp.RespondWith(http.StatusOK, "successful"))
			})
			It("raise sync request successfully", func() {
				Eventually(fakeSyncServer.ReceivedRequests).Should(HaveLen(1))
			})
		})

		Context("when sync server is not available", func() {
			BeforeEach(func() {
				fakeSyncServer.Close()
			})
			It("should error", func() {
				Eventually(buffer).Should(gbytes.Say("failed-to-send-sync-scheduler-request"))
			})
		})
	})
})
