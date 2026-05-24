package adminactivitylog

// ActivityLogEntry will be populated when module 06 (activity-log) is implemented.
type ActivityLogEntry struct{}

type ActivityLogResponse struct {
	Data       []ActivityLogEntry `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}
