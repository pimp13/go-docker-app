package service

// HealthStatus is the domain response for readiness/liveness checks.
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// HealthService defines health-check business logic.
type HealthService interface {
	Check() HealthStatus
}

type healthService struct {
	version string
}

// NewHealthService constructs a HealthService.
func NewHealthService(version string) HealthService {
	return &healthService{version: version}
}

func (s *healthService) Check() HealthStatus {
	return HealthStatus{
		Status:  "ok",
		Version: s.version,
	}
}
