package hive

// State contains all the current information from the Hive Home system
// Note: The actual Hive is much more complex than this, but this is what
// 			 I need right now. Feel free to extract more information and pad
//			 out as necessary
type State struct {
	Heating        bool /* True: On, False: Off */
	HeatingBoosted bool

	CurrentHeatingMode HeatCoolMode
	TargetHeatingMode  HeatCoolMode

	CurrentTemp float64 /* Celcius */
	TargetTemp  float64

	HotWater        bool /* True: On, False: Guess */
	HotWaterBoosted bool

	heatingNodeID  string
	hotWaterNodeID string
}

// HeatCoolMode is the mode of the Hive Home unit
type HeatCoolMode int

// HeatCoolModeOff ...
const (
	HeatCoolModeOff HeatCoolMode = iota
	HeatCoolModeHeating
	HeatCoolModeScheduled
)

func newStateFromNodes(nodes []nodeInfo) *State {
	state := &State{}
	for _, info := range nodes {
		attrs := info.Attributes

		if attrs.Temperature != nil && attrs.TargetHeatTemperature != nil {
			state.heatingNodeID = info.ID
			state.CurrentTemp = attrs.Temperature.ReportedValue
			state.TargetTemp = attrs.TargetHeatTemperature.ReportedValue

			if attrs.StateHeatingRelay != nil {
				state.Heating = attrs.StateHeatingRelay.ReportedValue == apiOn
			}

			if attrs.ActiveHeatCoolMode != nil {
				reported := attrs.ActiveHeatCoolMode.ReportedValue
				state.HeatingBoosted = reported == apiBoost

				state.CurrentHeatingMode = HeatCoolModeScheduled
				if state.HeatingBoosted {
					state.CurrentHeatingMode = HeatCoolModeHeating
				} else if reported == apiOff {
					state.CurrentHeatingMode = HeatCoolModeOff
				}

				state.TargetHeatingMode = HeatCoolModeScheduled
				target := attrs.ActiveHeatCoolMode.TargetValue
				if target == apiBoost {
					state.TargetHeatingMode = HeatCoolModeHeating
				} else if target == apiOff {
					state.TargetHeatingMode = HeatCoolModeOff
				}
			}
		}

		if attrs.SupportsHotWater != nil && attrs.SupportsHotWater.ReportedValue == true {
			state.hotWaterNodeID = info.ID
			if attrs.StateHotWaterRelay != nil {
				state.HotWater = attrs.StateHotWaterRelay.ReportedValue == apiOn
			}

			if attrs.ActiveHeatCoolMode != nil {
				state.HotWaterBoosted = attrs.ActiveHeatCoolMode.ReportedValue == apiBoost
			}
		}

	}
	return state
}
