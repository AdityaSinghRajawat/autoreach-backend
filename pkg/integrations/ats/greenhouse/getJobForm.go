package greenhouse

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
)

func (i *GreenhouseIntegration) GetJobForm(
	ctx context.Context,
	boardToken, jobID string,
) (applicationTypes.ApplicationForm, error) {
	response := jobResponse{}

	err := i.httpClient.Get(
		ctx,
		jobPath(boardToken, jobID)+config.GetGreenhouseQuestionsQuery(),
		nil,
		&response,
	)
	if err != nil {
		return applicationTypes.ApplicationForm{}, fmt.Errorf("get greenhouse job form: %w", err)
	}

	form := applicationTypes.ApplicationForm{}

	for _, item := range append(response.Questions, response.LocationQuestions...) {
		formQuestion := applicationTypes.FormQuestion{Label: item.Label, Required: item.Required}

		for _, itemField := range item.Fields {
			formField := applicationTypes.FormField{Name: itemField.Name, Type: itemField.Type}
			for _, value := range itemField.Values {
				formField.Options = append(formField.Options, applicationTypes.FormOption{
					Label: value.Label,
					Value: rawValue(value.Value),
				})
			}

			formQuestion.Fields = append(formQuestion.Fields, formField)
		}

		form.Questions = append(form.Questions, formQuestion)
	}

	return form, nil
}
