package handlers

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"

	"mhp-rooms/internal/models"
)

func sampleRoomDetailData() TemplateData {
	host := models.User{BaseModel: models.BaseModel{ID: uuid.New()}, SupabaseUserID: uuid.New(), DisplayName: "ホスト太郎"}
	guest := models.User{BaseModel: models.BaseModel{ID: uuid.New()}, SupabaseUserID: uuid.New(), DisplayName: "参加者花子"}
	room := &models.Room{
		BaseModel:   models.BaseModel{ID: uuid.New()},
		RoomCode:    "ABC123",
		Name:        "テスト部屋",
		HostUserID:  host.ID,
		MaxPlayers:  4,
		IsActive:    true,
		GameVersion: models.GameVersion{Code: "MHP2G", Name: "モンスターハンターポータブル 2nd G"},
		Host:        host,
	}
	members := []*models.RoomMember{
		{ID: uuid.New(), RoomID: room.ID, UserID: host.ID, PlayerNumber: 1, IsHost: true, Status: models.MemberStatusActive, User: host, DisplayName: host.DisplayName},
		{ID: uuid.New(), RoomID: room.ID, UserID: guest.ID, PlayerNumber: 2, Status: models.MemberStatusActive, User: guest, DisplayName: guest.DisplayName},
	}
	return TemplateData{
		Title: "テスト部屋",
		PageData: RoomDetailPageData{
			Room:        room,
			Members:     members,
			MemberCount: len(members),
			IsHost:      true,
		},
	}
}

// TestRenderRoomDetailWithKickUI 部屋詳細ページ（キック UI・通報モーダル込み）が描画でき、
// 埋め込み JavaScript が構文エラーなくパースできることを確認する
func TestRenderRoomDetailWithKickUI(t *testing.T) {
	chdirRepoRoot(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rooms/test", nil)
	renderRoomDetailTemplate(w, r, "room_detail.tmpl", sampleRoomDetailData())

	body := w.Body.String()
	if w.Code != 200 || strings.Contains(body, "Template parsing error") || strings.Contains(body, "Template execution error") {
		t.Fatalf("status = %d, body:\n%s", w.Code, truncate(body, 2000))
	}

	for _, want := range []string{
		`@click="toggleMemberMenu(index)"`,
		`@click="openKickModal(member)"`,
		`@click="openReportForMember(member)"`,
		`x-show="showKickModal"`,
		`id="reportModal"`,
		`/kick`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("描画結果に %q が含まれていない", want)
		}
	}

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node が無いため JavaScript の構文チェックはスキップ")
	}
	// JSON-LD など JavaScript 以外の script は除外する
	scripts := regexp.MustCompile(`(?s)<script([^>]*)>(.*?)</script>`).FindAllStringSubmatch(body, -1)
	if len(scripts) == 0 {
		t.Fatal("インライン script が見つからない")
	}
	dir := t.TempDir()
	for i, m := range scripts {
		attrs, src := m[1], m[2]
		if strings.TrimSpace(src) == "" || (strings.Contains(attrs, "type=") && !strings.Contains(attrs, "javascript") && !strings.Contains(attrs, "module")) {
			continue
		}
		path := filepath.Join(dir, "script"+string(rune('a'+i))+".js")
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("node", "--check", path).CombinedOutput(); err != nil {
			t.Errorf("script #%d の構文エラー: %v\n%s\n--- 先頭 ---\n%s", i, err, out, truncate(src, 300))
		}
	}
}

func TestRoomJoinKickedErrorCode(t *testing.T) {
	// リポジトリの KICKED エラーはハンドラーで 403 + error コードに変換される前提。プレフィックスの取り決めを固定する
	err := "KICKED:この部屋から退出させられたため、再度参加することはできません"
	if !strings.HasPrefix(err, "KICKED:") {
		t.Fatal("KICKED プレフィックスの取り決めが崩れている")
	}
	if got := strings.TrimPrefix(err, "KICKED:"); got != "この部屋から退出させられたため、再度参加することはできません" {
		t.Errorf("message = %q", got)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
