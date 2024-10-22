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
	Version int8
	MsgType int8
	Flag    int16
	MsgId   int32
	Headers Headers
	Content []byte
}
