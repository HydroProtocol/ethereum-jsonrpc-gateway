package core

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HydroProtocol/ethereum-jsonrpc-gateway/utils"
	"github.com/sirupsen/logrus"
)

type IStrategy interface {
	handle(*Request) ([]byte, error)
}

var _ IStrategy = &NaiveProxy{}
var _ IStrategy = &RaceProxy{}
var _ IStrategy = &FallbackProxy{}

type NaiveProxy struct{}

func newNaiveProxy() *NaiveProxy {
	return &NaiveProxy{}
}

func (p *NaiveProxy) handle(req *Request) ([]byte, error) {
	upstream := currentRunningConfig.Upstreams[0]
	bts, err := upstream.handle(req)

	if err != nil {
		return nil, err
	}

	return bts, err
}

type RaceProxy struct{}

func newRaceProxy() *RaceProxy {
	return &RaceProxy{}
}

func (p *RaceProxy) handle(req *Request) ([]byte, error) {
	startAt := time.Now()

	defer func() {
		logrus.Debugf("geth_gateway %f", float64(time.Since(startAt))/1000000)
	}()

	successfulResponse := make(chan []byte, len(currentRunningConfig.Upstreams))
	failedResponse := make(chan []byte, len(currentRunningConfig.Upstreams))
	errorResponseUpstreams := make(chan Upstream, len(currentRunningConfig.Upstreams))

	for _, upstream := range currentRunningConfig.Upstreams {
		go func(upstream Upstream) {
			defer func() {
				if err := recover(); err != nil {
					req.logger.Debugf("%v Upstream %s failed, err: %v\n", time.Now().Sub(startAt), upstream, err)
					errorResponseUpstreams <- upstream
				}
			}()

			bts, err := upstream.handle(req)

			if err != nil {
				req.logger.Debugf("%vms Upstream: %v, Error: %v\n", time.Now().Sub(startAt), upstream, err)
				failedResponse <- nil
				return
			}

			resBody := strings.TrimSpace(string(bts))

			diff := time.Now().Sub(startAt)
			if utils.NoErrorFieldInJSON(resBody) {
				req.logger.Debugf("%v Upstream: %v Success, Body: %v\n", diff, upstream, resBody)
				successfulResponse <- bts
			} else {
				req.logger.Debugf("%v Upstream: %v Failed, Body: %v\n", diff, upstream, resBody)
				failedResponse <- bts
			}
		}(upstream)
	}

	errorCount := 0

	for errorCount < len(currentRunningConfig.Upstreams) {
		select {
		case <-time.After(time.Second * 10):
			req.logger.Debugf("%v Final Timeout\n", time.Now().Sub(startAt))
			return nil, TimeoutError
		case res := <-successfulResponse:
			req.logger.Debugf("%v Final Success\n", time.Now().Sub(startAt))
			return res, nil
		case res := <-failedResponse:
			return res, nil
		case <-errorResponseUpstreams:
			errorCount++
		}
	}

	req.logger.Errorf("%v Final Failed\n", time.Now().Sub(startAt))

	logrus.Errorf("geth_gateway_fail")

	return nil, AllUpstreamsFailedError
}

type FallbackProxy struct {
	currentUpstreamIndex *atomic.Value
	upsteamStatus        *sync.Map
}

func newFallbackProxy() *FallbackProxy {
	v := &atomic.Value{}
	v.Store(0)

	p := &FallbackProxy{
		currentUpstreamIndex: v,
		upsteamStatus:        &sync.Map{},
	}

	for i := 0; i < len(currentRunningConfig.Upstreams); i++ {
		p.upsteamStatus.Store(i, true)
	}

	return p
}

func (p *FallbackProxy) handle(req *Request) ([]byte, error) {
	for i := 0; i < len(currentRunningConfig.Upstreams); i++ {
		index := p.currentUpstreamIndex.Load().(int)

		value, _ := p.upsteamStatus.Load(index)
		isUpstreamValid := value.(bool)

		if isUpstreamValid {
			bts, err := currentRunningConfig.Upstreams[index].handle(req)

			if err != nil {
				nextUpstreamIndex := int(math.Mod(float64(index+1), float64(len(currentRunningConfig.Upstreams))))
				p.currentUpstreamIndex.Store(nextUpstreamIndex)
				p.upsteamStatus.Store(i, false)

				logrus.Infof("upstream %d return err, switch to %d", index, nextUpstreamIndex)

				go func(i int) {
					<-time.After(5 * time.Second)
					p.upsteamStatus.Store(i, true)
				}(index)

				continue
			} else {
				return bts, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid upstream")
}
