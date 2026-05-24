package adminactivitylog

type Usecase interface {
	GetActivityLog(userID, page, limit int) (ActivityLogResponse, error)
}
