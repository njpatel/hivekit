package hive

// State contains all the current information from the Hive Home system
// Note: The actual Hive is much more complex than this, but this is what
// 			 I need right now. Feel free to extract more information and pad
//			 out as necessary
type State struct {
	Heating        bool /* True: On, False: Off */
	HeatingBoosted bool
	CurrentTemp    float32 /* Celcius */
	TargetTemp     float32

	HotWater        bool /* True: On, False: Guess */
	HotWaterBoosted bool
}

func newStateFromNodes(nodes []nodeInfo) *State {
	state := &State{}
	for _, info := range nodes {
		attrs := info.Attributes

		if attrs.Temperature != nil && attrs.TargetHeatTemperature != nil {
			state.CurrentTemp = attrs.Temperature.ReportedValue
			state.TargetTemp = attrs.TargetHeatTemperature.ReportedValue

			if attrs.StateHeatingRelay != nil {
				state.Heating = attrs.StateHeatingRelay.ReportedValue == "ON"
			}

			if attrs.ActiveHeatCoolMode != nil {
				state.HeatingBoosted = attrs.ActiveHeatCoolMode.ReportedValue == "BOOST"
			}
		}

		if attrs.SupportsHotWater != nil && attrs.SupportsHotWater.ReportedValue == true {
			if attrs.StateHotWaterRelay != nil {
				state.HotWater = attrs.StateHotWaterRelay.ReportedValue == "ON"
			}

			if attrs.ActiveHeatCoolMode != nil {
				state.HotWaterBoosted = attrs.ActiveHeatCoolMode.ReportedValue == "BOOST"
			}
		}

	}
	return state
}
