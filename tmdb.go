// Copyright 2014, Amahi.  All rights reserved.
// Use of this source code is governed by the
// license that can be found in the LICENSE file.

// Golang library for requesting metadata from themoviedb.org

package tmdb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

const BASE_URL string = "http://api.themoviedb.org/3"

type TMDB struct {
	api_key string
	config *tmdbConfig
}

func Init(api_key string) (* TMDB) {
	return &TMDB{api_key: api_key}
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

//Movie metadata structure
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

//The main call for getting movie data
func (tmdb *TMDB) getMovieData(MediaName string) (string, error) {
	var met string
	results, err := tmdb.searchMovie(MediaName)
	if err != nil {
		return met, err
	}
	if results.Total_results == 0 {
		return met, errors.New("No results found at TMDb")
	}
	if results.Results[0].Media_type == "person" {
		return met, errors.New("Metadata for persons not supported")
	} else if results.Results[0].Media_type == "tv" {
		return met, errors.New("Metadata for tv not supported inside a call for movie data")
	} else {

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
}

//search on TMDb for TV, persons and Movies with a given name
func (tmdb *TMDB) searchTmdbMulti(MediaName string) (tmdbResponse, error) {
	res, err := http.Get(BASE_URL + "/search/multi?api_key=" + tmdb.api_key + "&query=" + MediaName)
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, errors.New("Status Code 200 not recieved from TMDB")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

//search on TMDb for Movies with a given name
func (tmdb *TMDB) searchMovie(MediaName string) (tmdbResponse, error) {
	res, err := http.Get(BASE_URL + "/search/movie?api_key=" + tmdb.api_key + "&query=" + MediaName)
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, errors.New("Status Code 200 not recieved from TMDB")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

//search on TMDb for Tv Shows with a given name
func (tmdb *TMDB) searchTmdbTv(MediaName string) (tmdbResponse, error) {
	res, err := http.Get(BASE_URL + "/search/tv?api_key=" + tmdb.api_key + "&query=" + MediaName)
	var resp tmdbResponse
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, errors.New("Status Code 200 not recieved from TMDb")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return tmdbResponse{}, err
	}
	return resp, nil
}

//get configurations from TMDb
func (tmdb *TMDB) getConfig() (*tmdbConfig, error) {
	if tmdb.config.Images.Base_url == "" {
		res, err := http.Get(BASE_URL + "/configuration?api_key=" + tmdb.api_key)
		var conf = &tmdbConfig{}
		if err != nil {
			return conf, err
		}
		if res.StatusCode != 200 {
			return conf, errors.New("Status Code 200 not recieved from TMDb")
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

//get basic information for movie
func (tmdb *TMDB) getMovieDetails(MediaId string) (movieMetadata, error) {
	res, err := http.Get(BASE_URL + "/movie/" + MediaId + "?api_key=" + tmdb.api_key)
	var met movieMetadata
	if err != nil {
		return met, err
	}
	if res.StatusCode != 200 {
		return met, errors.New("Status Code 200 not recieved from TMDb")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &met)
	if err != nil {
		return movieMetadata{}, err
	}
	return met, nil
}

//get credits for movie
func (tmdb *TMDB) getMovieCredits(MediaId string) (tmdbCredits, error) {
	res, err := http.Get(BASE_URL + "/movie/" + MediaId + "/credits?api_key=" + tmdb.api_key)
	var cred tmdbCredits
	if err != nil {
		return cred, err
	}
	if res.StatusCode != 200 {
		return cred, errors.New("Status Code 200 not recieved from TMDb")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &cred)
	if err != nil {
		return tmdbCredits{}, err
	}
	return cred, nil
}

//get basic information for Tv
func (tmdb *TMDB) getTmdbTvDetails(MediaId string) (movieMetadata, error) {
	res, err := http.Get(BASE_URL + "/tv/" + MediaId + "?api_key=" + tmdb.api_key)
	var met movieMetadata
	if err != nil {
		return met, err
	}
	if res.StatusCode != 200 {
		return met, errors.New("Status Code 200 not recieved from TMDb")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &met)
	if err != nil {
		return movieMetadata{}, err
	}
	return met, nil
}

//get credits for Tv
func (tmdb *TMDB) getTmdbTvCredits(MediaId string) (tmdbCredits, error) {
	res, err := http.Get(BASE_URL + "/tv/" + MediaId + "/credits?api_key=" + tmdb.api_key)
	var cred tmdbCredits
	if err != nil {
		return cred, err
	}
	if res.StatusCode != 200 {
		return cred, errors.New("Status Code 200 not recieved from TMDb")
	}
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &cred)
	if err != nil {
		return tmdbCredits{}, err
	}
	return cred, nil
}

//filter out unwanted movie metadata before return to user
func (tmdb *TMDB) filterMovieData(data string) (string, error) {
	var f filtered_output
	var det movieMetadata
	err := json.Unmarshal([]byte(data), &det)
	if err != nil {
		return "", err
	}
	f.Title = det.Title
	f.Release_date = det.Release_date
	f.Release_date = f.Release_date[0:4]
	f.Artwork = det.Config.Images.Base_url + "original" + det.Poster_path

	metadata, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(metadata), nil
}
