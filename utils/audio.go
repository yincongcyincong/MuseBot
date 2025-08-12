package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"github.com/yincongcyincong/MuseBot/logger"
)

type ProtocolVersion byte
type MessageType byte
type MessageTypeSpecificFlags byte
type SerializationType byte
type CompressionType byte

const (
	SuccessCode = 1000
	
	PROTOCOL_VERSION    = ProtocolVersion(0b0001)
	DEFAULT_HEADER_SIZE = 0b0001
	
	PROTOCOL_VERSION_BITS            = 4
	HEADER_BITS                      = 4
	MESSAGE_TYPE_BITS                = 4
	MESSAGE_TYPE_SPECIFIC_FLAGS_BITS = 4
	MESSAGE_SERIALIZATION_BITS       = 4
	MESSAGE_COMPRESSION_BITS         = 4
	RESERVED_BITS                    = 8
	
	// Message Type:
	CLIENT_FULL_REQUEST       = MessageType(0b0001)
	CLIENT_AUDIO_ONLY_REQUEST = MessageType(0b0010)
	SERVER_FULL_RESPONSE      = MessageType(0b1001)
	SERVER_ACK                = MessageType(0b1011)
	SERVER_ERROR_RESPONSE     = MessageType(0b1111)
	
	// Message Type Specific Flags
	NO_SEQUENCE    = MessageTypeSpecificFlags(0b0000) // no check sequence
	POS_SEQUENCE   = MessageTypeSpecificFlags(0b0001)
	NEG_SEQUENCE   = MessageTypeSpecificFlags(0b0010)
	NEG_SEQUENCE_1 = MessageTypeSpecificFlags(0b0011)
	
	// Message Serialization
	NO_SERIALIZATION = SerializationType(0b0000)
	JSON             = SerializationType(0b0001)
	THRIFT           = SerializationType(0b0011)
	CUSTOM_TYPE      = SerializationType(0b1111)
	
	// Message Compression
	NO_COMPRESSION     = CompressionType(0b0000)
	GZIP               = CompressionType(0b0001)
	CUSTOM_COMPRESSION = CompressionType(0b1111)
)

// version: b0001 (4 bits)
// header size: b0001 (4 bits)
// message type: b0001 (Full client request) (4bits)
// message type specific flags: b0000 (none) (4bits)
// message serialization method: b0001 (JSON) (4 bits)
// message compression: b0001 (gzip) (4bits)
// reserved data: 0x00 (1 byte)
var DefaultFullClientWsHeader = []byte{0x11, 0x10, 0x11, 0x00}
var DefaultAudioOnlyWsHeader = []byte{0x11, 0x20, 0x11, 0x00}
var DefaultLastAudioWsHeader = []byte{0x11, 0x22, 0x11, 0x00}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func gzipDecompress(input []byte) []byte {
	b := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(b)
	out, _ := ioutil.ReadAll(r)
	r.Close()
	return out
}

type AsrResponse struct {
	Reqid    string   `json:"reqid"`
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Sequence int      `json:"sequence"`
	Results  []Result `json:"result,omitempty"`
}

type Result struct {
	// required
	Text       string `json:"text"`
	Confidence int    `json:"confidence"`
	// if show_language == true
	Language string `json:"language,omitempty"`
	// if show_utterances == true
	Utterances []Utterance `json:"utterances,omitempty"`
}

type Utterance struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Definite  bool   `json:"definite"`
	Words     []Word `json:"words"`
	// if show_language = true
	Language string `json:"language"`
}

type Word struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Pronounce string `json:"pronounce"`
	// in docs example - blank_time
	BlankDuration int `json:"blank_duration"`
}

type WsHeader struct {
	ProtocolVersion          ProtocolVersion
	DefaultHeaderSize        int
	MessageType              MessageType
	MessageTypeSpecificFlags MessageTypeSpecificFlags
	SerializationType        SerializationType
	CompressionType          CompressionType
}

type RequestAsr interface {
	requestAsr(audio_data []byte)
}

type AsrClient struct {
	Appid    string
	Token    string
	Cluster  string
	Workflow string
	Format   string
	Codec    string
	SegSize  int
	Url      string
}

func BuildAsrClient() AsrClient {
	client := AsrClient{}
	client.Workflow = "audio_in,resample,partition,vad,fe,decode"
	client.SegSize = 160000 // 5s in 16000 sample rate
	client.Format = "ogg"   // default wav audio
	client.Codec = "opus"   // default raw codec
	return client
}

func (client *AsrClient) RequestAsr(audioData []byte) (AsrResponse, error) {
	// set token header
	var tokenHeader = http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", client.Token)}}
	c, _, err := websocket.DefaultDialer.Dial("wss://openspeech.bytedance.com/api/v2/asr", tokenHeader)
	if err != nil {
		logger.Warn("dial fail", "err", err)
		return AsrResponse{}, err
	}
	defer c.Close()
	client.Format = DetectAudioFormat(audioData)
	
	// 1. send full client request
	req := client.ConstructRequest()
	payload := gzipCompress(req)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
	
	fullClientMsg := make([]byte, len(DefaultFullClientWsHeader))
	copy(fullClientMsg, DefaultFullClientWsHeader)
	fullClientMsg = append(fullClientMsg, payloadSizeArr...)
	fullClientMsg = append(fullClientMsg, payload...)
	c.WriteMessage(websocket.BinaryMessage, fullClientMsg)
	_, msg, err := c.ReadMessage()
	if err != nil {
		logger.Warn("fail to read message fail, err:", err.Error())
		return AsrResponse{}, err
	}
	asrResponse, err := client.parseResponse(msg)
	if err != nil {
		logger.Warn("fail to parse response ", err.Error())
		return AsrResponse{}, err
	}
	
	// 3. send segment audio request
	for sentSize := 0; sentSize < len(audioData); sentSize += client.SegSize {
		lastAudio := false
		if sentSize+client.SegSize >= len(audioData) {
			lastAudio = true
		}
		dataSlice := make([]byte, 0)
		audioMsg := make([]byte, len(DefaultAudioOnlyWsHeader))
		if !lastAudio {
			dataSlice = audioData[sentSize : sentSize+client.SegSize]
			copy(audioMsg, DefaultAudioOnlyWsHeader)
		} else {
			dataSlice = audioData[sentSize:]
			copy(audioMsg, DefaultLastAudioWsHeader)
		}
		payload = gzipCompress(dataSlice)
		payloadSize := len(payload)
		payloadSizeArr := make([]byte, 4)
		binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
		audioMsg = append(audioMsg, payloadSizeArr...)
		audioMsg = append(audioMsg, payload...)
		c.WriteMessage(websocket.BinaryMessage, audioMsg)
		_, msg, err := c.ReadMessage()
		if err != nil {
			logger.Error("fail to read message fail", "err", err.Error())
			return AsrResponse{}, err
		}
		asrResponse, err = client.parseResponse(msg)
		if err != nil {
			logger.Error("fail to parse response ", "err", err.Error())
			return AsrResponse{}, err
		}
	}
	return asrResponse, nil
}

func (client *AsrClient) ConstructRequest() []byte {
	reqid := uuid.NewV4().String()
	req := make(map[string]map[string]interface{})
	req["app"] = make(map[string]interface{})
	req["app"]["appid"] = client.Appid
	req["app"]["cluster"] = client.Cluster
	req["app"]["token"] = client.Token
	req["user"] = make(map[string]interface{})
	req["user"]["uid"] = "uid"
	req["request"] = make(map[string]interface{})
	req["request"]["reqid"] = reqid
	req["request"]["nbest"] = 1
	req["request"]["workflow"] = client.Workflow
	req["request"]["result_type"] = "full"
	req["request"]["sequence"] = 1
	req["audio"] = make(map[string]interface{})
	req["audio"]["format"] = client.Format
	req["audio"]["codec"] = client.Codec
	reqStr, _ := json.Marshal(req)
	return reqStr
}

func (client *AsrClient) parseResponse(msg []byte) (AsrResponse, error) {
	//protocol_version := msg[0] >> 4
	headerSize := msg[0] & 0x0f
	messageType := msg[1] >> 4
	//message_type_specific_flags := msg[1] & 0x0f
	serializationMethod := msg[2] >> 4
	messageCompression := msg[2] & 0x0f
	//reserved := msg[3]
	//header_extensions := msg[4:header_size * 4]
	payload := msg[headerSize*4:]
	payloadMsg := make([]byte, 0)
	payloadSize := 0
	//print('message type: {}'.format(message_type))
	
	if messageType == byte(SERVER_FULL_RESPONSE) {
		payloadSize = int(int32(binary.BigEndian.Uint32(payload[0:4])))
		payloadMsg = payload[4:]
	} else if messageType == byte(SERVER_ACK) {
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		if len(payload) >= 8 {
			payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
			payloadMsg = payload[8:]
		}
		logger.Info("SERVER_ACK", "seq", seq)
	} else if messageType == byte(SERVER_ERROR_RESPONSE) {
		code := int32(binary.BigEndian.Uint32(payload[:4]))
		payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payloadMsg = payload[8:]
		logger.Info("SERVER_ERROR_RESPONE", "code", code)
		return AsrResponse{}, errors.New(string(payloadMsg))
	}
	if payloadSize == 0 {
		return AsrResponse{}, errors.New("payload size if 0")
	}
	if messageCompression == byte(GZIP) {
		payloadMsg = gzipDecompress(payloadMsg)
	}
	
	var asrResponse = AsrResponse{}
	if serializationMethod == byte(JSON) {
		err := json.Unmarshal(payloadMsg, &asrResponse)
		if err != nil {
			logger.Info("fail to unmarshal response ", "err", err.Error())
			return AsrResponse{}, err
		}
	}
	return asrResponse, nil
}
