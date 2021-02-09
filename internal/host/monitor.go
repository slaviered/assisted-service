package host

import (
	"context"

	"github.com/go-openapi/swag"
	"github.com/openshift/assisted-service/models"
	"github.com/openshift/assisted-service/pkg/requestid"
	"github.com/thoas/go-funk"
)

func (m *Manager) StopMonitoring(h *models.Host) bool {
	stopMonitoringStates := []string{string(models.LogsStateCompleted), string(models.LogsStateTimeout), ""}
	return ((swag.StringValue(h.Status) == models.HostStatusError || swag.StringValue(h.Status) == models.HostStatusCancelled) &&
		funk.Contains(stopMonitoringStates, h.LogsInfo))
}

func (m *Manager) HostMonitoring() {
	if !m.leaderElector.IsLeader() {
		m.log.Debugf("Not a leader, exiting HostMonitoring")
		return
	}
	m.log.Debugf("Running HostMonitoring")
	var (
		offset    int
		limit     = m.Config.MonitorBatchSize
		requestID = requestid.NewID()
		ctx       = requestid.ToContext(context.Background(), requestID)
		log       = requestid.RequestIDLogger(m.log, requestID)
	)

	monitorStates := []string{
		models.HostStatusDiscovering,
		models.HostStatusKnown,
		models.HostStatusDisconnected,
		models.HostStatusInsufficient,
		models.HostStatusPendingForInput,
		models.HostStatusPreparingForInstallation,
		models.HostStatusInstalling,
		models.HostStatusInstallingInProgress,
		models.HostStatusInstalled,
		models.HostStatusInstallingPendingUserAction,
		models.HostStatusResettingPendingUserAction,
	}
	for {
		//for offset = 0; offset < count; offset += limit {
		hosts := make([]*models.Host, 0, limit)
		if err := m.db.Where("status IN (?)", monitorStates).Offset(offset).Limit(limit).
			Order("cluster_id, id").Find(&hosts).Error; err != nil {
			log.WithError(err).Errorf("failed to get hosts")
			return
		}
		if len(hosts) == 0 {
			break
		}
		for _, host := range hosts {
			if !m.StopMonitoring(host) {
				if !m.leaderElector.IsLeader() {
					m.log.Debugf("Not a leader, exiting HostMonitoring")
					return
				}
				if err := m.RefreshStatus(ctx, host, m.db); err != nil {
					log.WithError(err).Errorf("failed to refresh host %s state", *host.ID)
				}
			}
		}
		offset += limit
	}
}
