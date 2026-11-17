package userManagement

import userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"

type UserResponse struct {
	UserID      string   `json:"userId"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	CandidateID *string  `json:"candidateId"`
}

func NewUserResponse(user *userModel.User, candidateID string) UserResponse {
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
