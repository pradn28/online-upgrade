package util

/*
This file is somewhat inspired by the state tracking code in Terraform
	https://github.com/hashicorp/terraform
	helper/resource/state.go
*/

import (
	"fmt"
	"math"
	"time"
)

type StateRefreshFunc func() (state interface{}, err error)
type StateNextRefreshFunc func(tries int) time.Duration

type StateChange struct {
	Target      interface{}
	Timeout     time.Duration
	NextRefresh StateNextRefreshFunc
	Refresh     StateRefreshFunc
}

func ConstantRefresh(sleepTime time.Duration) func(int) time.Duration {
	return func(_ int) time.Duration {
		return sleepTime
	}
}

func ExponentialBackoffRefresh(tries int) time.Duration {
	n := math.Pow(2, float64(tries))
	wait := time.Duration(n) * 100 * time.Millisecond
	if wait > 10*time.Second {
		wait = 10 * time.Second
	}
	return wait
}

var defaultRefreshFunc = ConstantRefresh(1)

func (conf StateChange) WaitForState() error {
	var result interface{}
	var resultErr error

	// the doneFlag is for the controller to inform the worker to
	// exit in the case of timeout
	doneFlag := false
	// the doneCh is for the worker to inform the controller that
	// it has reached it's terminal state
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		for tries := 0; ; tries++ {
			if doneFlag {
				// the controller timed out
				return
			}

			if conf.NextRefresh == nil {
				time.Sleep(defaultRefreshFunc(tries))
			} else {
				time.Sleep(conf.NextRefresh(tries))
			}

			result, resultErr = conf.Refresh()
			if resultErr != nil {
				return
			}

			if result == conf.Target {
				return
			}
		}
	}()

	select {
	case <-doneCh:
		return resultErr
	case <-time.After(conf.Timeout):
		doneFlag = true
		return fmt.Errorf(
			"timeout while waiting for state to become '%v'",
			conf.Target)
	}
}
