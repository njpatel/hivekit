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
	username         string
	password         string
	pin              string
	boostDuration    int64
	hotWaterDuration int64
	verbose          bool

	hiveHome *hive.Hive

	thermostat         model.Thermostat
	hotWaterSwitch     model.Switch
	heatingBoostSwitch model.Switch
	transport          hap.Transport
	accessoryUpdate    sync.Mutex
)

func init() {
	flag.StringVar(&username, "username", os.Getenv("HIVEKIT_USER"), "Hive Home web service username (usually an email address)")
	flag.StringVar(&password, "password", os.Getenv("HIVEKIT_PASS"), "Hive Home web service password")
	flag.StringVar(&pin, "pin", os.Getenv("HIVEKIT_PIN"), "The HomeKit accessory pin (8 numeric chars)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Int64Var(&boostDuration, "boost-duration", 60, "Duration (minutes) to boost heating")
	flag.Int64Var(&hotWaterDuration, "boost-water", 60, "Duration (minutes) to boost hot water")
}

func main() {
	flag.Parse()

	log.Verbose = verbose

	setupHomeKit()
	setupHive()

	transport.Start()
}

func setupHomeKit() {
	aInfo := model.Info{
		Name:         "Hive Bridge",
		Manufacturer: "British Gas PLC",
	}
	a := accessory.New(aInfo)

	tInfo := model.Info{
		Name:         "Heating",
		Manufacturer: "British Gas PLC",
	}
	t := accessory.NewThermostat(tInfo, 20.0, hive.MinTemp, hive.MaxTemp, 0.5)
	t.OnTargetTempChange(targetTempChangeRequest)
	t.OnTargetModeChange(targetModeChangeRequest)
	thermostat = t

	bInfo := model.Info{
		Name:         "Heating Boost",
		Manufacturer: "British Gas PLC",
	}
	b := accessory.NewSwitch(bInfo)
	b.OnStateChanged(heatingBoostStateChangeRequest)
	heatingBoostSwitch = b

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
	transport, err = hap.NewIPTransport(config, a, t.Accessory, b.Accessory, h.Accessory)
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
		RefreshInterval: 30 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	hiveHome.HandleStateChange(func(state *hive.State) {
		accessoryUpdate.Lock()
		defer accessoryUpdate.Unlock()

		fmt.Println("Syncing status with HomeKit")

		hotWaterSwitch.SetOn(state.HotWater)

		heatingBoostSwitch.SetOn(state.HeatingBoosted)

		thermostat.SetTemperature(state.CurrentTemp)
		thermostat.SetTargetTemperature(state.TargetTemp)
		thermostat.SetMode(modeForHiveMode(state.CurrentHeatingMode))
		thermostat.SetMode(model.HeatCoolModeAuto)
		thermostat.SetTargetMode(modeForHiveMode(state.TargetHeatingMode))
	})
}

func modeForHiveMode(mode hive.HeatCoolMode) model.HeatCoolModeType {
	if mode == hive.HeatCoolModeOff {
		return model.HeatCoolModeOff
	}
	return model.HeatCoolModeAuto
}

func hotWaterStateChangeRequest(on bool) {
	err := hiveHome.ToggleHotWater(on, time.Minute*time.Duration(hotWaterDuration))
	if err != nil {
		fmt.Printf("Unable to toggle hot water: %v\n", err)
	}
}

func targetTempChangeRequest(temp float64) {
	err := hiveHome.SetTargetTemp(temp)
	if err != nil {
		fmt.Printf("Unable to set target temperature: %v\n", err)
	}
}

func targetModeChangeRequest(hcMode model.HeatCoolModeType) {
	fmt.Printf("Chaning target mode is unsupported at this time")
}

func heatingBoostStateChangeRequest(on bool) {
	err := hiveHome.ToggleHeatingBoost(on, time.Minute*time.Duration(boostDuration))
	if err != nil {
		fmt.Printf("Unable to set heating boost: %v\n", err)
	}
}
