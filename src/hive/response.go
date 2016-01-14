package hive

type errorReply struct {
	Reason string
}

type loginReply struct {
	Token string `json:"ApiSession"`
	Error errorReply
}

type nodesReply struct {
	Nodes []nodeInfo
}

type nodeInfo struct {
	ID         string
	Href       string
	LastSeen   int64
	Attributes nodeAttributes
}

type nodeAttributes struct {
	Temperature           *nodeReportFloat
	SupportsHotWater      *nodeReportBool
	StateHotWaterRelay    *nodeReportString
	StateHeatingRelay     *nodeReportString
	ActiveHeatCoolMode    *nodeReportString
	TargetHeatTemperature *nodeReportFloat
}

type nodeReportFloat struct {
	ReportedValue float32
	TargetValue   float32
}

type nodeReportBool struct {
	ReportedValue bool
}

type nodeReportString struct {
	ReportedValue string
	TargetValue   string
}
