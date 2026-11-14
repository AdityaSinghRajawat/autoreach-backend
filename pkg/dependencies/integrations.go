package dependencies

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	openaiIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/embeddings/openai"
	localStorage "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/storage/local"
)

func (d *Dependencies) initIntegrations() error {
	if d.Embedder == nil {
		d.Embedder = openaiIntegration.NewOpenAIIntegration()
	}

	if d.FileStore == nil {
		fileStore, err := localStorage.NewLocalStorageIntegration(config.GetStorageLocalDir())
		if err != nil {
			return err
		}

		d.FileStore = fileStore
	}

	return nil
}
