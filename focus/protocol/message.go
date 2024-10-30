package protocol

// Focus protocol structure looks like:
// 0  Version     Type        Reserved Flag   32
// |----------|----------|----------/----------|
//
//	Exchange Sequence
//
// |----------|----------|----------|----------|
//
//	Header Length
//
// |----------|----------|----------|----------|
//
//	Header Content
//
// |----------/----------/----------/----------|
//
//	Body Length
//
// |----------|----------|----------|----------|
//
//	Body Content
//
// |----------/----------/----------/----------|
type Message struct {
	Version  int8
	MsgType  int8
	Status   int16
	Sequence int32
	Headers  Headers
	Content  []byte
}
