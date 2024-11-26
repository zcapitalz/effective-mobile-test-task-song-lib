package songinfo

import (
	"song-lib/internal/domain"

	"github.com/pkg/errors"
	"golang.org/x/exp/rand"
)

type SongInfoIntegrationMock struct{}

var mockSongInfoResponseBodies = []*songInfoResponseBody{
	{
		ReleaseDate: "15.06.2023",
		Text:        "Lost in the Echo, a distant sound.\nEchoes linger, memories rebound.\nThrough valleys deep, they drift away.\nThe past returns at the break of day.",
		Link:        "https://example.com/lost-echo",
	},
	{
		ReleaseDate: "04.11.2021",
		Text:        "Beyond the stars, where dreams take flight.\nA world unknown, bathed in light.\nInfinite wonders, endless streams.\nThe cosmos whispers its hidden dreams.",
		Link:        "https://example.com/beyond-stars",
	},
	{
		ReleaseDate: "23.08.2018",
		Text:        "Winds of change, through skies they soar.\nBreaking barriers, forevermore.\nWhispering softly, they carry the tide.\nA song of freedom, far and wide.",
		Link:        "https://example.com/winds-change",
	},
	{
		ReleaseDate: "10.02.2015",
		Text:        "Shadows of time, a fleeting grace.\nMoments vanish without a trace.\nYet in their wake, a spark remains.\nIlluminating life's winding lanes.",
		Link:        "https://example.com/shadows-time",
	},
	{
		ReleaseDate: "01.09.2012",
		Text:        "Whispers in the dark, secrets untold.\nIn the midnight hour, mysteries unfold.\nA tale of love, a fleeting spark.\nLost forever in the boundless dark.",
		Link:        "https://example.com/whispers-dark",
	},
}

func (i *SongInfoIntegrationMock) GetSongInfo(songName, musicGroupName string) (*domain.IntegrationSongInfo, error) {
	respBody := mockSongInfoResponseBodies[rand.Intn(len(mockSongInfoResponseBodies))]
	songInfo, err := respBody.toDomainSongInfo()
	if err != nil {
		return nil, errors.Wrap(err, "parse response body")
	}

	return songInfo, nil
}
