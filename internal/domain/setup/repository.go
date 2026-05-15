package setup

import "context"

type SetupRepository interface {
	SaveCompleteSetup(ctx context.Context, initialSetup *UserInitialSetup) error
}
