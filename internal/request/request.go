package request

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	parserState parserState
	bodyRead    int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState int

const (
	requestStateInitialized parserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

var (
	validHttpVersions = []string{"1.1", "2.3", "3.1", "3.2"}
	bufferSize        = 8
)

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.parserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		totalBytesParsed += n
		if n == 0 {
			break
		}

	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.parserState {
	case requestStateInitialized:
		reqLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.parserState = requestStateParsingHeaders
		return n, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.parserState = requestStateParsingBody
		}

		return n, nil
	case requestStateParsingBody:
		contentLength := r.Headers.Get("content-length")
		if contentLength == "" {
			r.parserState = requestStateDone
			return len(data), nil
		}

		cl, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data...)
		r.bodyRead += len(data)

		if r.bodyRead > cl {
			return 0, errors.New("Content-Length too large")
		}

		if r.bodyRead == cl {
			r.parserState = requestStateDone
		}

		return len(data), nil
	case requestStateDone:
		return 0, errors.New("error: trying to read data in requestStateDone state")
	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIdx := 0
	req := &Request{
		parserState: requestStateInitialized,
		Headers:     headers.NewHeaders(),
		Body:        make([]byte, 0),
	}

	for req.parserState != requestStateDone {
		if readToIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// read into the buffer
		numBytesRead, err := reader.Read(buf[readToIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.parserState != requestStateDone {
					return nil, errors.New("incomplete request")
				}
				break
			}
			return nil, err
		}

		// Parse the buffer
		readToIdx += numBytesRead
		numBytesParsed, err := req.parse(buf[:readToIdx])
		if err != nil {
			return nil, err
		}

		// move the buffer window such that we don't parse the same data
		copy(buf, buf[numBytesParsed:])
		readToIdx -= numBytesParsed
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}

	// Only get the first line, until the CRLF
	requestLineText := string(data[:idx])
	requestLine, err := parseRequestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	// We add 2 for the CRLF at the end of the line
	return requestLine, idx + 2, err
}

func parseRequestLineFromString(requestLine string) (*RequestLine, error) {
	headerParts := strings.Split(requestLine, " ")
	if len(headerParts) != 3 {
		return nil, errors.New("invalid request line")
	}
	method := headerParts[0]
	if method != strings.ToUpper(method) {
		return nil, errors.New("invalid method")
	}

	requestTarget := headerParts[1]

	versionParts := strings.Split(headerParts[2], "/")
	if len(versionParts) != 2 {
		return nil, errors.New("malformed version request line")
	}

	httpVersionPart := versionParts[0]
	if !strings.EqualFold("HTTP", httpVersionPart) {
		return nil, errors.New("malformed version")
	}

	version := versionParts[1]
	if !slices.Contains(validHttpVersions, versionParts[1]) {
		return nil, errors.New("invalid HTTP version")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
