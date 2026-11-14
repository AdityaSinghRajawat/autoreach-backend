package storage

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage/local"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage/s3"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func NewFileStore(ctx context.Context) (sharedTypes.FileStore, error) {
	driver := config.GetStorageDriver()
	utils.LogInfo(ctx, "file storage selected", "driver", driver)

	switch driver {
	case config.StorageDriverLocal:
		return local.NewLocalStorage(config.GetStorageLocalDir())
	case config.StorageDriverS3:
		return s3.NewS3Storage(ctx, config.GetS3Settings())
	default:
		return nil, fmt.Errorf("new file store: %w: %q", config.ErrUnknownStorageDriver, driver)
	}
}
