package services

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// WhatsAppCleanupJob handles periodic cleanup of old audit logs
type WhatsAppCleanupJob struct {
	auditService  *WhatsAppAuditService
	retentionDays int
	interval      time.Duration
	stopChan      chan struct{}
}

// NewWhatsAppCleanupJob creates a new cleanup job
func NewWhatsAppCleanupJob(auditService *WhatsAppAuditService, retentionDays int, intervalHours int) *WhatsAppCleanupJob {
	return &WhatsAppCleanupJob{
		auditService:  auditService,
		retentionDays: retentionDays,
		interval:      time.Duration(intervalHours) * time.Hour,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the cleanup job
func (j *WhatsAppCleanupJob) Start(ctx context.Context) {
	log.Info().
		Int("retention_days", j.retentionDays).
		Dur("interval", j.interval).
		Msg("Starting WhatsApp audit cleanup job")

	// Run cleanup immediately on start
	j.runCleanup()

	// Start periodic cleanup
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.runCleanup()
		case <-ctx.Done():
			log.Info().Msg("Stopping WhatsApp audit cleanup job (context cancelled)")
			return
		case <-j.stopChan:
			log.Info().Msg("Stopping WhatsApp audit cleanup job")
			return
		}
	}
}

// Stop stops the cleanup job
func (j *WhatsAppCleanupJob) Stop() {
	close(j.stopChan)
}

// runCleanup executes the cleanup operation
func (j *WhatsAppCleanupJob) runCleanup() {
	log.Info().Msg("Running WhatsApp audit cleanup")

	deletedCount, err := j.auditService.CleanupOldMessages(j.retentionDays)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to cleanup old WhatsApp audit messages")
		return
	}

	if deletedCount > 0 {
		log.Info().
			Int64("deleted_count", deletedCount).
			Int("retention_days", j.retentionDays).
			Msg("Successfully cleaned up old WhatsApp audit messages")
	} else {
		log.Debug().
			Int("retention_days", j.retentionDays).
			Msg("No old WhatsApp audit messages to cleanup")
	}
}

// RunOnce runs the cleanup job once (useful for testing or manual execution)
func (j *WhatsAppCleanupJob) RunOnce() (int64, error) {
	log.Info().Msg("Running one-time WhatsApp audit cleanup")
	return j.auditService.CleanupOldMessages(j.retentionDays)
}
