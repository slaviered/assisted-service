package hostcommands

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/internal/dbc"
	"github.com/openshift/assisted-service/internal/host/hostutil"
	"github.com/openshift/assisted-service/models"
)

var _ = Describe("free_addresses", func() {
	ctx := context.Background()
	var host models.Host
	var db *gorm.DB
	var fCmd *freeAddressesCmd
	var id, clusterId strfmt.UUID
	var stepReply []*models.Step
	var stepErr error
	var dbName string

	BeforeEach(func() {
		db, dbName = dbc.PrepareTestDB()
		fCmd = NewFreeAddressesCmd(common.GetTestLog(), "quay.io/ocpmetal/free_addresses:latest")

		id = strfmt.UUID(uuid.New().String())
		clusterId = strfmt.UUID(uuid.New().String())
		host = hostutil.GenerateTestHost(id, clusterId, models.HostStatusInsufficient)
		host.Inventory = common.GenerateTestDefaultInventory()
		Expect(db.Create(&host).Error).ShouldNot(HaveOccurred())
	})

	It("happy flow", func() {
		stepReply, stepErr = fCmd.GetSteps(ctx, &host)
		Expect(stepReply).ToNot(BeNil())
		Expect(stepReply[0].StepType).To(Equal(models.StepTypeFreeNetworkAddresses))
		Expect(stepErr).ShouldNot(HaveOccurred())
	})

	It("Illegal inventory", func() {
		host.Inventory = "blah"
		stepReply, stepErr = fCmd.GetSteps(ctx, &host)
		Expect(stepReply).To(BeNil())
		Expect(stepErr).To(HaveOccurred())
	})

	It("Missing networks", func() {
		host.Inventory = "{}"
		stepReply, stepErr = fCmd.GetSteps(ctx, &host)
		Expect(stepReply).To(BeNil())
		Expect(stepErr).To(HaveOccurred())
	})

	AfterEach(func() {
		// cleanup
		dbc.DeleteTestDB(db, dbName)
		stepReply = nil
		stepErr = nil
	})
})
