package hive

type errorReply struct {
	Reason string
}

type loginReply struct {
	Token string `json:"ApiSession"`
	Error errorReply
}
