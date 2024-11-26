package songinfo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"song-lib/internal/config"
	"song-lib/internal/domain"
	"time"

	"github.com/pkg/errors"
)

type SongInfoIntegration struct {
	songInfoURL string
}

func NewSongInfoIntegration(
	cfg config.SongInfoIntegrationAPIConfig,
) *SongInfoIntegration {
	return &SongInfoIntegration{
		songInfoURL: fmt.Sprintf(
			"%s://%s%s",
			cfg.Scheme, cfg.Domain,
			cfg.SongInfoPath),
	}
}

const (
	songInfoAPIPath          = "/info"
	songInfoReleseDateLayout = "02.01.2006"
)

var newError = func(err error) domain.SongInfoIntegrationError {
	return domain.SongInfoIntegrationError(err)
}

type songInfoResponseBody struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func (i *SongInfoIntegration) GetSongInfo(
	songName, musicGroupName string,
) (*domain.IntegrationSongInfo, error) {
	req, err := http.NewRequest(http.MethodGet, i.songInfoURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}
	q := req.URL.Query()
	q.Add("song", songName)
	q.Add("group", musicGroupName)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, newError(errors.Wrap(err, "make request"))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, newError(fmt.Errorf(
			"response status code %d, body: %s", resp.StatusCode,
			string(bodyBytes)))
	}

	var respBody songInfoResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, newError(errors.Wrap(err, "parse response body"))
	}
	songInfo, err := respBody.toDomainSongInfo()
	if err != nil {
		return nil, newError(errors.Wrap(err, "parse response body"))
	}

	return songInfo, nil
}

func (b *songInfoResponseBody) toDomainSongInfo() (*domain.IntegrationSongInfo, error) {
	releaseDate, err := time.Parse(songInfoReleseDateLayout, b.ReleaseDate)
	if err != nil {
		return nil, errors.Wrap(err, "parse release date")
	}

	return &domain.IntegrationSongInfo{
		ReleaseDate: releaseDate,
		Text:        b.Text,
		Link:        b.Link,
	}, nil
}
