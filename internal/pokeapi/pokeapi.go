package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jkellogg01/pokedexcli/internal/pokecache"
)

type ApiService struct {
	cache *pokecache.Cache
}

func NewApiService(interval time.Duration) ApiService {
	log.Debug("Initializing cache for api service")
	cache := pokecache.NewCache(interval)
	return ApiService{cache: cache}
}

type PaginatedLocations struct {
	Count   int               `json:"count"`
	Next    *string           `json:"next"`
	Prev    *string           `json:"previous"`
	Results []LocationGeneral `json:"results"`
}

type LocationGeneral struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (svc *ApiService) GetLocations(url string) (*PaginatedLocations, error) {
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	}
	raw, err := svc.Get(url)
	if err != nil {
		return nil, err
	}
	data := new(PaginatedLocations)
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}
	return data, err
}

type LocationArea struct {
	PKMNEncounters []struct {
		PKMN struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func (svc *ApiService) GetLocationPkmn(name string) ([]string, error) {
	var data LocationArea
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", name)
	raw, err := svc.Get(url)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(data.PKMNEncounters))
	for i, pkmn := range data.PKMNEncounters {
		result[i] = pkmn.PKMN.Name
	}
	return result, nil
}

type Pokemon struct {
	Name   string `json:"name"`
	BaseXP int    `json:"base_experience"`
}

func (svc *ApiService) Get(url string) ([]byte, error) {
	data, ok := svc.cache.Get(url)
	if ok {
		return data, nil
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API returned non-2XX status: %v", res.Status)
	}
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	svc.cache.Add(url, data)
	return data, nil
}
