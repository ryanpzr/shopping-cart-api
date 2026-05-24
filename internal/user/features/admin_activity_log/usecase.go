package adminactivitylog

// usecase is a stub — returns empty paginated response.
// TODO (module 06): replace with real implementation backed by activitylog.Repository.
// See sdd/specs/modules/06-activity-log.md for the full checklist.
type usecase struct{}

func NewUsecase() Usecase {
	return &usecase{}
}

func (u *usecase) GetActivityLog(userID, page, limit int) (ActivityLogResponse, error) {
	return ActivityLogResponse{
		Data:       []ActivityLogEntry{},
		Total:      0,
		Page:       page,
		Limit:      limit,
		TotalPages: 0,
	}, nil
}
