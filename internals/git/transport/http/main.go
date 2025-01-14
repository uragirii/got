package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/uragirii/got/internals"
)

var _Service = []byte("# service=git-upload-pack\n")

var _Version = []byte("version 2\n")

var ErrMalformedResponse = errors.New("malformed response")

var _FlushPacketBytes = []byte("0000")
var _DelimPacketBytes = []byte("0001")

var _FlushPacket = []byte("flush")
var _DelimPacket = []byte("delim")

var _CapabilityAdvResponsePrefix = [][]byte{
	_Service,
	_FlushPacket,
	_Version,
}

type Capability struct {
	LsRefs       string
	Fetch        []string
	ServerOption string
	ObjectFormat string
}

var _CapabilityParserMap = map[string]func(args string, c *Capability) error{
	"agent": func(args string, c *Capability) error {
		return nil
	},
	"ls-refs": func(args string, c *Capability) error {
		c.LsRefs = args

		return nil
	},
	"fetch": func(args string, c *Capability) error {
		c.Fetch = strings.Split(args, " ")

		return nil
	},
	"server-option": func(args string, c *Capability) error {
		c.ServerOption = args

		return nil
	},
	"object-format": func(args string, c *Capability) error {
		c.ObjectFormat = args

		return nil
	},
}

func CapabilityAdvertisement(gitUrl string) (*Capability, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/info/refs?service=git-upload-pack", gitUrl), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Git-Protocol", "version=2")
	req.Header.Add("User-Agent", internals.Agent())

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return parseCapabiltyResponse(resp.Body)
}

func parseCapabiltyResponse(r io.Reader) (*Capability, error) {
	readOneLine := func() ([]byte, error) {
		lineLen := make([]byte, 4)

		_, err := r.Read(lineLen)

		if err != nil {
			return []byte{}, err
		}

		/**
		* i want to know better way to handle this behaviour in generic way
		* other option is throw errors for this, but then reader code would be
		* cluttered with error checking everytime, idk
		 */
		if bytes.Equal(_FlushPacketBytes, lineLen) {
			return _FlushPacket, nil
		}

		if bytes.Equal(_DelimPacketBytes, lineLen) {
			return _DelimPacket, nil
		}

		len, err := strconv.ParseInt(string(lineLen), 16, 64)

		if err != nil {
			return []byte{}, err
		}

		data := make([]byte, len-4)

		_, err = r.Read(data)

		if err != nil {
			return []byte{}, err
		}

		return data, nil
	}

	for _, prefix := range _CapabilityAdvResponsePrefix {
		line, err := readOneLine()

		if err != nil {
			return nil, err
		}

		if !bytes.Equal(prefix, line) {
			return nil, ErrMalformedResponse
		}
	}

	var capability Capability

	for {
		line, err := readOneLine()

		if err != nil {
			return nil, err
		}

		if bytes.Equal(_FlushPacket, line) {
			return &capability, nil
		}

		capabiltyLine := strings.Split(strings.TrimRight(string(line), "\n"), "=")

		key := capabiltyLine[0]
		var value string

		if len(capabiltyLine) > 1 {
			value = capabiltyLine[1]
		}

		parser, ok := _CapabilityParserMap[key]

		if !ok {
			continue
		}

		parser(value, &capability)
	}
}
