//go:build integration

package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage/local"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage/s3"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestS3StorageRoundTripsAndTreatsMissingObjectsLikeLocalDisk(t *testing.T) {
	store, err := s3.NewS3Storage(context.Background(), setup.S3Settings(t))
	require.NoError(t, err)

	ctx := context.Background()
	key := "resumes/storage-test/" + t.Name() + ".pdf"
	content := helpers.BuildPDF(t, "Stored in S3")

	require.NoError(t, store.Save(ctx, key, content))

	read, err := store.Read(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, content, read)

	require.NoError(
		t,
		store.Save(ctx, key, []byte("%PDF-replaced")),
		"saving to the same key overwrites",
	)

	replaced, err := store.Read(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "%PDF-replaced", string(replaced))

	require.NoError(t, store.Delete(ctx, key))
	require.NoError(t, store.Delete(ctx, key), "deleting a missing object is not an error")

	_, err = store.Read(ctx, key)
	require.ErrorIs(t, err, config.ErrStoredFileNotFound)

	for _, invalid := range []string{"", "../outside.pdf", "/absolute.pdf"} {
		require.ErrorIs(
			t,
			store.Save(ctx, invalid, content),
			config.ErrInvalidStorageKey,
			"key %q",
			invalid,
		)
	}
}

func TestS3StorageRefusesToStartWhenTheBucketIsMissing(t *testing.T) {
	settings := setup.S3Settings(t)
	settings.Bucket = "autoreach-bucket-that-does-not-exist"

	_, err := s3.NewS3Storage(context.Background(), settings)

	require.ErrorIs(t, err, config.ErrStorageUnavailable)
}

func TestResumeIsStoredInS3WhenS3IsTheConfiguredStore(t *testing.T) {
	settings := setup.S3Settings(t)
	fixture := newJobFixture(t)

	store, err := s3.NewS3Storage(context.Background(), settings)
	require.NoError(t, err)

	fixture.deps.FileStore = store
	require.NoError(t, fixture.deps.Wire())

	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		localStore, err := local.NewLocalStorage(config.GetStorageLocalDir())
		require.NoError(t, err)

		fixture.deps.FileStore = localStore
		require.NoError(t, fixture.deps.Wire())
	})

	resume := fixture.processedResume(t, "Alex Candidate, Go and Kubernetes")
	assert.Equal(
		t,
		modelHelpers.ResumeStatusEmbedded,
		resume.Status,
		"the worker read the file back from S3",
	)

	stored, err := fixture.deps.Repositories.Resume.FindByID(context.Background(), resume.ID)
	require.NoError(t, err)

	content, err := store.Read(context.Background(), stored.StorageKey)
	require.NoError(t, err)
	assert.Equal(t, stored.SizeBytes, int64(len(content)))

	deleted := helpers.Do(
		t, fixture.router, http.MethodDelete,
		resumesPath(fixture.candidateID)+"/"+resume.ID, nil, bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusNoContent, deleted.Recorder.Code)

	require.Eventually(t, func() bool {
		_, err := store.Read(context.Background(), stored.StorageKey)

		return err != nil
	}, 5_000_000_000, 100_000_000, "deleting the resume removes the object")
}
