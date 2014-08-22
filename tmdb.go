// Copyright 2014, Amahi.  All rights reserved.
// Use of this source code is governed by the
// license that can be found in the LICENSE file.

// Golang library for requesting metadata from themoviedb.org
// It's used by
//
// 1) Initializing the library via Init(), with the caller's API
// key from http://www.themoviedb.org like
//
// 2) Calling MovieData() to get the actual data, like
//
// For example
//
//	package main
//
//	import "fmt"
//	import "github.com/amahi/go-themoviedb"
//
//	func main() {
//		tmdb := tmdb.Init("your-api-key")
//		metadata, err := tmdb.MovieData("Pulp Fiction")
//		if err != nil {
//			fmt.Printf("Error: %s\n", err)
//		} else {
//			fmt.Printf("TMDb Metadata: %s\n", metadata)
//		}
//	}
//
// the metadata is returned in XML format according to TMDb guidelines.
//
package tmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const base_url string = "http://api.themoviedb.org/3"

type TMDb struct {
	api_key string
	config  *tmdbConfig
}

func Init(api_key string) *TMDb {
	return &TMDb{api_key: api_key}
}

type filtered_output struct {
	Title        string `json:"title"`
	Artwork      string `json:"artwork"`
	Release_date string `json:"year"`
}

// response of search/multi
type tmdbResponse struct {
	Page          int
	Results       []tmdbResult
	Total_pages   int
	Total_results int
}

// results format from Tmdb
type tmdbResult struct {
	Adult          bool
	Name           string
	Backdrop_path  string
	Id             int
	Original_name  string
	Original_title string
	First_air_date string
	Release_date   string
	Poster_path    string
	Title          string
	Media_type     string
	Profile_path   string
}

// response of config
type tmdbConfig struct {
	Images imageConfig
}

// Image configurtion
type imageConfig struct {
	Base_url        string
	Secure_base_url string

	//possible sizes for images
	Backdrop_sizes []string
	Logo_sizes     []string
	Poster_sizes   []string
	Profile_sizes  []string
	Still_sizes    []string
}

// Movie metadata structure
type movieMetadata struct {
	Id            int
	Media_type    string
	Backdrop_path string
	Poster_path   string
	Credits       tmdbCredits
	Config        *tmdbConfig
	Imdb_id       string
	Overview      string
	Title         string
	Release_date  string
}

type tmdbCredits struct {
	Id   int
	Cast []tmdbCast
	Crew []tmdbCrew
}

type tmdbCast struct {
	Character    string
	Name         string
	Profile_path string
}

type tmdbCrew struct {
	Department   string
	Name         string
	Job          string
	Profile_path string
}

// The main call for getting movie data media_name is the (plain) name of
// the movie information to be retrieved without year or other information
func (tmdb *TMDb) MovieData(media_name string) (string, error) {
	var met string
	results, err := tmdb.searchMovie(media_name)
	if err != nil {
		return met, err
	}
	if results.Total_results == 0 {
		return met, errors.New("No results found at TMDb")
	}
	if results.Results[0].Media_type == "person" {
		return met, errors.New("Metadata for persons not supported")
	}
	if results.Results[0].Media_type == "tv" {
		return met, errors.New("Metadata for tv not supported inside a call for movie data")
	}

	// otherwise
	movie_details, err := tmdb.getMovieDetails(strconv.Itoa(results.Results[0].Id))
	if err != nil {
		return met, err
	}
	movie_details.Credits, err = tmdb.getMovieCredits(strconv.Itoa(results.Results[0].Id))
	if err != nil {
		return met, err
	}
	movie_details.Config, err = tmdb.getConfig()
	if err != nil {
		return met, err
	}
	movie_details.Id = results.Results[0].Id
	movie_details.Media_type = "movie"

	metadata, err := json.Marshal(movie_details)
	if err != nil {
		return met, err
	}
	met = string(metadata)
	return met, nil
}

// Search on TMDb for TV, persons and Movies with a given name
func (tmdb *TMDb) searchTmdbMulti(media_name string) (tmdbResponse, error) {
	res, err := http.Get(base_url + "/search/multi?api_key=" + tmdb.api_key + "&query=" + url.QueryEscape(media_name))
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

// Search on TMDb for Movies with a given name
func (tmdb *TMDb) searchMovie(media_name string) (tmdbResponse, error) {
	res, err := http.Get(base_url + "/search/movie?api_key=" + tmdb.api_key + "&query=" + url.QueryEscape(media_name))
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

// Search on TMDb for Tv Shows with a given name
func (tmdb *TMDb) searchTmdbTv(media_name string) (tmdbResponse, error) {
	res, err := http.Get(base_url + "/search/tv?api_key=" + tmdb.api_key + "&query=" + url.QueryEscape(media_name))
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

// Get configurations from TMDb
func (tmdb *TMDb) getConfig() (*tmdbConfig, error) {
	if tmdb.config == nil || tmdb.config.Images.Base_url == "" {
		res, err := http.Get(base_url + "/configuration?api_key=" + tmdb.api_key)
		var conf = &tmdbConfig{}
		if err != nil {
			return conf, err
		}
		if res.StatusCode != 200 {
			return conf, error_status(res.StatusCode)
		}
		body, err := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(body, &conf)
		if err != nil {
			return &tmdbConfig{}, err
		}
		tmdb.config = conf
		return tmdb.config, nil
	} else {
		return tmdb.config, nil
	}
}

// Get basic information for movie
func (tmdb *TMDb) getMovieDetails(MediaId string) (movieMetadata, error) {
	res, err := http.Get(base_url + "/movie/" + MediaId + "?api_key=" + tmdb.api_key)
	var met movieMetadata
	if err != nil {
		return met, err
	}
	if res.StatusCode != 200 {
		return met, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &met)
	if err != nil {
		return movieMetadata{}, err
	}
	return met, nil
}

// Get credits for movie
func (tmdb *TMDb) getMovieCredits(MediaId string) (tmdbCredits, error) {
	res, err := http.Get(base_url + "/movie/" + MediaId + "/credits?api_key=" + tmdb.api_key)
	var cred tmdbCredits
	if err != nil {
		return cred, err
	}
	if res.StatusCode != 200 {
		return cred, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &cred)
	if err != nil {
		return tmdbCredits{}, err
	}
	return cred, nil
}

// Get basic information for Tv
func (tmdb *TMDb) getTmdbTvDetails(MediaId string) (movieMetadata, error) {
	res, err := http.Get(base_url + "/tv/" + MediaId + "?api_key=" + tmdb.api_key)
	var met movieMetadata
	if err != nil {
		return met, err
	}
	if res.StatusCode != 200 {
		return met, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &met)
	if err != nil {
		return movieMetadata{}, err
	}
	return met, nil
}

// Get credits for Tv
func (tmdb *TMDb) getTmdbTvCredits(MediaId string) (tmdbCredits, error) {
	res, err := http.Get(base_url + "/tv/" + MediaId + "/credits?api_key=" + tmdb.api_key)
	var cred tmdbCredits
	if err != nil {
		return cred, err
	}
	if res.StatusCode != 200 {
		return cred, error_status(res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &cred)
	if err != nil {
		return tmdbCredits{}, err
	}
	return cred, nil
}

// Transform the simplified movie metadata in JSON format
// This output is rather arbitrary to our (Amahi's) needs and could be customized a little
func (tmdb *TMDb) ToJSON(data string) (string, error) {
	var f filtered_output
	var det movieMetadata
	err := json.Unmarshal([]byte(data), &det)
	if err != nil {
		return "", err
	}

	f.Title = det.Title
	f.Release_date = det.Release_date
	if len(det.Release_date) > 4 {
		f.Release_date = det.Release_date[0:4]
	}
	// default width of the poster
	size := det.poster_size("w154")
	f.Artwork = det.Config.Images.Base_url + size + det.Poster_path

	metadata, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(metadata), nil
}

// return the requested size, the original if there are none
// and the first one if the requested size does not exist
func (md *movieMetadata) poster_size(size string) string {
	if len(md.Config.Images.Poster_sizes) == 0 {
		return "original"
	}
	for i := range md.Config.Images.Poster_sizes {
		if md.Config.Images.Poster_sizes[i] == size {
			return size
		}
	}
	return md.Config.Images.Poster_sizes[0]
}

func error_status(status int) error {
	return errors.New(fmt.Sprintf("Status Code %d received from TMDb", status))
}
