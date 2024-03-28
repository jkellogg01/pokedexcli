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
	data := new(PaginatedLocations)
	if v, ok := svc.cache.Get(url); ok {
		log.Debug("Using cached data")
		err := json.Unmarshal(v, &data)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("Fetching map data")
		rawData, err := svc.Get(url)
		if err != nil {
			return nil, err
		}
		log.Debug("Caching fetched map data")
		svc.cache.Add(url, rawData)
		log.Debug("Decoding JSON response")
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

type LocationArea struct {
	PKMNEncounters []struct {
		PKMN struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func (svc *ApiService) GetPokemon(name string) ([]string, error) {
	var data LocationArea
	if v, ok := svc.cache.Get(name); ok {
		log.Debug("Using cached data")
		err := json.Unmarshal(v, &data)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("Fetching explore data")
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", name)
		rawData, err := svc.Get(url)
		if err != nil {
			return nil, err
		}
		log.Debug("Caching fetched explore data")
		svc.cache.Add(url, rawData)
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return nil, err
		}
	}
	result := make([]string, len(data.PKMNEncounters))
	for i, pkmn := range data.PKMNEncounters {
		result[i] = pkmn.PKMN.Name
	}
	return result, nil

}

func (svc *ApiService) Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API returned non-2XX status: %v", res.Status)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
