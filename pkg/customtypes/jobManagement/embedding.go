package jobManagement

import "strings"

type UpsertedJob struct {
	ID             string
	NeedsEmbedding bool
}

type JobEmbeddingSource struct {
	ID              string
	CompanyName     string
	JobTitle        string
	Location        *string
	Department      *string
	JobDescription  string
	DescriptionHash string
}

type JobEmbeddingUpdate struct {
	ID              string
	DescriptionHash string
	Embedding       []float32
}

type EmbedJobsResult struct {
	Requested int
	Embedded  int
}

func NewJobEmbeddingUpdate(source JobEmbeddingSource, embedding []float32) JobEmbeddingUpdate {
	return JobEmbeddingUpdate{
		ID:              source.ID,
		DescriptionHash: source.DescriptionHash,
		Embedding:       embedding,
	}
}

func (s JobEmbeddingSource) EmbeddingText(maxChars int) string {
	builder := strings.Builder{}
	builder.WriteString("Title: " + s.JobTitle + "\n")
	builder.WriteString("Company: " + s.CompanyName + "\n")

	if s.Location != nil && *s.Location != "" {
		builder.WriteString("Location: " + *s.Location + "\n")
	}

	if s.Department != nil && *s.Department != "" {
		builder.WriteString("Department: " + *s.Department + "\n")
	}

	builder.WriteString("\n" + s.JobDescription)

	text := []rune(builder.String())
	if maxChars > 0 && len(text) > maxChars {
		text = text[:maxChars]
	}

	return string(text)
}
