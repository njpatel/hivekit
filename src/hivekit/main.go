package main

import (
	"flag"
	"os"
	"sync"
	"time"

	"hive"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/hap"
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

	thermostat         accessory.Thermostat
	hotWaterSwitch     accessory.Switch
	heatingBoostSwitch accessory.Switch
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
	aInfo := accessory.Info{
		Name:         "Hive Bridge",
		Manufacturer: "British Gas PLC",
	}
	a := accessory.New(aInfo, accessory.TypeBridge)

	tInfo := accessory.Info{
		Name:         "Heating",
		Manufacturer: "British Gas PLC",
	}
	t := accessory.NewThermostat(tInfo, 20.0, hive.MinTemp, hive.MaxTemp, 0.5)
	t.Thermostat.TargetTemperature.OnValueUpdate(targetTempChangeRequest)
	t.Thermostat.OnTargetModeChange(targetModeChangeRequest)
	thermostat = t

	bInfo := accessory.Info{
		Name:         "Heating Boost",
		Manufacturer: "British Gas PLC",
	}
	b := accessory.NewSwitch(bInfo)
	b.OnStateChanged(heatingBoostStateChangeRequest)
	heatingBoostSwitch = b

	sInfo := accessory.Info{
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

		log.Printf("[VERB] Syncing status with HomeKit\n")

		hotWaterSwitch.SetOn(state.HotWater)

		heatingBoostSwitch.SetOn(state.HeatingBoosted)

		thermostat.SetTemperature(state.CurrentTemp)
		thermostat.SetTargetTemperature(state.TargetTemp)
		thermostat.SetMode(modeForHiveMode(state.CurrentHeatingMode))
		thermostat.SetMode(accessory.HeatCoolModeAuto)
		thermostat.SetTargetMode(modeForHiveMode(state.TargetHeatingMode))
	})
}

func modeForHiveMode(mode hive.HeatCoolMode) accessory.HeatCoolModeType {
	if mode == hive.HeatCoolModeOff {
		return accessory.HeatCoolModeOff
	}
	return accessory.HeatCoolModeAuto
}

func hotWaterStateChangeRequest(on bool) {
	err := hiveHome.ToggleHotWater(on, time.Minute*time.Duration(hotWaterDuration))
	if err != nil {
		log.Printf("[WARN] Unable to toggle hot water: %v\n", err)
	}
}

func targetTempChangeRequest(c *Characteristic, newValue, oldValue interface{}) {
	err := hiveHome.SetTargetTemp(float64(temp))
	if err != nil {
		log.Printf("[WARN] Unable to set target temperature: %v\n", err)
	}
}

func targetModeChangeRequest(hcMode accessory.HeatCoolModeType) {
	log.Printf("[Warn] Changing target mode is unsupported at this time")
}

func heatingBoostStateChangeRequest(on bool) {
	err := hiveHome.ToggleHeatingBoost(on, time.Minute*time.Duration(boostDuration))
	if err != nil {
		log.Printf("[WARN] Unable to set heating boost: %v\n", err)
	}
}
