package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"hive"

	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/model"
	"github.com/brutella/hc/model/accessory"
	"github.com/brutella/log"
)

var (
	username string
	password string
	pin      string
	verbose  bool

	hiveHome *hive.Hive

	thermostat      model.Thermostat
	hotWaterSwitch  model.Switch
	transport       hap.Transport
	accessoryUpdate sync.Mutex
)

func init() {
	flag.StringVar(&username, "username", os.Getenv("HIVEKIT_USER"), "Hive Home web service username (usually an email address)")
	flag.StringVar(&password, "password", os.Getenv("HIVEKIT_PASS"), "Hive Home web service password")
	flag.StringVar(&pin, "pin", os.Getenv("HIVEKIT_PIN"), "The HomeKit accessory pin (8 numeric chars)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
}

func main() {
	flag.Parse()

	log.Verbose = verbose

	setupHomeKit()
	setupHive()

	transport.Start()
}

func setupHomeKit() {
	tInfo := model.Info{
		Name:         "Heating",
		Manufacturer: "British Gas PLC",
	}
	t := accessory.NewThermostat(tInfo, 20.0, hive.MinTemp, hive.MaxTemp, 0.5)
	thermostat = t

	sInfo := model.Info{
		Name:         "Hot Water",
		Manufacturer: "British Gas PLC",
	}
	h := accessory.NewSwitch(sInfo)
	h.OnStateChanged(hotWaterStateChangeRequest)
	hotWaterSwitch = h

	config := hap.Config{
		Pin: pin,
	}

	var err error
	transport, err = hap.NewIPTransport(config, t.Accessory, h.Accessory)
	if err != nil {
		log.Fatal(err)
	}
}

func setupHive() {
	// Connect to Hive
	var err error
	hiveHome, err = hive.Connect(hive.Config{
		Username:        username,
		Password:        password,
		RefreshInterval: 60 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	hiveHome.HandleStateChange(func(state *hive.State) {
		accessoryUpdate.Lock()
		defer accessoryUpdate.Unlock()

		hotWaterSwitch.SetOn(state.HotWater)

		thermostat.SetTemperature(state.CurrentTemp)
		thermostat.SetTargetTemperature(state.TargetTemp)
		thermostat.SetMode(modeForHiveMode(state.CurrentHeatingMode))
		thermostat.SetTargetMode(modeForHiveMode(state.TargetHeatingMode))
	})
}

func modeForHiveMode(mode hive.HeatCoolMode) model.HeatCoolModeType {
	if mode == hive.HeatCoolModeOff {
		return model.HeatCoolModeOff
	} else if mode == hive.HeatCoolModeHeating {
		return model.HeatCoolModeHeat
	}
	return model.HeatCoolModeAuto
}

func hotWaterStateChangeRequest(on bool) {
	err := hiveHome.ToggleHotWater(on, time.Minute*60)
	if err != nil {
		fmt.Printf("Unable to toggle hot water: %v\n", err)
	}
}
