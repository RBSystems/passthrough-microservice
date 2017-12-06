package passthrough

type request struct {
	GW       string
	Path     string
	Interval int
	RespChan chan response
}

type response struct {
	RespCode    int
	ContentType string
	Content     []byte
	Error       error
}

type channelWrapper struct {
	Channel chan request
}
