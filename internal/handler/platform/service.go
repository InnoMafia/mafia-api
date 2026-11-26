package platform

import (
	platformv1 "github.com/InnoMafia/mafia-api/gen/go/mafia/platform/v1"
	"github.com/InnoMafia/mafia-api/internal/usecase"
)

type Service struct {
	platformv1.UnimplementedPlatformServiceServer
	getServiceStatus *usecase.GetServiceStatus
}

func New(getServiceStatus *usecase.GetServiceStatus) *Service {
	return &Service{getServiceStatus: getServiceStatus}
}
