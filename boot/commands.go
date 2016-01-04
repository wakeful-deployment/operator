package boot

import (
	"fmt"
	"github.com/wakeful-deployment/operator/consul"
	"time"
)

func detectOrBootConsul(config *Config) error {
	return nil
}

func Boot(configPath string) (*Config, error) {
	config, err := ReadConfigFile(configPath)

	if err != nil {
		return nil, err
	}

	err = detectOrBootConsul(config)

	if err != nil {
		return nil, err
	}

	currentState, err := CurrentState()

	if err != nil {
		return nil, err
	}

	err = Normalize(config, currentState)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewConfigAndNormalize(config *Config, desiredState *consul.State) error {
	desiredConfig, err := NewConfig(config, desiredState)

	if err != nil {
		return err
	}

	currentState, err := CurrentState()

	if err != nil {
		return err
	}

	err = Normalize(desiredConfig, currentState)

	if err != nil {
		return err
	}

	return nil
}

func Once(currentConfig *Config) error {
	stateUrl := consul.StateURL{Wait: "0s"}
	desiredState, err := consul.DesiredState(stateUrl.String())

	if err != nil {
		return err
	}

	err = NewConfigAndNormalize(currentConfig, desiredState)

	if err != nil {
		return err
	}

	return nil
}

func Loop(currentConfig *Config) {
	stateUrl := consul.StateURL{Wait: "5m"}

	for {
		desiredState, err := consul.DesiredState(stateUrl.String()) // this will block for some time

		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		err = NewConfigAndNormalize(currentConfig, desiredState)

		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}

		stateUrl.Index = desiredState.Index // for next iteration

		time.Sleep(time.Second)
	}
}
