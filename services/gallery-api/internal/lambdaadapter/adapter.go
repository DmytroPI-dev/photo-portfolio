package lambdaadapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// New adapts API Gateway HTTP API (payload v2) events to the same net/http
// handler used locally. Keeping this boundary explicit avoids a framework lock
// and lets handler tests exercise the business-facing HTTP contract directly.
func New(handler http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(context context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		request, err := requestFromEvent(context, event)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, fmt.Errorf("build HTTP request: %w", err)
		}

		response := newResponseRecorder()
		handler.ServeHTTP(response, request)

		return response.toAPIGatewayResponse(), nil
	}
}

func requestFromEvent(context context.Context, event events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	method := event.RequestContext.HTTP.Method
	if method == "" {
		method = http.MethodGet
	}

	path := event.RawPath
	if path == "" {
		path = event.RequestContext.HTTP.Path
	}
	if path == "" {
		path = "/"
	}

	var body []byte
	if event.Body != "" {
		body = []byte(event.Body)
		if event.IsBase64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(event.Body)
			if err != nil {
				return nil, fmt.Errorf("decode base64 request body: %w", err)
			}
			body = decoded
		}
	}

	request, err := http.NewRequestWithContext(
		context,
		method,
		"https://lambda.local"+path+querySuffix(event.RawQueryString),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	for key, value := range event.Headers {
		request.Header.Set(key, value)
	}
	if len(event.Cookies) > 0 {
		request.Header.Set("Cookie", strings.Join(event.Cookies, "; "))
	}

	return request, nil
}

func querySuffix(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	return "?" + rawQuery
}

type responseRecorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header)}
}

func (response *responseRecorder) Header() http.Header {
	return response.header
}

func (response *responseRecorder) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}

func (response *responseRecorder) Write(body []byte) (int, error) {
	if response.status == 0 {
		response.status = http.StatusOK
	}
	return response.body.Write(body)
}

func (response *responseRecorder) toAPIGatewayResponse() events.APIGatewayV2HTTPResponse {
	headers := make(map[string]string, len(response.header))
	cookies := make([]string, 0)
	for key, values := range response.header {
		if strings.EqualFold(key, "Set-Cookie") {
			cookies = append(cookies, values...)
			continue
		}
		headers[key] = strings.Join(values, ",")
	}

	status := response.status
	if status == 0 {
		status = http.StatusOK
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    headers,
		Cookies:    cookies,
		Body:       response.body.String(),
	}
}
