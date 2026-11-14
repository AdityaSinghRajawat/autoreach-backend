package shared

import "github.com/AdityaSinghRajawat/autoreach-backend/config"

type PageParams struct {
	Page    int `form:"page"`
	PerPage int `form:"perPage"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

func (p PageParams) Normalize() PageParams {
	normalized := p
	if normalized.Page < 1 {
		normalized.Page = config.GetDefaultPage()
	}

	if normalized.PerPage < 1 {
		normalized.PerPage = config.GetDefaultPerPage()
	}

	if normalized.PerPage > config.GetMaxPerPage() {
		normalized.PerPage = config.GetMaxPerPage()
	}

	return normalized
}

func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func NewPaginationMeta(params PageParams, totalItems int) PaginationMeta {
	totalPages := 0
	if params.PerPage > 0 {
		totalPages = (totalItems + params.PerPage - 1) / params.PerPage
	}

	return PaginationMeta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
