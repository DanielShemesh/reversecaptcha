package artists

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ArtistsData struct {
	Artists []struct {
		ID string
	}
}

type ReleaseGroup struct {
	Title            string   `json:"title"`
	PrimaryTypeID    string   `json:"primary-type-id"`
	PrimaryType      string   `json:"primary-type"`
	SecondaryTypeIds []string `json:"secondary-type-ids"`
	FirstReleaseDate string   `json:"first-release-date"`
	ID               string   `json:"id"`
	SecondaryTypes   []string `json:"secondary-types"`
	Disambiguation   string   `json:"disambiguation"`
}

type ReleaseGroupsData struct {
	ReleaseGroups      []ReleaseGroup `json:"release-groups"`
	ReleaseGroupCount  int            `json:"release-group-count"`
	ReleaseGroupOffset int            `json:"release-group-offset"`
}

type ReleasesData struct {
	Releases []struct {
		ID string
	}
}

type Track struct {
	Title string `json:"title"`
}

type TracksData struct {
	Media []struct {
		Tracks []Track
	}
}

type AlbumCover struct {
	Images []struct {
		Approved   bool   `json:"approved"`
		Back       bool   `json:"back"`
		Comment    string `json:"comment"`
		Edit       int    `json:"edit"`
		Front      bool   `json:"front"`
		ID         int64  `json:"id"`
		Image      string `json:"image"`
		Thumbnails struct {
			Num250  string `json:"250"`
			Num500  string `json:"500"`
			Num1200 string `json:"1200"`
			Large   string `json:"large"`
			Small   string `json:"small"`
		} `json:"thumbnails"`
		Types []string `json:"types"`
	} `json:"images"`
	Release string `json:"release"`
}

func SearchArtist(artistName string) (string, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/artist?query=artist:%v&fmt=json", url.QueryEscape(artistName))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var artistsData ArtistsData
	json.Unmarshal(body, &artistsData)
	return artistsData.Artists[0].ID, nil
}

func GetAlbums(artistID string) ([]ReleaseGroup, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group?artist=%v&type=album&fmt=json", artistID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var releaseGroupsData ReleaseGroupsData
	json.Unmarshal(body, &releaseGroupsData)

	albums := releaseGroupsData.ReleaseGroups
	return albums, nil
}

func GetReleaseID(album ReleaseGroup) (string, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release?release-group=%v&fmt=json", album.ID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var releasesData ReleasesData
	if err := json.Unmarshal(body, &releasesData); err != nil {
		return "", err
	}
	if len(releasesData.Releases) == 0 {
		return "", fmt.Errorf("no releases found for album %v", album.ID)
	}
	return releasesData.Releases[0].ID, nil
}

func GetAlbumSongs(album ReleaseGroup) ([]Track, error) {
	releaseID, err := GetReleaseID(album)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release/%v?inc=recordings&fmt=json", releaseID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tracks TracksData
	if err := json.Unmarshal(body, &tracks); err != nil {
		return nil, err
	}
	if len(tracks.Media) == 0 {
		return nil, fmt.Errorf("no media found for release %v", releaseID)
	}
	return tracks.Media[0].Tracks, nil
}

func GetAlbumCover(album ReleaseGroup) (string, error) {
	albumID, err := GetReleaseID(album)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("http://coverartarchive.org/release/%v", albumID)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {

	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var albumCover AlbumCover
	json.Unmarshal(body, &albumCover)
	return albumCover.Images[0].Thumbnails.Small, nil

}
