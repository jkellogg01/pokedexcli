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

type ResourceList struct {
	Count   int           `json:"count"`
	Next    *string       `json:"next"`
	Prev    *string       `json:"previous"`
	Results []APIResource `json:"results"`
}

type APIResource struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (svc *ApiService) GetLocations(url string) (*ResourceList, error) {
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	}
	raw, err := svc.Get(url)
	if err != nil {
		return nil, err
	}
	data := new(ResourceList)
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
	Name      string        `json:"name"`
	BaseXP    int           `json:"base_experience"`
	Height    int           `json:"height"`
	IsDefault bool          `json:"is_default"`
	Order     int           `json:"order"`
	Weight    int           `json:"weight"`
	Abilities []PkmnAbility `json:"abilities"`
	Forms     []APIResource `json:"forms"`
	Moves     []PkmnMove    `json:"moves"`
	Stats     []PkmnStat    `json:"stats"`
	Types     []PkmnType    `json:"types"`
}

type PkmnAbility struct {
	IsHidden bool        `json:"is_hidden"`
	Slot     int         `json:"slot"`
	Ability  APIResource `json:"ability"`
}

type PkmnMove struct {
	Move                APIResource       `json:"move"`
	VersionGroupDetails []PkmnMoveVersion `json:"version_group_details"`
}

type PkmnMoveVersion struct {
	LearnMethod  APIResource `json:"move_learn_method"`
	VersionGroup APIResource `json:"version_group"`
	LearnedAt    int         `json:"level_learned_at"`
}

type PkmnStat struct {
	Stat      APIResource `json:"stat"`
	EffortVal int         `json:"effort"`
	Base      int         `json:"base_stat"`
}

type PkmnType struct {
	Slot int         `json:"slot"`
	Type APIResource `json:"type"`
}

func (svc *ApiService) GetPkmn(name string) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)
	raw, err := svc.Get(url)
	if err != nil {
		return Pokemon{}, err
	}
	var data Pokemon
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return Pokemon{}, err
	}
	return data, nil
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
