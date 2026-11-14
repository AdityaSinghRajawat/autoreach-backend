package config

type matchConstants struct {
	resumeIDParam   string
	resumeFormField string
	storagePrefix   string
}

var matchConstI = matchConstants{
	resumeIDParam:   "resumeId",
	resumeFormField: "file",
	storagePrefix:   "resumes",
}

var skillDictionaryI = []string{
	"go", "golang", "python", "java", "javascript", "typescript", "rust", "c++", "c#", "ruby",
	"php", "kotlin", "swift", "scala", "sql", "postgresql", "mysql", "mongodb", "redis",
	"elasticsearch", "kafka", "rabbitmq", "grpc", "graphql", "rest", "docker", "kubernetes",
	"terraform", "ansible", "aws", "gcp", "azure", "linux", "git", "ci/cd", "jenkins",
	"github actions", "react", "next.js", "vue", "angular", "node.js", "django", "flask",
	"fastapi", "spring", "gin", "microservices", "distributed systems", "system design",
	"machine learning", "deep learning", "pytorch", "tensorflow", "nlp", "data engineering",
	"spark", "airflow", "pandas", "tableau", "figma", "agile", "scrum", "product management",
	"prometheus", "grafana", "observability", "security", "oauth", "html", "css", "tailwind",
}

func GetResumeIDParam() string       { return matchConstI.resumeIDParam }
func GetResumeFormField() string     { return matchConstI.resumeFormField }
func GetResumeStoragePrefix() string { return matchConstI.storagePrefix }

func GetSkillDictionary() []string {
	return append([]string(nil), skillDictionaryI...)
}

func CalibrateMatchScore(similarity float64) float64 {
	low, high := envConfigI.match.calibrationLow, envConfigI.match.calibrationHigh
	if high <= low {
		return clampUnit(similarity)
	}

	return clampUnit((similarity - low) / (high - low))
}

func clampUnit(value float64) float64 {
	return max(0, min(1, value))
}
