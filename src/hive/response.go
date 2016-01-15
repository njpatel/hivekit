package hive

const (
	apiOff   = "OFF"
	apiOn    = "ON"
	apiBoost = "BOOST"
	apiHeat  = "HEAT"
)

type errorReply struct {
	Reason string
}

type loginReply struct {
	Token string `json:"ApiSession"`
	Error errorReply
}

type nodesReply struct {
	Nodes []nodeInfo `json:"nodes"`
}

type nodeInfo struct {
	ID         string
	Href       string
	LastSeen   int64
	Attributes nodeAttributes `json:"attributes"`
}

type nodeAttributes struct {
	ActiveHeatCoolMode    *nodeReportString `json:"activeHeatCoolMode,omitempty"`
	ActiveScheduleLock    *nodeReportBool   `json:"activeScheduleLock,omitempty"`
	ScheduleLockDuration  *nodeReportInt    `json:"scheduleLockDuration,omitempty"`
	StateHeatingRelay     *nodeReportString `json:",omitempty"`
	StateHotWaterRelay    *nodeReportString `json:",omitempty"`
	SupportsHotWater      *nodeReportBool   `json:",omitempty"`
	TargetHeatTemperature *nodeReportFloat  `json:"targetHeatTemperature,omitempty"`
	Temperature           *nodeReportFloat  `json:",omitempty"`
}

type nodeReportFloat struct {
	ReportedValue float64 `json:",omitempty"`
	TargetValue   float64 `json:"targetValue"`
}

type nodeReportBool struct {
	ReportedValue bool `json:",omitempty"`
	TargetValue   bool `json:"targetValue"`
}

type nodeReportString struct {
	ReportedValue string `json:",omitempty"`
	TargetValue   string `json:"targetValue"`
}

type nodeReportInt struct {
	ReportedValue int32 `json:",omitempty"`
	TargetValue   int32 `json:"targetValue"`
}
