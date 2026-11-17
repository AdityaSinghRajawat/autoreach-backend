package authManagement

import (
	"strings"

	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
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

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,max=256"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"    validate:"required,max=256"`
	Password string `json:"password" validate:"required,min=10,max=128"`
}

type PasswordResetPendingResponse struct {
	ResetEmailSent bool `json:"resetEmailSent"`
}

type RegistrationPendingResponse struct {
	VerificationRequired bool `json:"verificationRequired"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,max=512"`
}

type SessionResponse struct {
	User         userTypes.UserResponse `json:"user"`
	AccessToken  string                 `json:"accessToken"`
	RefreshToken string                 `json:"refreshToken"`
	TokenType    string                 `json:"tokenType"`
	ExpiresIn    int                    `json:"expiresIn"`
}

func (r *RegisterRequest) ToUserModel() *userModel.User {
	user := userModel.NewUser()
	user.Email = modelHelpers.NormalizeEmail(r.Email)

	return user
}

func (r *RegisterRequest) ToCandidateModel(user *userModel.User) *candidateModel.Candidate {
	candidate := candidateModel.NewCandidate()
	candidate.UserID = user.ID
	candidate.Email = user.Email
	candidate.FullName = strings.TrimSpace(r.FullName)

	return candidate
}

func (r *ForgotPasswordRequest) NormalizedEmail() string {
	return modelHelpers.NormalizeEmail(r.Email)
}

func (r *LoginRequest) NormalizedEmail() string {
	return modelHelpers.NormalizeEmail(r.Email)
}
