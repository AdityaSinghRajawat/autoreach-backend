package userManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *UserService) ExportUserData(
	ctx context.Context,
	principal sharedTypes.Principal,
) (*userTypes.UserDataExport, *config.CustomError) {
	userID := principal.Subject
	export := &userTypes.UserDataExport{ExportedAt: utils.GetCurrTimeStamp()}

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{Name: "export user", Run: func(ctx context.Context) (err error) {
			export.User, err = s.userRepo.FindByID(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{Name: "export candidate", Run: func(ctx context.Context) (err error) {
			export.Candidate, err = s.candidateRepo.FindByUserID(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{Name: "export resumes", Run: func(ctx context.Context) (err error) {
			export.Resumes, err = s.userDataExportRepo.ListResumes(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{
			Name: "export job matches",
			Run: func(ctx context.Context) (err error) {
				export.JobMatches, err = s.userDataExportRepo.ListJobMatches(ctx, userID)

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "export applications",
			Run: func(ctx context.Context) (err error) {
				export.Applications, err = s.userDataExportRepo.ListApplications(ctx, userID)

				return err
			},
		},
		utils.ConcurrentTask{Name: "export mailboxes", Run: func(ctx context.Context) (err error) {
			export.ChannelAccounts, err = s.userDataExportRepo.ListChannelAccounts(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{Name: "export templates", Run: func(ctx context.Context) (err error) {
			export.OutreachTemplates, err = s.userDataExportRepo.ListOutreachTemplates(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{Name: "export campaigns", Run: func(ctx context.Context) (err error) {
			export.OutreachCampaigns, err = s.userDataExportRepo.ListOutreachCampaigns(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{Name: "export messages", Run: func(ctx context.Context) (err error) {
			export.OutreachMessages, err = s.userDataExportRepo.ListOutreachMessages(ctx, userID)

			return err
		}},
		utils.ConcurrentTask{
			Name: "export suppressions",
			Run: func(ctx context.Context) (err error) {
				export.Suppressions, err = s.userDataExportRepo.ListSuppressions(ctx, userID)

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "export contact signals",
			Run: func(ctx context.Context) (err error) {
				export.ContactSignals, err = s.userDataExportRepo.ListContactSignals(ctx, userID)

				return err
			},
		},
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("export user data: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return export, nil
}
