package app

import (
	"strconv"

	app "twist-atomic-clock/app/interface"
	"twist-atomic-clock/app/signalbus"
	ticker "twist-atomic-clock/services/ticker"

	log "github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"github.com/spf13/viper"
)

type App struct {
	id        uint64
	flake     *sonyflake.Sonyflake
	signalbus *signalbus.SignalBus
}

func CreateApp() *App {

	// Genereate a unique ID for instance
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		return nil
	}

	idStr := strconv.FormatUint(id, 16)

	return &App{
		id:    id,
		flake: flake,
		signalbus: signalbus.CreateConnector(
			viper.GetString("signal_server.host"),
			idStr,
		),
	}
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
	tickerService.StartTicker(100) //unit: ms

	return nil
}

func (a *App) GetSignalBus() app.SignalBusImpl {
	return app.SignalBusImpl(a.signalbus)
}
