package bosh

import (
	"fmt"
	"io/ioutil"
	"os"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

func buildUAA(host string, cacert string, logger boshlog.Logger) (boshuaa.UAA, error) {
	uaaConfig, err := boshuaa.NewConfigFromURL(fmt.Sprintf("https://%s:8443", host))
	if err != nil {
		return nil, err
	}
	uaaConfig.Client = os.Getenv("BOSH_CLIENT")
	uaaConfig.ClientSecret = os.Getenv("BOSH_CLIENT_SECRET")
	uaaConfig.CACert = cacert
	uaaFactory := boshuaa.NewFactory(logger)
	return uaaFactory.New(uaaConfig)
}

func getDirector() (boshdir.Director, error) {
	logger := boshlog.NewLogger(boshlog.LevelInfo)
	boshFactory := boshdir.NewFactory(logger)

	factoryConfig, err := boshdir.NewConfigFromURL(os.Getenv("BOSH_ENVIRONMENT"))
	if err != nil {
		return nil, err
	}

	cacertEnvVar := os.Getenv("BOSH_CA_CERT")
	if _, err = os.Stat(cacertEnvVar); err != nil {
		factoryConfig.CACert = cacertEnvVar
	} else {
		cert, err := ioutil.ReadFile(cacertEnvVar)
		if err != nil {
			return nil, err
		}
		factoryConfig.CACert = string(cert)
	}

	uaa, err := buildUAA(factoryConfig.Host, factoryConfig.CACert, logger)
	if err != nil {
		return nil, err
	}
	factoryConfig.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc

	director, err := boshFactory.New(factoryConfig, nil, nil)
	if err != nil {
		return nil, err
	}
	return director, nil
}

func NewFromEnv() (boshdir.Director, error) {
	return getDirector()
}
