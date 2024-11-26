package songinfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"song-lib/internal/config"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetSongInfoSuccess(t *testing.T) {
	songInfoAPIPath := "/info"
	songName := "XLR8"
	musicGroupName := "REAPER"
	songText := "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"
	songLink := "https://www.youtube.com/watch?v=Xsp3_a-PMTw"
	songReleaseDateStr := "16.07.2006"
	songReleaseDate, err := time.Parse(songInfoReleseDateLayout, songReleaseDateStr)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != songInfoAPIPath {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		query := req.URL.Query()
		if query.Get("song") != songName {
			t.Logf(`query param "song" is invalid`)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if query.Get("group") != musicGroupName {
			t.Logf(`query param "group" is invalid`)
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		respBody, err := json.Marshal(songInfoResponseBody{
			ReleaseDate: songReleaseDateStr,
			Text:        songText,
			Link:        songLink,
		})
		if err != nil {
			t.Log("could not marshal response body:", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Write(respBody)
	}))
	defer server.Close()

	urlSplit := strings.Split(server.URL, "://")
	songInfoIntegration := NewSongInfoIntegration(config.SongInfoIntegrationAPIConfig{
		Scheme:       urlSplit[0],
		Domain:       urlSplit[1],
		SongInfoPath: songInfoAPIPath,
	})

	songInfo, err := songInfoIntegration.GetSongInfo(songName, musicGroupName)
	require.NoError(t, err)
	require.NotNil(t, songInfo)
	require.Equal(t, songReleaseDate, songInfo.ReleaseDate)
	require.Equal(t, songText, songInfo.Text)
	require.Equal(t, songLink, songInfo.Link)

}
