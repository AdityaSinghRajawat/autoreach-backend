package unit

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/storage/local"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestPDFTextIsExtracted(t *testing.T) {
	document := helpers.BuildPDF(t, "Alex Candidate", "Senior Go Engineer (Kubernetes)")

	text, err := utils.ExtractPDFText(document, 10_000)

	require.NoError(t, err)
	assert.Contains(t, text, "Alex Candidate")
	assert.Contains(t, text, "Senior Go Engineer (Kubernetes)")
}

func TestDocxTextIsExtractedParagraphByParagraph(t *testing.T) {
	document := helpers.BuildDocx(t, "Alex Candidate", "Go & Kubernetes <platform>")

	text, err := utils.ExtractDocxText(document, 10_000)

	require.NoError(t, err)
	assert.Equal(t, "Alex Candidate\nGo & Kubernetes <platform>", text)
}

func TestDocumentExtractionRejectsCorruptAndOversizedInput(t *testing.T) {
	_, err := utils.ExtractPDFText([]byte("%PDF-1.4 this is not really a pdf"), 10_000)
	require.ErrorIs(t, err, config.ErrUnreadableDocument)

	_, err = utils.ExtractDocxText([]byte("PK\x03\x04 word/document.xml truncated"), 10_000)
	require.ErrorIs(t, err, config.ErrUnreadableDocument)

	huge := helpers.BuildDocx(t, strings.Repeat("resume ", 50_000))
	_, err = utils.ExtractDocxText(huge, 1_000)
	require.ErrorIs(t, err, config.ErrDocumentTooLarge)

	truncated, err := utils.ExtractPDFText(helpers.BuildPDF(t, strings.Repeat("a", 500)), 100)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(truncated), 100)
}

func TestResumeContentTypeIsDetectedFromBytesNotFromTheFileName(t *testing.T) {
	assert.Equal(
		t,
		modelHelpers.ContentTypePDF,
		modelHelpers.DetectResumeContentType(helpers.BuildPDF(t, "x")),
	)
	assert.Equal(
		t,
		modelHelpers.ContentTypeDOCX,
		modelHelpers.DetectResumeContentType(helpers.BuildDocx(t, "x")),
	)
	assert.Empty(t, modelHelpers.DetectResumeContentType([]byte("plain text named resume.pdf")))
	assert.Empty(
		t,
		modelHelpers.DetectResumeContentType([]byte("PK\x03\x04 a zip that is not a docx")),
	)
	assert.Empty(t, modelHelpers.DetectResumeContentType(nil))
}

func TestFileNamesAreReducedToASafeBaseName(t *testing.T) {
	cases := map[string]string{
		"resume.pdf":                   "resume.pdf",
		"../../etc/passwd":             "passwd",
		`C:\Users\alex\My Resume.docx`: "My Resume.docx",
		"  spaced.pdf  ":               "spaced.pdf",
		"line\nbreak\x00.pdf":          "linebreak.pdf",
		"":                             "resume",
		"/":                            "resume",
	}
	for input, expected := range cases {
		assert.Equal(t, expected, modelHelpers.SanitizeFileName(input), "input %q", input)
	}

	assert.LessOrEqual(t, len(modelHelpers.SanitizeFileName(strings.Repeat("a", 600)+".pdf")), 255)
}

func TestUploadResumeRequestConvertsToAResumeEntity(t *testing.T) {
	setup.Config(t)

	candidateID := uuid.NewString()
	content := helpers.BuildPDF(t, "Alex Candidate")
	request := candidateTypes.UploadResumeRequest{}
	request.SetFile("../secret/Alex Resume.pdf", content)

	resume := request.ToResumeModel(candidateID)

	assert.Equal(t, candidateID, resume.CandidateID)
	assert.Equal(t, "Alex Resume.pdf", resume.FileName)
	assert.Equal(t, modelHelpers.ContentTypePDF, resume.ContentType)
	assert.Equal(t, int64(len(content)), resume.SizeBytes)
	assert.Equal(t, utils.Fingerprint(string(content)), resume.ContentSHA256)
	assert.Equal(t, "resumes/"+candidateID+"/"+resume.ID+".pdf", resume.StorageKey)
	assert.Equal(t, modelHelpers.ResumeStatusUploaded, resume.Status)
	assert.False(t, resume.IsActive)
	require.NoError(t, resume.Validate())

	unsupported := candidateTypes.UploadResumeRequest{}
	unsupported.SetFile("notes.txt", []byte("just text"))
	require.Error(t, unsupported.ToResumeModel(candidateID).Validate())
}

func TestLocalStorageRoundTripsAndRefusesKeysOutsideItsRoot(t *testing.T) {
	store, err := local.NewLocalStorage(t.TempDir())
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, store.Save(ctx, "resumes/c1/r1.pdf", []byte("content")))

	content, err := store.Read(ctx, "resumes/c1/r1.pdf")
	require.NoError(t, err)
	assert.Equal(t, "content", string(content))

	require.NoError(t, store.Delete(ctx, "resumes/c1/r1.pdf"))
	require.NoError(
		t,
		store.Delete(ctx, "resumes/c1/r1.pdf"),
		"deleting a missing file is not an error",
	)

	_, err = store.Read(ctx, "resumes/c1/r1.pdf")
	require.ErrorIs(t, err, config.ErrStoredFileNotFound)

	for _, key := range []string{"", "../outside.pdf", "resumes/../../outside.pdf", "/etc/passwd"} {
		require.ErrorIs(
			t,
			store.Save(ctx, key, []byte("x")),
			config.ErrInvalidStorageKey,
			"key %q",
			key,
		)
	}
}

func TestMatchScoreCalibrationAndStorageThreshold(t *testing.T) {
	setup.Config(t)

	assert.InDelta(t, 0.9, config.CalibrateMatchScore(0.9), 0.0001, "identity mapping by default")
	assert.InDelta(t, 0, config.CalibrateMatchScore(-0.2), 0.0001)
	assert.InDelta(t, 1, config.CalibrateMatchScore(1.3), 0.0001)

	match := matchingTypes.JobSimilarity{JobID: uuid.NewString(), Similarity: 1.0000001}.
		ToJobMatchModel(uuid.NewString(), uuid.NewString())
	assert.InDelta(t, 1, match.RawSimilarity, 0.0001)
	require.NoError(t, match.Validate(), "floating point overshoot must not fail validation")

	assert.InDelta(
		t,
		min(0.8, config.GetMatchMinStoredScore()),
		matchingTypes.StorageThreshold(0.8),
		0.0001,
	)
	assert.InDelta(t, 0, matchingTypes.StorageThreshold(0), 0.0001)
}

func TestPercentileCosineAndCalibrationSuggestion(t *testing.T) {
	values := []float64{0.5, 0.1, 0.3, 0.2, 0.4}

	assert.InDelta(t, 0.1, utils.Percentile(values, 0), 0.0001)
	assert.InDelta(t, 0.3, utils.Percentile(values, 0.5), 0.0001)
	assert.InDelta(t, 0.5, utils.Percentile(values, 1), 0.0001)
	assert.InDelta(t, 0.15, utils.Percentile(values, 0.125), 0.0001)
	assert.Zero(t, utils.Percentile(nil, 0.5))

	assert.InDelta(t, 1, utils.CosineSimilarity([]float32{1, 2}, []float32{2, 4}), 0.0001)
	assert.InDelta(t, 0, utils.CosineSimilarity([]float32{1, 0}, []float32{0, 1}), 0.0001)
	assert.Zero(t, utils.CosineSimilarity([]float32{1}, []float32{1, 2}))

	good := []float64{0.50, 0.55, 0.60, 0.65, 0.70}
	bad := []float64{0.20, 0.25, 0.30, 0.35, 0.40}
	low, high := utils.SuggestCalibration(good, bad)

	assert.InDelta(t, 0.30, low, 0.0001, "the median bad pair maps to zero")
	assert.InDelta(
		t,
		0.80,
		(0.55-low)/(high-low),
		0.0001,
		"the lower-quartile good pair maps to 0.80",
	)

	low, high = utils.SuggestCalibration(bad, good)
	assert.InDelta(t, 0, low, 0.0001, "labels that do not separate leave the score unchanged")
	assert.InDelta(t, 1, high, 0.0001)
}

func TestStorageKeysMustBeRelativePathsWithoutTraversal(t *testing.T) {
	setup.Config(t)

	for _, valid := range []string{"resumes/c1/r1.pdf", "a", "a/b/c.docx", "resumes/c-1/r_1.pdf"} {
		assert.NoError(t, utils.ValidateStorageKey(valid), "key %q", valid)
	}

	for _, invalid := range []string{
		"", "/etc/passwd", "../outside.pdf", "resumes/../../outside.pdf", "resumes/./r1.pdf",
		"resumes//r1.pdf", "resumes/r1.pdf/", `resumes\r1.pdf`, "C:/resumes/r1.pdf", "..",
	} {
		assert.ErrorIs(
			t,
			utils.ValidateStorageKey(invalid),
			config.ErrInvalidStorageKey,
			"key %q",
			invalid,
		)
	}
}

func TestLocalStorageIsTheDefaultDriver(t *testing.T) {
	setup.Config(t)

	assert.Equal(t, config.StorageDriverLocal, config.GetStorageDriver())
	assert.Equal(
		t,
		"us-east-1",
		config.GetS3Settings().Region,
		"the SDK needs a region even for S3-compatible services",
	)
}

func TestPDFThatInflatesPastTheLimitIsRefusedBeforeParsing(t *testing.T) {
	bomb := helpers.BuildPDFBomb(t, int(config.GetPDFMaxDecodedBytes())+1)
	require.Less(t, len(bomb), 1<<20, "the file itself is small")

	_, err := utils.ExtractPDFText(bomb, 10_000)
	require.ErrorIs(t, err, config.ErrDocumentTooLarge)
}
