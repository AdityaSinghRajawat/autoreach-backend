package authManagement

import (
	"strings"

	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=10,max=128"`
	FullName string `json:"fullName" validate:"required,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=128"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,max=512"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,max=512"`
}

type UserResponse struct {
	UserID      string   `json:"userId"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	CandidateID *string  `json:"candidateId"`
}

type SessionResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	TokenType    string       `json:"tokenType"`
	ExpiresIn    int          `json:"expiresIn"`
}

func (r *RegisterRequest) ToUserModel() *authModel.User {
	user := authModel.NewUser()
	user.Email = authModel.NormalizeEmail(r.Email)

	return user
}

func (r *RegisterRequest) ToCandidateModel(user *authModel.User) *candidateModel.Candidate {
	candidate := candidateModel.NewCandidate()
	candidate.UserID = user.ID
	candidate.Email = user.Email
	candidate.FullName = strings.TrimSpace(r.FullName)

	return candidate
}

func (r *LoginRequest) NormalizedEmail() string {
	return authModel.NormalizeEmail(r.Email)
}

func NewUserResponse(user *authModel.User, candidateID string) UserResponse {
	response := UserResponse{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
	}
	if candidateID != "" {
		response.CandidateID = &candidateID
	}

	return response
}
