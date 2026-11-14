package authManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) issueSession(
	ctx context.Context,
	user *authModel.User,
	candidateID, familyID string,
) (*authTypes.SessionResponse, error) {
	now := utils.GetCurrTimeStamp()
	principal := sharedTypes.Principal{UserID: user.ID, CandidateID: candidateID, Role: user.Role}

	accessToken, expiresAt, err := utils.IssueAccessToken(principal, now)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenHash, err := utils.NewOpaqueToken()
	if err != nil {
		return nil, err
	}

	record := authModel.NewRefreshToken(config.GetRefreshTokenTTL())
	record.UserID = user.ID
	record.TokenHash = refreshTokenHash

	if familyID != "" {
		record.FamilyID = familyID
	}

	if err := record.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.refreshTokenRepo.Upsert(ctx, record); err != nil {
		return nil, err
	}

	return &authTypes.SessionResponse{
		User:         authTypes.NewUserResponse(user, candidateID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    config.GetTokenTypeBearer(),
		ExpiresIn:    int(expiresAt.Sub(now).Seconds()),
	}, nil
}

func (s *AuthService) findActiveUserWithCandidate(
	ctx context.Context,
	userID string,
) (*authModel.User, string, error) {
	var (
		user        *authModel.User
		candidateID string
	)

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "find user",
			Run: func(ctx context.Context) error {
				found, err := s.userRepo.FindByID(ctx, userID)
				user = found

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "find candidate",
			Run: func(ctx context.Context) error {
				found, err := s.findCandidateID(ctx, userID)
				candidateID = found

				return err
			},
		},
	)
	if err != nil {
		return nil, "", err
	}

	if user == nil || !user.IsActive {
		return nil, "", config.ErrUserInactive
	}

	return user, candidateID, nil
}

func (s *AuthService) findCandidateID(ctx context.Context, userID string) (string, error) {
	candidate, err := s.candidateRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if candidate == nil {
		return "", nil
	}

	return candidate.ID, nil
}
