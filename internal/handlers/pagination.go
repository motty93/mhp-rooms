package handlers

import (
	"math"
	"net/http"
	"strconv"
)

// tabPerPage プロフィールのタブ（作成した部屋・アクティビティ）で1ページに表示する件数
const tabPerPage = 20

// Pagination ページ送りUIの描画に必要な情報
type Pagination struct {
	CurrentPage int    `json:"currentPage"`
	TotalPages  int    `json:"totalPages"`
	Total       int64  `json:"total"`
	PerPage     int    `json:"perPage"`
	BaseURL     string `json:"baseUrl"`
}

// parsePageParam クエリパラメータ page を1以上の整数として返す（未指定・不正値は1）
func parsePageParam(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// newPagination 総件数からページ情報を組み立てる（総ページ数は最低1）
func newPagination(total int64, page, perPage int, baseURL string) Pagination {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	return Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		Total:       total,
		PerPage:     perPage,
		BaseURL:     baseURL,
	}
}
