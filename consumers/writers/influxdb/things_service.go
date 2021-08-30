package influxdb

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/mainflux/mainflux/things"
)

type ThingsService struct {
	Url    string
	Token  string
	Client http.Client
}

type ThingsServiceConfig struct {
	Token string
	Url   string
}

func NewThingsService(config *ThingsServiceConfig) *ThingsService {
	return &ThingsService{
		Token:  config.Token,
		Url:    config.Url,
		Client: http.Client{},
	}
}

func (thingService *ThingsService) GetThingMetaById(id string) (things.Metadata, error) {
	req, err := http.NewRequest("GET", thingService.Url+"/things/"+id, nil)

	if err != nil {
		return things.Metadata{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", thingService.Token)

	resp, err := thingService.Client.Do(req)
	if err != nil {
		return things.Metadata{}, err
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return things.Metadata{}, err
	}

	var thing things.Thing
	err = json.Unmarshal(content, &thing)

	if err != nil {
		return things.Metadata{}, err
	}
	return thing.Metadata, nil
}
