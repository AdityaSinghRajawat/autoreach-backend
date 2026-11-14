package helpers

type UserRole string

const (
	UserRoleCandidate UserRole = "CANDIDATE"
	UserRoleAdmin     UserRole = "ADMIN"
)

func (r UserRole) IsValid() bool {
	return r == UserRoleCandidate || r == UserRoleAdmin
}

func (r UserRole) IsAdmin() bool {
	return r == UserRoleAdmin
}
