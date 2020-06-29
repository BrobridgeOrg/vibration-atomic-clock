package app

import (
	"strconv"
	"time"

	app "vibration-atomic-clock/app/interface"
	"vibration-atomic-clock/app/signalbus"
	ticker "vibration-atomic-clock/services/ticker"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"github.com/spf13/viper"
)

type App struct {
	id        uint64
	flake     *sonyflake.Sonyflake
	signalbus *signalbus.SignalBus
	isReady   bool
}

func CreateApp() *App {

	// Genereate a unique ID for instance
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return nil
	}

	idStr := strconv.FormatUint(id, 16)

	a := &App{
		id:    id,
		flake: flake,
	}

	a.signalbus = signalbus.CreateConnector(
		viper.GetString("signal_server.host"),
		idStr,
		func(natsConn *nats.Conn) {
			for {
				log.Warn("re-connect to signal server")

				// Connect to NATS Server
				err := a.signalbus.Connect()
				if err != nil {
					log.Error("Failed to connect to signal server")
					time.Sleep(time.Duration(1) * time.Second)
					continue
				}

				a.isReady = true

				break
			}
		},
		func(natsConn *nats.Conn) {
			a.isReady = false
		},
	)

	return a
}

func (a *App) Init() error {

	log.WithFields(log.Fields{
		"a_id": a.id,
	}).Info("Starting application")

	// Connect to signal server
	err := a.signalbus.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Uninit() {
}

func (a *App) Run() error {

	tickerService := ticker.CreateService(app.AppImpl(a))
	if viper.GetBool("atomic_clock.ha_mode") {
		log.Info("Run at high availability mode")
		tickerService.RunTickerCluster()
	} else {
		log.Info("Run at single mode")
		tickerService.StartTicker(100) //unit: ms
	}

	return nil
}

func (a *App) GetSignalBus() app.SignalBusImpl {
	return app.SignalBusImpl(a.signalbus)
}
