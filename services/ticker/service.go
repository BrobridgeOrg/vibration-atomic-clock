package timer

import (
	"math/rand"
	"strconv"
	"time"

	app "vibration-atomic-clock/app/interface"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	app       app.AppImpl
	ticker    *time.Ticker
	isRunning bool
	stopChan  chan bool
}

func CreateService(a app.AppImpl) *Service {

	// Preparing service
	service := &Service{
		app: a,
	}

	service.isRunning = false
	service.stopChan = make(chan bool)

	return service
}

func GenerateRangeNum() int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(1000-500) + 500
	return randNum
}

func (service *Service) RunTickerCluster() {
	log.Info("Stanby ...")
	timer := time.AfterFunc(1100*time.Millisecond,
		func() {
			service.StartTicker(100)
		},
	)

	oldData := ""
	//subscribe queue
	signalBus := service.app.GetSignalBus()
	signalBus.Subscribe(
		"timer.ticker",
		func(m *nats.Msg) {
			//log.Info(string(m.Data))
			timer.Reset(1100 * time.Millisecond)
			if oldData == string(m.Data) && service.isRunning == true {
				// Have multiple sender,  Stop ticker
				service.StopTicker()
				timer.Reset(time.Duration(GenerateRangeNum()) * time.Millisecond)
			}
			oldData = string(m.Data)
		})

}

func (service *Service) StartTicker(duration time.Duration) {

	if service.isRunning {
		//Ticker already is running.
		return
	}

	log.Info("Is Master.")
	// Start ticker
	service.ticker = time.NewTicker(duration * time.Millisecond)
	defer service.ticker.Stop()

	var old int64 = 0
	for {

		select {
		case <-service.ticker.C:
			service.isRunning = true
			now := time.Now().UTC().Unix()
			if now == old || now < old {
				continue
			}

			//Publish to queue
			signalBus := service.app.GetSignalBus()
			signalBus.Emit("timer.ticker", []byte(strconv.FormatInt(now, 10)))
			old = now
		case stop := <-service.stopChan:
			if stop {
				service.ticker.Stop()
				service.isRunning = false
				return
			}
		}
	}

}

func (service *Service) StopTicker() {

	// Stop timer
	if service.isRunning {
		//	service.ticker.Stop()
		service.stopChan <- true
		log.Info("Stop ticker....")
	}

}
