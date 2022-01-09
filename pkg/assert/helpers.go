package assert

import (
	"fmt"
	"strings"
	"time"
)

const defaultMaxWait = "60s"

func getMaxRetries(waitTime string, interval int) int {

	// Get max wait time and retries/interval
	maxWait, err := time.ParseDuration(waitTime)
	if err != nil {
		maxWait, _ = time.ParseDuration(defaultMaxWait)
	}
	return int(maxWait.Seconds()) / interval

}

func unpackResource(resourcePath string) (string, string, string, error) {

	path := strings.TrimSuffix(strings.TrimPrefix(resourcePath, ":"), ":")
	nvk := strings.Split(path, ":")

	if len(nvk) == 2 {
		return nvk[0], nvk[1], "", nil
	}

	if len(nvk) == 3 {
		return nvk[0], nvk[1], nvk[2], nil
	}

	return "", "", "", fmt.Errorf("can't unpack resource path, wrong syntax")

}
