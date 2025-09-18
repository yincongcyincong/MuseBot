package utils

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/yincongcyincong/MuseBot/logger"
)

var (
	errNoVersionAndSize              = errors.New("no protocol version and header size byte")
	errNoTypeAndFlag                 = errors.New("no message type and specific flag byte")
	errNoSerializationAndCompression = errors.New("no serialization and compression method byte")
	errRedundantBytes                = errors.New("there are redundant bytes in data")
	errInvalidMessageType            = errors.New("invalid message type bits")
	errInvalidSerialization          = errors.New("invalid serialization bits")
	errInvalidCompression            = errors.New("invalid compression bits")
	errNoEnoughHeaderBytes           = errors.New("no enough header bytes")
	errReadEvent                     = errors.New("read event number")
	errReadSessionIDSize             = errors.New("read session ID size")
	errReadConnectIDSize             = errors.New("read connection ID size")
	errReadPayloadSize               = errors.New("read payload size")
	errReadSequence                  = errors.New("read sequence number")
	errReadErrorCode                 = errors.New("read error code")
	
	protocol = NewBinaryProtocol()
	
	wsWriteLock sync.Mutex
)

type (
	// MsgType defines message type which determines how the message will be
	// serialized with the 
	MsgType int32
	// MsgTypeFlagBits defines the 4-bit message-type specific flags. The specific
	// values should be defined in each specific usage scenario.
	MsgTypeFlagBits uint8
	
	// VersionBits defines the 4-bit version type.
	VersionBits uint8
	// HeaderSizeBits defines the 4-bit header-size type.
	HeaderSizeBits uint8
	// SerializationBits defines the 4-bit serialization method type.
	SerializationBits uint8
	// CompressionBits defines the 4-bit compression method type.
	CompressionBits uint8
)

// Values that a MsgType variable can take.
const (
	MsgTypeInvalid MsgType = iota
	MsgTypeFullClient
	MsgTypeAudioOnlyClient
	MsgTypeFullServer
	MsgTypeAudioOnlyServer
	MsgTypeFrontEndResultServer
	MsgTypeError
	
	MsgTypeServerACK = MsgTypeAudioOnlyServer
)

func (t MsgType) String() string {
	switch t {
	case MsgTypeFullClient:
		return "FullClient"
	case MsgTypeAudioOnlyClient:
		return "AudioOnlyClient"
	case MsgTypeFullServer:
		return "FullServer"
	case MsgTypeAudioOnlyServer:
		return "AudioOnlyServer/ServerACK"
	case MsgTypeError:
		return "Error"
	case MsgTypeFrontEndResultServer:
		return "TtsFrontEndResult"
	default:
		return fmt.Sprintf("invalid message type: %d", t)
	}
}

// Values that a MsgTypeFlagBits variable can take.
const (
	// For common 
	MsgTypeFlagNoSeq       MsgTypeFlagBits = 0     // Non-terminal packet with no sequence
	MsgTypeFlagPositiveSeq MsgTypeFlagBits = 0b1   // Non-terminal packet with sequence > 0
	MsgTypeFlagLastNoSeq   MsgTypeFlagBits = 0b10  // last packet with no sequence
	MsgTypeFlagNegativeSeq MsgTypeFlagBits = 0b11  // last packet with sequence < 0
	MsgTypeFlagWithEvent   MsgTypeFlagBits = 0b100 // Payload contains event number (int32)
)

// Values that a VersionBits variable can take.
const (
	Version1 VersionBits = (iota + 1) << 4
	Version2
	Version3
	Version4
)

// Values that a HeaderSizeBits variable can take.
const (
	HeaderSize4 HeaderSizeBits = iota + 1
	HeaderSize8
	HeaderSize12
	HeaderSize16
)

// Values that a SerializationBits variable can take.
const (
	SerializationRaw    SerializationBits = 0
	SerializationJSON   SerializationBits = 0b1 << 4
	SerializationThrift SerializationBits = 0b11 << 4
	SerializationCustom SerializationBits = 0b1111 << 4
)

// Values that a CompressionBits variable can take.
const (
	CompressionNone   CompressionBits = 0
	CompressionGzip   CompressionBits = 0b1
	CompressionCustom CompressionBits = 0b1111
)

var (
	msgTypeToBits = map[MsgType]uint8{
		MsgTypeFullClient:           0b1 << 4,
		MsgTypeAudioOnlyClient:      0b10 << 4,
		MsgTypeFullServer:           0b1001 << 4,
		MsgTypeAudioOnlyServer:      0b1011 << 4,
		MsgTypeFrontEndResultServer: 0b1100 << 4,
		MsgTypeError:                0b1111 << 4,
	}
	bitsToMsgType = make(map[uint8]MsgType, len(msgTypeToBits))
	
	serializations = map[SerializationBits]bool{
		SerializationRaw:    true,
		SerializationJSON:   true,
		SerializationThrift: true,
		SerializationCustom: true,
	}
	
	compressions = map[CompressionBits]bool{
		CompressionNone:   true,
		CompressionGzip:   true,
		CompressionCustom: true,
	}
)

func init() {
	// Construct inverse mapping of msgTypeToBits.
	for msgType, bits := range msgTypeToBits {
		bitsToMsgType[bits] = msgType
	}
}

// ContainsSequenceFunc defines the functional type that checks whether the
// MsgTypeFlagBits indicates the existence of a sequence number in serialized
// data. The background is that not all responses contain a sequence number,
// and whether a response contains one depends on the the message type specific
// flag bits. What makes it more complicated is that this dependency varies in
// each use case (eg, TTS protocol has its own dependency specification, more
// details at: https://bytedance.feishu.cn/docs/doccn8MD4cZHQuvobbtouWfUVsV).
type ContainsSequenceFunc func(MsgTypeFlagBits) bool

// CompressFunc defines the functional type that does the compression operation.
type CompressFunc func([]byte) ([]byte, error)

type readFunc func(*bytes.Buffer) error
type writeFunc func(*bytes.Buffer) error

// Unmarshal deserializes the binary `data` into a Message and also returns
// the Binary
func Unmarshal(data []byte, containsSequence ContainsSequenceFunc) (*Message, *BinaryProtocol, error) {
	var (
		buf      = bytes.NewBuffer(data)
		readSize int
	)
	
	versionSize, err := buf.ReadByte()
	if err != nil {
		return nil, nil, errNoVersionAndSize
	}
	readSize++
	
	prot := &BinaryProtocol{
		versionAndHeaderSize: versionSize,
		containsSequence:     containsSequence,
	}
	
	typeAndFlag, err := buf.ReadByte()
	if err != nil {
		return nil, nil, errNoTypeAndFlag
	}
	readSize++
	
	msg, err := NewMessageFromByte(typeAndFlag)
	if err != nil {
		return nil, nil, err
	}
	
	serializationCompression, err := buf.ReadByte()
	if err != nil {
		return nil, nil, errNoSerializationAndCompression
	}
	readSize++
	prot.serializationAndCompression = serializationCompression
	if _, ok := serializations[prot.Serialization()]; !ok {
		return nil, nil, fmt.Errorf("%w: %b", errInvalidSerialization, prot.Serialization())
	}
	if _, ok := compressions[prot.Compression()]; !ok {
		return nil, nil, fmt.Errorf("%w: %b", errInvalidCompression, prot.Compression())
	}
	
	// Read all the remaining zero-padding bytes in the header.
	if paddingSize := prot.HeaderSize() - readSize; paddingSize > 0 {
		if n, err := buf.Read(make([]byte, paddingSize)); err != nil || n < paddingSize {
			return nil, nil, fmt.Errorf("%w: %d", errNoEnoughHeaderBytes, n)
		}
	}
	
	readers, err := msg.readers(containsSequence)
	if err != nil {
		return nil, nil, err
	}
	for _, read := range readers {
		if err := read(buf); err != nil {
			return nil, nil, err
		}
	}
	
	if _, err := buf.ReadByte(); err != io.EOF {
		return nil, nil, errRedundantBytes
	}
	return msg, prot, nil
}

// Message defines the general message content type.
type Message struct {
	Type            MsgType
	typeAndFlagBits uint8
	
	Event     int32
	SessionID string
	ConnectID string
	Sequence  int32
	ErrorCode uint32
	// Raw payload (not Gzip compressed). BinaryMarshal will do the
	// compression for you.
	Payload []byte
}

// NewMessage returns a new Message instance of the given message type with the
// specific flag.
func NewMessage(msgType MsgType, typeFlag MsgTypeFlagBits) (*Message, error) {
	bits, ok := msgTypeToBits[msgType]
	if !ok {
		return nil, fmt.Errorf("invalid message type: %d", msgType)
	}
	return &Message{
		Type:            msgType,
		typeAndFlagBits: bits + uint8(typeFlag),
	}, nil
}

// NewMessageFromByte reads the byte as the message type and specific flag bits
// and composes a new Message instance from them.
func NewMessageFromByte(typeAndFlag byte) (*Message, error) {
	bits := typeAndFlag &^ 0b00001111
	msgType, ok := bitsToMsgType[bits]
	if !ok {
		return nil, fmt.Errorf("%w: %b", errInvalidMessageType, bits>>4)
	}
	return &Message{
		Type:            msgType,
		typeAndFlagBits: typeAndFlag,
	}, nil
}

// TypeFlag returns the message type specific flag.
func (m *Message) TypeFlag() MsgTypeFlagBits {
	return MsgTypeFlagBits(m.typeAndFlagBits &^ 0b11110000)
}

func (m *Message) writers(compress CompressFunc) (writers []writeFunc, _ error) {
	if compress != nil {
		payload, err := compress(m.Payload)
		if err != nil {
			return nil, fmt.Errorf("compress payload failed: %w", err)
		}
		m.Payload = payload
	}
	
	if containsSequence(m.TypeFlag()) {
		writers = append(writers, m.writeSequence)
	}
	
	if containsEvent(m.TypeFlag()) {
		writers = append(writers, m.writeEvent, m.writeSessionID)
	}
	
	writers = append(writers, m.writePayload)
	return writers, nil
}

func (m *Message) writeEvent(buf *bytes.Buffer) error {
	if err := binary.Write(buf, binary.BigEndian, m.Event); err != nil {
		return fmt.Errorf("write sequence number (%d): %w", m.Event, err)
	}
	return nil
}

func (m *Message) writeSessionID(buf *bytes.Buffer) error {
	switch m.Event {
	case 1, 2, 50, 51, 52: // StartConnection, FinishConnection, ConnectionStarted, ConnectionFailed, ConnectionFinished
		return nil
	}
	
	size := len(m.SessionID)
	if size > math.MaxUint32 {
		return fmt.Errorf("payload size (%d) exceeds max(uint32)", size)
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(size)); err != nil {
		return fmt.Errorf("write payload size (%d): %w", size, err)
	}
	buf.WriteString(m.SessionID)
	return nil
}

func (m *Message) writeSequence(buf *bytes.Buffer) error {
	if err := binary.Write(buf, binary.BigEndian, m.Sequence); err != nil {
		return fmt.Errorf("write sequence number (%d): %w", m.Sequence, err)
	}
	return nil
}

func (m *Message) writeErrorCode(buf *bytes.Buffer) error {
	if err := binary.Write(buf, binary.BigEndian, m.ErrorCode); err != nil {
		return fmt.Errorf("write error code (%d): %w", m.ErrorCode, err)
	}
	return nil
}

func (m *Message) writePayload(buf *bytes.Buffer) error {
	size := len(m.Payload)
	if size > math.MaxUint32 {
		return fmt.Errorf("payload size (%d) exceeds max(uint32)", size)
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(size)); err != nil {
		return fmt.Errorf("write payload size (%d): %w", size, err)
	}
	buf.Write(m.Payload)
	return nil
}

func (m *Message) readers(containsSequence ContainsSequenceFunc) (readers []readFunc, _ error) {
	
	switch m.Type {
	case MsgTypeFullClient, MsgTypeFullServer, MsgTypeFrontEndResultServer:
	
	case MsgTypeAudioOnlyClient:
		if containsSequence == nil || containsSequence(m.TypeFlag()) {
			readers = append(readers, m.readSequence)
		}
	
	case MsgTypeAudioOnlyServer:
		if containsSequence != nil && containsSequence(m.TypeFlag()) {
			readers = append(readers, m.readSequence)
		}
	
	case MsgTypeError:
		readers = append(readers, m.readErrorCode)
	
	default:
		return nil, fmt.Errorf("cannot deserialize message with invalid type: %d", m.Type)
	}
	
	if containsEvent(m.TypeFlag()) {
		readers = append(readers, m.readEvent, m.readSessionID, m.readConnectID)
	}
	
	readers = append(readers, m.readPayload)
	return readers, nil
}

func (m *Message) readEvent(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &m.Event); err != nil {
		return fmt.Errorf("%w: %v", errReadEvent, err)
	}
	return nil
}

func (m *Message) readSessionID(buf *bytes.Buffer) error {
	switch m.Event {
	case 1, 2, 50, 51, 52: //StartConnection, FinishConnection, ConnectionStarted, ConnectionFailed, ConnectionFinished
		return nil
	}
	
	var size uint32
	if err := binary.Read(buf, binary.BigEndian, &size); err != nil {
		return fmt.Errorf("%w: %v", errReadSessionIDSize, err)
	}
	
	if size > 0 {
		m.SessionID = string(buf.Next(int(size)))
	}
	return nil
}

func (m *Message) readConnectID(buf *bytes.Buffer) error {
	switch m.Event {
	case 50, 51, 52: // ConnectionStarted, event.Type_ConnectionFailed, ConnectionFinished
	default:
		return nil
	}
	
	var size uint32
	if err := binary.Read(buf, binary.BigEndian, &size); err != nil {
		return fmt.Errorf("%w: %v", errReadConnectIDSize, err)
	}
	
	if size > 0 {
		m.ConnectID = string(buf.Next(int(size)))
	}
	return nil
}

func (m *Message) readSequence(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &m.Sequence); err != nil {
		return fmt.Errorf("%w: %v", errReadSequence, err)
	}
	return nil
}

func (m *Message) readErrorCode(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &m.ErrorCode); err != nil {
		return fmt.Errorf("%w: %v", errReadErrorCode, err)
	}
	return nil
}

func (m *Message) readPayload(buf *bytes.Buffer) error {
	var size uint32
	if err := binary.Read(buf, binary.BigEndian, &size); err != nil {
		return fmt.Errorf("%w: %v", errReadPayloadSize, err)
	}
	
	if size > 0 {
		m.Payload = buf.Next(int(size))
	}
	if m.Type == MsgTypeFullClient || m.Type == MsgTypeFullServer || m.Type == MsgTypeError {
	}
	return nil
}

// ContainsSequence reports whether a message type specific flag indicates
// messages with this kind of flag contain a sequence number in its serialized
// value. This determiner function should be used for common binary 
func ContainsSequence(bits MsgTypeFlagBits) bool {
	return bits&MsgTypeFlagPositiveSeq == MsgTypeFlagPositiveSeq || bits&MsgTypeFlagNegativeSeq == MsgTypeFlagNegativeSeq
}

func containsEvent(bits MsgTypeFlagBits) bool {
	return bits&MsgTypeFlagWithEvent == MsgTypeFlagWithEvent
}

func containsSequence(bits MsgTypeFlagBits) bool {
	return bits&MsgTypeFlagPositiveSeq == MsgTypeFlagPositiveSeq || bits&MsgTypeFlagNegativeSeq == MsgTypeFlagNegativeSeq
}

// BinaryProtocol implements the binary protocol serialization and deserialization
// used in Lab-Speech MDD, TTS, ASR, etc. services. For more details, read:
// https://bytedance.feishu.cn/docs/doccnT0t71J4LCQCS0cnB4Eca8D
type BinaryProtocol struct {
	versionAndHeaderSize        uint8
	serializationAndCompression uint8
	
	containsSequence ContainsSequenceFunc
	compress         CompressFunc
}

// NewBinaryProtocol returns a new BinaryProtocol instance.
func NewBinaryProtocol() *BinaryProtocol {
	b := new(BinaryProtocol)
	b.SetVersion(Version1)
	b.SetHeaderSize(HeaderSize4)
	b.SetSerialization(SerializationJSON)
	b.SetCompression(CompressionNone, nil)
	b.containsSequence = ContainsSequence
	return b
}

// Clone returns a clone of current BinaryProtocol
func (p *BinaryProtocol) Clone() *BinaryProtocol {
	clonedBinaryProtocal := new(BinaryProtocol)
	clonedBinaryProtocal.versionAndHeaderSize = p.versionAndHeaderSize
	clonedBinaryProtocal.serializationAndCompression = p.serializationAndCompression
	clonedBinaryProtocal.containsSequence = p.containsSequence
	clonedBinaryProtocal.compress = p.compress
	return clonedBinaryProtocal
}

// SetVersion sets the protocol version.
func (p *BinaryProtocol) SetVersion(v VersionBits) {
	// Clear the higher 4 bits in `p.versionAndHeaderSize` and reset them to `v`.
	p.versionAndHeaderSize = (p.versionAndHeaderSize &^ 0b11110000) + uint8(v)
}

// Version returns the integral version value.
func (p *BinaryProtocol) Version() int {
	return int(p.versionAndHeaderSize >> 4)
}

// SetHeaderSize sets the protocol header size.
func (p *BinaryProtocol) SetHeaderSize(s HeaderSizeBits) {
	// Clear the lower 4 bits in `p.versionAndHeaderSize` and reset them to `s`.
	p.versionAndHeaderSize = (p.versionAndHeaderSize &^ 0b00001111) + uint8(s)
}

// HeaderSize returns the protocol header size.
func (p *BinaryProtocol) HeaderSize() int {
	return 4 * int(p.versionAndHeaderSize&^0b11110000)
}

// SetSerialization sets the serialization method.
func (p *BinaryProtocol) SetSerialization(s SerializationBits) {
	// Clear the higher 4 bits in `p.serializationAndCompression` and reset them to `s`.
	p.serializationAndCompression = (p.serializationAndCompression &^ 0b11110000) + uint8(s)
}

// Serialization returns the bits value of protocol serialization method.
func (p *BinaryProtocol) Serialization() SerializationBits {
	return SerializationBits(p.serializationAndCompression &^ 0b00001111)
}

// SetCompression sets the compression method.
func (p *BinaryProtocol) SetCompression(c CompressionBits, f CompressFunc) {
	// Clear the lower 4 bits in `p.serializationAndCompression` and reset them to `c`.
	p.serializationAndCompression = (p.serializationAndCompression &^ 0b00001111) + uint8(c)
	p.compress = f
}

// Compression returns the bits value of protocol compression method.
func (p *BinaryProtocol) Compression() CompressionBits {
	return CompressionBits(p.serializationAndCompression &^ 0b11110000)
}

// Marshal serializes the message to a sequence of binary data.
func (p *BinaryProtocol) Marshal(msg *Message) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := p.writeHeader(buf, msg); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}
	
	writers, err := msg.writers(p.compress)
	if err != nil {
		return nil, err
	}
	for _, write := range writers {
		if err := write(buf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (p *BinaryProtocol) writeHeader(buf *bytes.Buffer, msg *Message) error {
	return binary.Write(buf, binary.BigEndian, p.header(msg))
}

func (p *BinaryProtocol) header(msg *Message) []byte {
	header := []uint8{
		p.versionAndHeaderSize,
		msg.typeAndFlagBits,
		p.serializationAndCompression,
	}
	if padding := p.HeaderSize() - len(header); padding > 0 {
		header = append(header, make([]uint8, padding)...)
	}
	return header
}

type StartSessionPayload struct {
	ASR    ASRPayload    `json:"asr"`
	TTS    TTSPayload    `json:"tts"`
	Dialog DialogPayload `json:"dialog"`
}

type SayHelloPayload struct {
	Content string `json:"content"`
}

type ChatTTSTextPayload struct {
	Start   bool   `json:"start"`
	End     bool   `json:"end"`
	Content string `json:"content"`
}

type ASRPayload struct {
	Extra map[string]interface{} `json:"extra"`
}

type TTSPayload struct {
	Speaker     string      `json:"speaker"`
	AudioConfig AudioConfig `json:"audio_config"`
}

type AudioConfig struct {
	Channel    int    `json:"channel"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
}

type DialogPayload struct {
	DialogID      string                 `json:"dialog_id"`
	BotName       string                 `json:"bot_name"`
	SystemRole    string                 `json:"system_role"`
	SpeakingStyle string                 `json:"speaking_style"`
	Location      *LocationInfo          `json:"location,omitempty"`
	Extra         map[string]interface{} `json:"extra"`
}

type LocationInfo struct {
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Province    string  `json:"province"`
	District    string  `json:"district"`
	Town        string  `json:"town"`
	CountryCode string  `json:"country_code"`
	Address     string  `json:"address"`
}

type ChatTextQueryPayload struct {
	Content string `json:"content"`
}

type UsageResponse struct {
	Usage *Usage `json:"usage"`
}

type Usage struct {
	InputTextTokens   int `json:"input_text_tokens"`
	InputAudioTokens  int `json:"input_audio_tokens"`
	CachedTextTokens  int `json:"cached_text_tokens"`
	CachedAudioTokens int `json:"cached_audio_tokens"`
	OutputTextTokens  int `json:"output_text_tokens"`
	OutputAudioTokens int `json:"output_audio_tokens"`
}

func StartConnection(conn *websocket.Conn) error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create StartSession request message: %w", err)
	}
	msg.Event = 1
	msg.Payload = []byte("{}")
	
	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal StartConnection request message: %w", err)
	}
	
	if err := sendRequest(conn, frame); err != nil {
		return fmt.Errorf("send StartConnection request: %w", err)
	}
	
	// Read ConnectionStarted message.
	mt, frame, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read ConnectionStarted response: %w", err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return fmt.Errorf("unexpected Websocket message type: %d", mt)
	}
	
	msg, _, err = Unmarshal(frame, containsSequence)
	if err != nil {
		return fmt.Errorf("unmarshal ConnectionStarted response message: %w", err)
	}
	if msg.Type != MsgTypeFullServer {
		return fmt.Errorf("unexpected ConnectionStarted message type: %s", msg.Type)
	}
	if msg.Event != 50 {
		return fmt.Errorf("unexpected response event (%d) for StartConnection request", msg.Event)
	}
	
	return nil
}

func StartSession(conn *websocket.Conn, sessionID string, req *StartSessionPayload) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal StartSession request payload: %w", err)
	}
	
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create StartSession request message: %w", err)
	}
	msg.Event = 100
	msg.SessionID = sessionID
	msg.Payload = payload
	
	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal StartSession request message: %w", err)
	}
	
	if err := sendRequest(conn, frame); err != nil {
		return fmt.Errorf("send StartSession request: %w", err)
	}
	
	// Read SessionStarted message.
	mt, frame, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read SessionStarted response: %w", err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return fmt.Errorf("unexpected Websocket message type: %d", mt)
	}
	
	// Validate SessionStarted message.
	msg, _, err = Unmarshal(frame, containsSequence)
	if err != nil {
		return fmt.Errorf("unmarshal SessionStarted response message: %w", err)
	}
	if msg.Type != MsgTypeFullServer {
		return fmt.Errorf("unexpected SessionStarted message type: %s", msg.Type)
	}
	if msg.Event != 150 {
		return fmt.Errorf("unexpected response event (%d) for StartSession request", msg.Event)
	}
	var jsonData map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &jsonData); err != nil {
		return fmt.Errorf("unmarshal SessionStarted response payload: %w", err)
	}
	return nil
}

func SendAudio(c *websocket.Conn, sessionID string, data []byte) error {
	protocol.SetSerialization(SerializationRaw)
	msg, err := NewMessage(MsgTypeAudioOnlyClient, MsgTypeFlagWithEvent)
	if err != nil {
		return err
	}
	msg.Event = 200
	msg.SessionID = sessionID
	msg.Payload = data
	
	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal TaskRequest message: %w", err)
	}
	if err := sendRequest(c, frame); err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	
	time.Sleep(20 * time.Millisecond)
	return nil
}

func SendAudioFromWav(ctx context.Context, c *websocket.Conn, sessionID string, content []byte) error {
	
	protocol.SetSerialization(SerializationRaw)
	
	// 16000Hz, 16bit, 1 channel
	sleepDuration := 20 * time.Millisecond
	bufferSize := 640
	curPos := 0
	for curPos < len(content) {
		if curPos+bufferSize >= len(content) {
			bufferSize = len(content) - curPos
		}
		msg, err := NewMessage(MsgTypeAudioOnlyClient, MsgTypeFlagWithEvent)
		if err != nil {
			return fmt.Errorf("create Task request message: %w", err)
		}
		msg.Event = 200
		msg.SessionID = sessionID
		msg.Payload = content[curPos : curPos+bufferSize]
		
		frame, err := protocol.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal TaskRequest message: %w", err)
		}
		if err = sendRequest(c, frame); err != nil {
			return fmt.Errorf("send TaskRequest request: %w", err)
		}
		
		curPos += bufferSize
		// 非最后一片时，休眠对应时长（模拟实时输入）
		if curPos < len(content) {
			// 休眠期间也监听上下文取消（避免长时间阻塞）
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled during sleep: %w", ctx.Err())
			case <-time.After(sleepDuration):
			}
		}
	}
	return nil
}

const (
	sampleRate    = 24000
	bufferSeconds = 100 // 最多缓冲100秒数据
)

var (
	bufferLock sync.Mutex
	buffer     = make([]float32, 0, sampleRate*bufferSeconds)
	s16Buffer  = make([]int16, 0, sampleRate*bufferSeconds)
)

/**
 * 结合api接入文档对二进制协议进行理解，上下行统一理解
 * - header(4bytes)
 *     - (4bits)version(v1) + (4bits)header_size
 *     - (4bits)messageType + (4bits)messageTypeFlags
 *         -- 0001	CompleteClient  | -- 0001 optional has sequence
 *         -- 0010	AudioOnlyClient | -- 0100 optional has event
 *         -- 1001 CompleteServer   | -- 1111 optional has error code
 *         -- 1011 AudioOnlyServer  | --
 *     - (4bits)payloadFormat + (4bits)compression
 *     - (8bits) reserve
 * - payload
 *     - [optional 4 bytes] event
 *     - [optional] session ID
 *       -- (4 bytes)session ID len
 *       -- session ID data
 *     - (4 bytes)data len
 *     - data
 */
func ReceiveMessage(conn *websocket.Conn) (*Message, error) {
	mt, frame, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return nil, fmt.Errorf("unexpected Websocket message type: %d", mt)
	}
	
	msg, _, err := Unmarshal(frame, ContainsSequence)
	if err != nil {
		if len(frame) > 500 {
			frame = frame[:500]
		}
		return nil, fmt.Errorf("unmarshal response message: %w", err)
	}
	return msg, nil
}

func HandleIncomingAudio(data []byte) {
	sampleCount := len(data) / 2
	samples := make([]int16, sampleCount)
	for i := 0; i < sampleCount; i++ {
		bits := binary.LittleEndian.Uint16(data[i*2 : (i+1)*2])
		samples[i] = int16(bits)
	}
	// 将音频加载到缓冲区
	bufferLock.Lock()
	defer bufferLock.Unlock()
	s16Buffer = append(s16Buffer, samples...)
	if len(buffer) > sampleRate*bufferSeconds {
		s16Buffer = s16Buffer[len(s16Buffer)-(sampleRate*bufferSeconds):]
	}
}

func SendSilenceAudio(c *websocket.Conn, sessionID string) error {
	// 16000Hz, 16bit, 1 channel, 10ms silence
	silence := make([]byte, 320) // 160 samples * 2 bytes/sample = 320 bytes
	protocol.SetSerialization(SerializationRaw)
	msg, err := NewMessage(MsgTypeAudioOnlyClient, MsgTypeFlagWithEvent)
	if err != nil {
		return err
	}
	msg.Event = 200
	msg.SessionID = sessionID
	msg.Payload = silence
	frame, err := protocol.Marshal(msg)
	if err != nil {
		return err
	}
	
	if err := sendRequest(c, frame); err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	
	time.Sleep(10 * time.Millisecond)
	return nil
}

func sendRequest(conn *websocket.Conn, frame []byte) error {
	wsWriteLock.Lock()
	defer wsWriteLock.Unlock()
	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send SayHello request: %w", err)
	}
	return nil
}

func GetDialogUsage(data []byte) *UsageResponse {
	usage := &UsageResponse{}
	err := json.Unmarshal(data, &usage)
	if err != nil {
		logger.Error("unmarshal usage response fail", "err", err)
		return nil
	}
	
	return usage
}
