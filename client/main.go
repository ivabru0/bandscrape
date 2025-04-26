package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BandcampRequestJSON describes the JSON data sent to the Bandcamp API to
// retreive track info.
type BandcampRequestJSON struct {
	TralbumType string `json:"tralbum_type"`
	BandID      int    `json:"band_id"`
	TralbumID   int    `json:"tralbum_id"`
}

// BandcampResponseJSON describes the JSON data received from the Bandcamp API
// in response to BandcampRequestJSON.
type BandcampResponseJSON struct {
	Error         bool   `json:"error"`
	ErrorMessage  string `json:"error_message"`
	Title         string `json:"title"`
	AlbumTitle    string `json:"album_title"`
	TralbumArtist string `json:"tralbum_artist"`
	BandcampURL   string `json:"bandcamp_url"`
}

// Track describes a Bandcamp track.
type Track struct {
	TrackID    int    `json:"track_id"`
	TrackTitle string `json:"track_title"`
	AlbumTitle string `json:"album_title"`
	BandName   string `json:"band_name"`
	TrackURL   string `json:"track_url"`
}

func submitTracks(tracks []Track) error {
	requestJSON, err := json.Marshal(tracks)
	if err != nil {
		return fmt.Errorf("encode request JSON: %w", err)
	}

	// Evading AV false-positives (unknown URLs embedded in executable)
	response, err := http.Post(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						//"LttpCSS127D0D0D1C8585Ssubmit", "L", "h",
						"LttpsCSSbsDlibremcDnetSsubmit", "L", "h",
					), "C", ":",
				), "S", "/",
			), "D", ".",
		),
		"application/json",
		bytes.NewReader(requestJSON),
	)
	if err != nil {
		return fmt.Errorf("send POST request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	bodyFmt := strings.TrimSpace(string(body))

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("non-201 response code: %s", bodyFmt)
	}

	return nil
}

func getTrack(id int) (*Track, error) {
	for {
		requestJSON, err := json.Marshal(
			BandcampRequestJSON{
				TralbumType: "t",
				BandID:      1,
				TralbumID:   id,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("encode request JSON: %w", err)
		}

		response, err := http.Post(
			"https://bandcamp.com/api/mobile/26/tralbum_details",
			"application/json",
			bytes.NewBuffer(requestJSON),
		)
		if err != nil {
			return nil, fmt.Errorf("send POST request: %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusTooManyRequests {
			retryAfter, err := strconv.Atoi(response.Header.Get("retry-after"))
			if err != nil {
				retryAfter = 3
			}

			log.Println(id, "- WAIT -", retryAfter, "s")
			time.Sleep(time.Duration(retryAfter) * time.Second)
			continue
		}

		if response.StatusCode != http.StatusOK {
			return nil, errors.New("non-200 response status code")
		}

		var responseJSON BandcampResponseJSON
		if err := json.NewDecoder(response.Body).Decode(&responseJSON); err != nil {
			return nil, fmt.Errorf("decode response JSON: %w", err)
		}

		if responseJSON.Error {
			if strings.HasPrefix(responseJSON.ErrorMessage, "No such tralbum") {
				return nil, nil
			} else {
				return nil, errors.New("error true in response: " + responseJSON.ErrorMessage)
			}
		}

		return &Track{
			TrackID:    id,
			TrackTitle: responseJSON.Title,
			AlbumTitle: responseJSON.AlbumTitle,
			BandName:   responseJSON.TralbumArtist,
			TrackURL:   responseJSON.BandcampURL,
		}, nil
	}
}

func main() {
	for {
		var tracks []Track
		for range 100 {
			id := int(rand.Uint32())
			startTime := time.Now()
			track, err := getTrack(id)

			sleepTime := 1000 - time.Since(startTime).Milliseconds()
			if sleepTime > 0 {
				log.Println(id, "- WAIT -", sleepTime, "ms")
				time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			}

			if err != nil {
				log.Println(id, "- ERR -", err)
				continue
			}

			if track == nil {
				log.Println(id, "- NOK")
				continue
			}

			log.Println(id, "- OK")
			tracks = append(tracks, *track)
		}

		if err := submitTracks(tracks); err != nil {
			log.Println("Failed to submit tracks:", err)
		} else {
			log.Println("Submitted", len(tracks), "track(s)!")
		}
	}
}
