package dto

type WebResponse[T any] struct {
	Message string         `json:"message,omitempty"`
	Data    T              `json:"data,omitempty"`
	Paging  *PageMetadata  `json:"paging,omitempty"`
	Errors  *ErrorResponse `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

type PageResponse[T any] struct {
	Data         []T          `json:"data,omitempty"`
	PageMetadata PageMetadata `json:"paging,omitempty"`
}

type PageMetadata struct {
	Page        int   `json:"page"`
	Size        int   `json:"size"`
	TotalItem   int64 `json:"total_item"`
	TotalPage   int64 `json:"total_page"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}
