package edged

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Bootstrap(configFilePath, bootstrapServerURL string, logger *log.Logger) error {
	macAddresses, err := macAddresses()
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading MAC addresses on the device %s", err))
		return err
	}

	var cfg []byte = nil
	for _, address := range macAddresses {
		logger.Debug(fmt.Sprintf("Requesting config for %s from %s", address, bootstrapServerURL))
		cfg, err = config(address, bootstrapServerURL)
		if err != nil {
			logger.Debug(fmt.Sprintf("Getting config for %s from %s failed: %s", address, bootstrapServerURL, err))
			continue
		}
		logger.Debug(fmt.Sprintf("Getting config for %s from %s failed: %s", address, bootstrapServerURL, err))
		break
	}
	if cfg == nil {
		return errors.New(fmt.Sprintf("Configuration retrieval unsuccessful: %s", err.Error()))
	}
	logger.Debug(fmt.Sprintf("Got config %s", cfg))

	c := make(map[string]string)
	err = json.Unmarshal(cfg, &c)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshalling config %s", err))
		return err
	}
	file, err := os.Create(configFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating config file %s", err))
		return err
	}
	defer file.Close()
	for k, v := range c {
		// Indent metadata fields.
		if k == "metadata" {
			fmt.Fprintf(file, "%s:\n\t%s\n", k, strings.Replace(v, "\n", "\n\t", -1))
			continue
		}
		fmt.Fprintf(file, "%s: %s\n", k, v)
	}
	logger.Debug(fmt.Sprintf("Wrote new config retrieved from bootstrap server"))
	return nil
}

func macAddresses() ([]string, error) {
	var addresses []string
	cmd := "cat /sys/class/net/$(ip route show default | awk '/default/ {print $5}')/address"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}
	addresses = append(addresses, strings.Trim(string(out), "\n"))
	return addresses, nil
}

func config(address, bootstrapServerURL string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", bootstrapServerURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", address)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}
