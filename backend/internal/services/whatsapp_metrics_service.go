package services

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// WhatsAppMetricsService handles Prometheus metrics for WhatsApp integration
type WhatsAppMetricsService struct {
	// Message metrics
	messagesTotal             *prometheus.CounterVec
	messagesInbound           prometheus.Counter
	messagesOutbound          prometheus.Counter
	messagesFailed            *prometheus.CounterVec
	messageProcessingDuration prometheus.Histogram

	// Command metrics
	commandsTotal   *prometheus.CounterVec
	commandDuration *prometheus.HistogramVec
	commandErrors   *prometheus.CounterVec

	// Error metrics
	errorsTotal      *prometheus.CounterVec
	errorsByCategory *prometheus.CounterVec

	// Webhook metrics
	webhookRequestsTotal *prometheus.CounterVec
	webhookDuration      prometheus.Histogram

	// Rate limiting metrics
	rateLimitHits    prometheus.Counter
	rateLimitBlocked prometheus.Counter

	// Context metrics
	activeContexts prometheus.Gauge
	contextExpired prometheus.Counter
}

// NewWhatsAppMetricsService creates a new metrics service with Prometheus collectors
func NewWhatsAppMetricsService() *WhatsAppMetricsService {
	return &WhatsAppMetricsService{
		// Message metrics
		messagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_messages_total",
				Help: "Total number of WhatsApp messages processed",
			},
			[]string{"direction", "type", "status"},
		),
		messagesInbound: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatsapp_messages_inbound_total",
				Help: "Total number of inbound WhatsApp messages",
			},
		),
		messagesOutbound: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatsapp_messages_outbound_total",
				Help: "Total number of outbound WhatsApp messages",
			},
		),
		messagesFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_messages_failed_total",
				Help: "Total number of failed WhatsApp messages",
			},
			[]string{"direction", "error_type"},
		),
		messageProcessingDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "whatsapp_message_processing_duration_seconds",
				Help:    "Duration of message processing in seconds",
				Buckets: prometheus.DefBuckets,
			},
		),

		// Command metrics
		commandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_commands_total",
				Help: "Total number of WhatsApp commands executed",
			},
			[]string{"command", "status"},
		),
		commandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "whatsapp_command_duration_seconds",
				Help:    "Duration of command execution in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"command"},
		),
		commandErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_command_errors_total",
				Help: "Total number of command execution errors",
			},
			[]string{"command", "error_type"},
		),

		// Error metrics
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_errors_total",
				Help: "Total number of errors in WhatsApp integration",
			},
			[]string{"component", "error_type"},
		),
		errorsByCategory: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_errors_by_category_total",
				Help: "Total number of errors by category",
			},
			[]string{"category"},
		),

		// Webhook metrics
		webhookRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "whatsapp_webhook_requests_total",
				Help: "Total number of webhook requests received",
			},
			[]string{"status"},
		),
		webhookDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "whatsapp_webhook_duration_seconds",
				Help:    "Duration of webhook processing in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
			},
		),

		// Rate limiting metrics
		rateLimitHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatsapp_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
		),
		rateLimitBlocked: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatsapp_rate_limit_blocked_total",
				Help: "Total number of requests blocked by rate limiting",
			},
		),

		// Context metrics
		activeContexts: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "whatsapp_active_contexts",
				Help: "Number of active conversation contexts",
			},
		),
		contextExpired: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "whatsapp_context_expired_total",
				Help: "Total number of expired conversation contexts",
			},
		),
	}
}

// RecordInboundMessage records an inbound message metric
func (m *WhatsAppMetricsService) RecordInboundMessage(messageType, status string) {
	m.messagesTotal.WithLabelValues("inbound", messageType, status).Inc()
	m.messagesInbound.Inc()
}

// RecordOutboundMessage records an outbound message metric
func (m *WhatsAppMetricsService) RecordOutboundMessage(messageType, status string) {
	m.messagesTotal.WithLabelValues("outbound", messageType, status).Inc()
	m.messagesOutbound.Inc()
}

// RecordMessageFailed records a failed message metric
func (m *WhatsAppMetricsService) RecordMessageFailed(direction, errorType string) {
	m.messagesFailed.WithLabelValues(direction, errorType).Inc()
}

// RecordMessageProcessingDuration records the duration of message processing
func (m *WhatsAppMetricsService) RecordMessageProcessingDuration(duration time.Duration) {
	m.messageProcessingDuration.Observe(duration.Seconds())
}

// RecordCommandExecution records a command execution metric
func (m *WhatsAppMetricsService) RecordCommandExecution(command, status string) {
	m.commandsTotal.WithLabelValues(command, status).Inc()
}

// RecordCommandDuration records the duration of command execution
func (m *WhatsAppMetricsService) RecordCommandDuration(command string, duration time.Duration) {
	m.commandDuration.WithLabelValues(command).Observe(duration.Seconds())
}

// RecordCommandError records a command error metric
func (m *WhatsAppMetricsService) RecordCommandError(command, errorType string) {
	m.commandErrors.WithLabelValues(command, errorType).Inc()
}

// RecordError records a general error metric
func (m *WhatsAppMetricsService) RecordError(component, errorType string) {
	m.errorsTotal.WithLabelValues(component, errorType).Inc()
}

// RecordErrorByCategory records an error by category
func (m *WhatsAppMetricsService) RecordErrorByCategory(category string) {
	m.errorsByCategory.WithLabelValues(category).Inc()
}

// RecordWebhookRequest records a webhook request metric
func (m *WhatsAppMetricsService) RecordWebhookRequest(status string) {
	m.webhookRequestsTotal.WithLabelValues(status).Inc()
}

// RecordWebhookDuration records the duration of webhook processing
func (m *WhatsAppMetricsService) RecordWebhookDuration(duration time.Duration) {
	m.webhookDuration.Observe(duration.Seconds())
}

// RecordRateLimitHit records a rate limit hit
func (m *WhatsAppMetricsService) RecordRateLimitHit() {
	m.rateLimitHits.Inc()
}

// RecordRateLimitBlocked records a blocked request due to rate limiting
func (m *WhatsAppMetricsService) RecordRateLimitBlocked() {
	m.rateLimitBlocked.Inc()
}

// SetActiveContexts sets the number of active conversation contexts
func (m *WhatsAppMetricsService) SetActiveContexts(count float64) {
	m.activeContexts.Set(count)
}

// RecordContextExpired records an expired context
func (m *WhatsAppMetricsService) RecordContextExpired() {
	m.contextExpired.Inc()
}

// MessageProcessingTimer returns a timer for measuring message processing duration
func (m *WhatsAppMetricsService) MessageProcessingTimer() *prometheus.Timer {
	return prometheus.NewTimer(m.messageProcessingDuration)
}

// WebhookTimer returns a timer for measuring webhook processing duration
func (m *WhatsAppMetricsService) WebhookTimer() *prometheus.Timer {
	return prometheus.NewTimer(m.webhookDuration)
}

// CommandTimer returns a timer for measuring command execution duration
func (m *WhatsAppMetricsService) CommandTimer(command string) *prometheus.Timer {
	return prometheus.NewTimer(m.commandDuration.WithLabelValues(command))
}
