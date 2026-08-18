package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestParsePageParam(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "未指定は1", query: "", want: 1},
		{name: "正の整数はそのまま", query: "page=3", want: 3},
		{name: "0は1に丸める", query: "page=0", want: 1},
		{name: "負数は1に丸める", query: "page=-2", want: 1},
		{name: "数値以外は1", query: "page=abc", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/profile/rooms?"+tt.query, nil)
			if got := parsePageParam(r); got != tt.want {
				t.Errorf("parsePageParam() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name           string
		total          int64
		page           int
		wantTotalPages int
	}{
		{name: "0件でも総ページ数は1", total: 0, page: 1, wantTotalPages: 1},
		{name: "1ページに収まる", total: 20, page: 1, wantTotalPages: 1},
		{name: "端数は切り上げ", total: 21, page: 2, wantTotalPages: 2},
		{name: "複数ページ", total: 45, page: 3, wantTotalPages: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newPagination(tt.total, tt.page, tabPerPage, "/api/profile/rooms")
			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}
			if got.CurrentPage != tt.page {
				t.Errorf("CurrentPage = %d, want %d", got.CurrentPage, tt.page)
			}
			if got.Total != tt.total {
				t.Errorf("Total = %d, want %d", got.Total, tt.total)
			}
			if got.PerPage != tabPerPage {
				t.Errorf("PerPage = %d, want %d", got.PerPage, tabPerPage)
			}
			if got.BaseURL != "/api/profile/rooms" {
				t.Errorf("BaseURL = %q, want %q", got.BaseURL, "/api/profile/rooms")
			}
		})
	}
}
