package hostcommands

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/connectivity"
	"github.com/openshift/assisted-service/internal/dbc"
	"github.com/openshift/assisted-service/internal/events"
	"github.com/openshift/assisted-service/internal/hardware"
	"github.com/openshift/assisted-service/internal/host/hostutil"
	"github.com/openshift/assisted-service/internal/oc"
	"github.com/openshift/assisted-service/internal/versions"
	"github.com/openshift/assisted-service/models"
	"github.com/thoas/go-funk"
)

var _ = Describe("instruction_manager", func() {
	var (
		ctx               = context.Background()
		host              models.Host
		db                *gorm.DB
		mockEvents        *events.MockHandler
		mockVersions      *versions.MockHandler
		stepsReply        models.Steps
		hostId, clusterId strfmt.UUID
		stepsErr          error
		instMng           *InstructionManager
		ctrl              *gomock.Controller
		hwValidator       *hardware.MockValidator
		mockRelease       *oc.MockRelease
		cnValidator       *connectivity.MockValidator
		instructionConfig InstructionConfig
		dbName            string
	)

	BeforeEach(func() {
		db, dbName = dbc.PrepareTestDB()
		ctrl = gomock.NewController(GinkgoT())
		mockEvents = events.NewMockHandler(ctrl)
		mockVersions = versions.NewMockHandler(ctrl)
		hwValidator = hardware.NewMockValidator(ctrl)
		mockRelease = oc.NewMockRelease(ctrl)
		cnValidator = connectivity.NewMockValidator(ctrl)
		instMng = NewInstructionManager(common.GetTestLog(), db, hwValidator, mockRelease, instructionConfig, cnValidator, mockEvents, mockVersions)
		hostId = strfmt.UUID(uuid.New().String())
		clusterId = strfmt.UUID(uuid.New().String())
		host = hostutil.GenerateTestHost(hostId, clusterId, "unknown invalid state")
		host.Role = models.HostRoleMaster
		host.Inventory = hostutil.GenerateMasterInventory()
		Expect(db.Create(&host).Error).ShouldNot(HaveOccurred())
		anotherHost := hostutil.GenerateTestHost(strfmt.UUID(uuid.New().String()), clusterId, "insufficient")
		Expect(db.Create(&anotherHost).Error).ShouldNot(HaveOccurred())
	})

	checkStep := func(state string, expectedStepTypes []models.StepType) {
		checkStepsByState(state, &host, db, mockEvents, instMng, hwValidator, mockRelease, mockVersions, cnValidator, ctx, expectedStepTypes)
	}

	Context("No DHCP", func() {
		BeforeEach(func() {
			cluster := dbc.Cluster{Cluster: models.Cluster{ID: &clusterId, VipDhcpAllocation: swag.Bool(false), MachineNetworkCidr: "1.2.3.0/24"}}
			Expect(db.Create(&cluster).Error).ShouldNot(HaveOccurred())
		})
		Context("get_next_steps", func() {
			It("invalid_host_state", func() {
				stepsReply, stepsErr = instMng.GetNextSteps(ctx, &host)
				Expect(stepsReply.Instructions).To(HaveLen(0))
				Expect(stepsErr).Should(BeNil())
			})
			It("discovering", func() {
				checkStep(models.HostStatusDiscovering, []models.StepType{
					models.StepTypeInventory,
				})
			})
			It("known", func() {
				checkStep(models.HostStatusKnown, []models.StepType{
					models.StepTypeConnectivityCheck, models.StepTypeFreeNetworkAddresses,
					models.StepTypeInventory, models.StepTypeNtpSynchronizer,
				})
			})
			It("disconnected", func() {
				checkStep(models.HostStatusDisconnected, []models.StepType{
					models.StepTypeInventory,
				})
			})
			It("insufficient", func() {
				checkStep(models.HostStatusInsufficient, []models.StepType{
					models.StepTypeInventory, models.StepTypeConnectivityCheck,
					models.StepTypeFreeNetworkAddresses, models.StepTypeNtpSynchronizer,
				})
			})
			It("pending-for-input", func() {
				checkStep(models.HostStatusPendingForInput, []models.StepType{
					models.StepTypeInventory, models.StepTypeConnectivityCheck,
					models.StepTypeFreeNetworkAddresses, models.StepTypeNtpSynchronizer,
				})
			})
			It("error", func() {
				checkStep(models.HostStatusError, []models.StepType{
					models.StepTypeExecute, models.StepTypeExecute,
				})
			})
			It("error with already uploades logs", func() {
				host.LogsCollectedAt = strfmt.DateTime(time.Now())
				db.Save(&host)
				checkStep(models.HostStatusError, []models.StepType{
					models.StepTypeExecute,
				})
			})
			It("cancelled", func() {
				checkStep(models.HostStatusCancelled, []models.StepType{
					models.StepTypeExecute, models.StepTypeExecute,
				})
			})
			It("installing", func() {
				checkStep(models.HostStatusInstalling, []models.StepType{
					models.StepTypeInstall,
				})
			})
			It("reset", func() {
				checkStep(models.HostStatusResetting, []models.StepType{
					models.StepTypeResetInstallation,
				})
			})
		})
	})

	Context("With DHCP", func() {
		BeforeEach(func() {
			cluster := dbc.Cluster{Cluster: models.Cluster{ID: &clusterId, VipDhcpAllocation: swag.Bool(true), MachineNetworkCidr: "1.2.3.0/24"}}
			Expect(db.Create(&cluster).Error).ShouldNot(HaveOccurred())
		})
		Context("get_next_steps", func() {
			It("invalid_host_state", func() {
				stepsReply, stepsErr = instMng.GetNextSteps(ctx, &host)
				Expect(stepsReply.Instructions).To(HaveLen(0))
				Expect(stepsErr).Should(BeNil())
			})
			It("discovering", func() {
				checkStep(models.HostStatusDiscovering, []models.StepType{
					models.StepTypeInventory,
				})
			})
			It("known", func() {
				checkStep(models.HostStatusKnown, []models.StepType{
					models.StepTypeConnectivityCheck, models.StepTypeFreeNetworkAddresses,
					models.StepTypeDhcpLeaseAllocate, models.StepTypeInventory,
					models.StepTypeNtpSynchronizer,
				})
			})
			It("disconnected", func() {
				checkStep(models.HostStatusDisconnected, []models.StepType{
					models.StepTypeInventory,
				})
			})
			It("insufficient", func() {
				checkStep(models.HostStatusInsufficient, []models.StepType{
					models.StepTypeInventory, models.StepTypeConnectivityCheck,
					models.StepTypeFreeNetworkAddresses, models.StepTypeDhcpLeaseAllocate,
					models.StepTypeNtpSynchronizer,
				})
			})
			It("pending-for-input", func() {
				checkStep(models.HostStatusPendingForInput, []models.StepType{
					models.StepTypeInventory, models.StepTypeConnectivityCheck,
					models.StepTypeFreeNetworkAddresses, models.StepTypeDhcpLeaseAllocate,
					models.StepTypeNtpSynchronizer,
				})
			})
			It("error", func() {
				checkStep(models.HostStatusError, []models.StepType{
					models.StepTypeExecute, models.StepTypeExecute,
				})
			})
			It("cancelled", func() {
				checkStep(models.HostStatusCancelled, []models.StepType{
					models.StepTypeExecute, models.StepTypeExecute,
				})
			})
			It("installing", func() {
				checkStep(models.HostStatusInstalling, []models.StepType{
					models.StepTypeInstall, models.StepTypeDhcpLeaseAllocate,
				})
			})
			It("installing-in-progress", func() {
				checkStep(models.HostStatusInstallingInProgress, []models.StepType{
					models.StepTypeInventory, models.StepTypeDhcpLeaseAllocate,
				})
			})
			It("reset", func() {
				checkStep(models.HostStatusResetting, []models.StepType{
					models.StepTypeResetInstallation,
				})
			})
		})
	})

	AfterEach(func() {
		// cleanup
		dbc.DeleteTestDB(db, dbName)
		ctrl.Finish()
		stepsReply = models.Steps{}
		stepsErr = nil
	})

})

func checkStepsByState(state string, host *models.Host, db *gorm.DB, mockEvents *events.MockHandler,
	instMng *InstructionManager, mockValidator *hardware.MockValidator, mockRelease *oc.MockRelease, mockVersions *versions.MockHandler,
	mockConnectivity *connectivity.MockValidator, ctx context.Context, expectedStepTypes []models.StepType) {

	mockEvents.EXPECT().AddEvent(gomock.Any(), host.ClusterID, host.ID, hostutil.GetEventSeverityFromHostStatus(state), gomock.Any(), gomock.Any())
	updateReply, updateErr := hostutil.UpdateHostStatus(ctx, common.GetTestLog(), db, mockEvents, host.ClusterID, *host.ID, *host.Status, state, "")
	ExpectWithOffset(1, updateErr).ShouldNot(HaveOccurred())
	ExpectWithOffset(1, updateReply).ShouldNot(BeNil())
	h := hostutil.GetHostFromDB(*host.ID, host.ClusterID, db)
	ExpectWithOffset(1, swag.StringValue(h.Status)).Should(Equal(state))
	mockVersions.EXPECT().GetReleaseImage(gomock.Any()).Return(defaultReleaseImage, nil).AnyTimes()
	mockValidator.EXPECT().GetHostInstallationPath(gomock.Any()).Return("/dev/disk/by-id/wwn-sda").AnyTimes()
	mockRelease.EXPECT().GetMCOImage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(defaultMCOImage, nil).AnyTimes()
	mockRelease.EXPECT().GetMustGatherImage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(defaultMustGatherImage, nil).AnyTimes()
	if funk.Contains(expectedStepTypes, models.StepTypeConnectivityCheck) {
		mockConnectivity.EXPECT().GetHostValidInterfaces(gomock.Any()).Return([]*models.Interface{
			{
				Name: "eth0",
				IPV4Addresses: []string{
					"1.2.3.10/24",
				},
				MacAddress: "52:54:00:09:de:93",
			},
		}, nil).Times(1)
	}

	stepsReply, stepsErr := instMng.GetNextSteps(ctx, h)
	ExpectWithOffset(1, stepsReply.Instructions).To(HaveLen(len(expectedStepTypes)))
	if stateValues, ok := instMng.installingClusterStateToSteps[state]; ok {
		Expect(stepsReply.NextInstructionSeconds).Should(Equal(stateValues.NextStepInSec))
	} else {
		Expect(stepsReply.NextInstructionSeconds).Should(Equal(defaultNextInstructionInSec))
	}

	ExpectWithOffset(1, *stepsReply.PostStepAction).Should(Equal(models.StepsPostStepActionContinue))

	for i, step := range stepsReply.Instructions {
		ExpectWithOffset(1, step.StepType).Should(Equal(expectedStepTypes[i]))
	}
	ExpectWithOffset(1, stepsErr).ShouldNot(HaveOccurred())
}

func TestHostCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	dbc.InitializeDBTest()
	defer dbc.TerminateDBTest()
	RunSpecs(t, "Host commands test Suite")
}
