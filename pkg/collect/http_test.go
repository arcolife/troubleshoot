package collect

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	"github.com/stretchr/testify/assert"
)

// func TestDoRequestGET(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		rw.WriteHeader(http.StatusOK)
// 		rw.Write([]byte("Hello, World!"))
// 	}))
// 	defer server.Close()

// 	response, err := doRequest("GET", server.URL, nil, "", false, 5*time.Second)

// 	if err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	if response.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
// 	}
// }

// func TestDoRequestPOST(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		rw.WriteHeader(http.StatusOK)
// 		rw.Write([]byte("Hello, World!"))
// 	}))
// 	defer server.Close()

// 	response, err := doRequest("POST", server.URL, nil, "", false, 5*time.Second)

// 	if err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	if response.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
// 	}
// }

// func TestDoRequestPUT(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		rw.WriteHeader(http.StatusOK)
// 		rw.Write([]byte("Hello, World!"))
// 	}))
// 	defer server.Close()

// 	response, err := doRequest("PUT", server.URL, nil, "", false, 5*time.Second)

// 	if err != nil {
// 		t.Fatalf("Unexpected error: %v", err)
// 	}

// 	if response.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
// 	}
// }

func JSONIgnoreFieldsEqual(t *testing.T, expected, actual string, ignoreFields []string) {
	var expectedMap map[string]interface{}
	var actualMap map[string]interface{}

	err := json.Unmarshal([]byte(expected), &expectedMap)
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(actual), &actualMap)
	assert.NoError(t, err)

	// Remove ignored fields from both maps
	for _, field := range ignoreFields {
		delete(expectedMap, field)
		delete(actualMap, field)
	}

	// Compare the modified maps
	assert.Equal(t, expectedMap, actualMap)
}

type Headers struct {
	ContentLength string `json:"Content-Length"`
	ContentType   string `json:"Content-Type"`
	Date          string `json:"Date",omitempty`
}

type Response struct {
	Status  int     `json:"status"`
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

type ResponseData struct {
	Response Response `json:"response"`
}

type ErrorResponse struct {
	Error HTTPError `json:"error"`
}

// ToJSON converts the ResponseData struct to JSON bytes
func (r ResponseData) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func TestCollectHTTP_Collect(t *testing.T) {

	// ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 	rw.WriteHeader(http.StatusOK)
	// 	rw.Write([]byte("Hello, World!"))
	// }))

	// ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 	time.Sleep(3 * time.Second) // Simulate a slow server
	// }))

	// ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 	switch req.Method {
	// 	case http.MethodGet:
	// 		rw.WriteHeader(http.StatusOK)
	// 		rw.Write([]byte("Hello, GET!"))
	// 	case http.MethodPost:
	// 		rw.WriteHeader(http.StatusOK)
	// 		rw.Write([]byte("Hello, POST!"))
	// 	case http.MethodPut:
	// 		rw.WriteHeader(http.StatusOK)
	// 		rw.Write([]byte("Hello, PUT!"))
	// 	default:
	// 		rw.WriteHeader(http.StatusMethodNotAllowed)
	// 		rw.Write([]byte("Method not allowed"))
	// 	}
	// }))

	// defer ts.Close()
	// url := ts.URL

	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{\"status\": \"healthy\"}"))
	})
	mux.HandleFunc("/post", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Hello, POST!"))
	})
	mux.HandleFunc("/put", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Hello, PUT!"))
	})

	sample_get_response := &ResponseData{
		Response: Response{
			Status: 200,
			Body:   "{\"status\": \"healthy\"}",
			Headers: Headers{
				ContentLength: "21",
				ContentType:   "application/json; charset=utf-8",
			},
		},
	}
	sample_get_bytes, _ := sample_get_response.ToJSON()

	sample_post_response := &ResponseData{
		Response: Response{
			Status: 200,
			Body:   "Hello, POST!",
			Headers: Headers{
				ContentLength: "12",
				ContentType:   "text/plain; charset=utf-8",
			},
		},
	}
	sample_post_bytes, _ := sample_post_response.ToJSON()

	sample_put_response := &ResponseData{
		Response: Response{
			Status: 200,
			Body:   "Hello, PUT!",
			Headers: Headers{
				ContentLength: "11",
				ContentType:   "text/plain; charset=utf-8",
			},
		},
	}
	sample_put_bytes, _ := sample_put_response.ToJSON()

	type args struct {
		progressChan chan<- interface{}
	}
	tests := []struct {
		name       string
		httpServer *http.Server
		isHttps    bool
		Collector  *troubleshootv1beta2.HTTP
		args       args
		want       CollectorResult
		// wantHeaders map[string]string
		wantErr bool
	}{
		{
			// check valid file path when CollectorName is not supplied
			name: "GET: collector name unset",
			Collector: &troubleshootv1beta2.HTTP{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "",
				},
				Get: &troubleshootv1beta2.Get{},
			},
			args: args{
				progressChan: nil,
			},
			want: CollectorResult{
				"result.json": sample_get_bytes,
			},
			wantErr: false,
			isHttps: false,
		},
		{
			// check valid file path when CollectorName is supplied
			name: "GET: valid collect",
			Collector: &troubleshootv1beta2.HTTP{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "example-com",
				},
				Get: &troubleshootv1beta2.Get{
					// InsecureSkipVerify: true,
					// Timeout: 5,
				},
			},
			args: args{
				progressChan: nil,
			},
			want: CollectorResult{
				"example-com.json": sample_get_bytes,
			},
			wantErr: false,
			isHttps: false,
		},
		{
			// check valid file path when CollectorName is supplied
			name: "POST: valid collect",
			Collector: &troubleshootv1beta2.HTTP{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "example-com",
				},
				Post: &troubleshootv1beta2.Post{
					InsecureSkipVerify: true,
					Body:               `{"id": 123, "name": "John Doe"}`,
					// Timeout:            5,
					Headers: map[string]string{"Content-Type": "application/json"},
				},
			},
			args: args{
				progressChan: nil,
			},
			want: CollectorResult{
				"example-com.json": sample_post_bytes,
			},
			wantErr: false,
			isHttps: false,
		},
		{
			// check valid file path when CollectorName is supplied
			name: "PUT: valid collect",
			Collector: &troubleshootv1beta2.HTTP{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "example-com",
				},
				Put: &troubleshootv1beta2.Put{
					Body: `{"id": 123, "name": "John Doe"}`,
					// Timeout: 5,
					Headers: map[string]string{"Content-Type": "application/json"},
				},
			},
			args: args{
				progressChan: nil,
			},
			want: CollectorResult{
				"example-com.json": sample_put_bytes,
			},
			wantErr: false,
			isHttps: false,
		},
		// add TLS cert case
	}
	for _, tt := range tests {
		var ts *httptest.Server
		if tt.isHttps {
			ts = httptest.NewTLSServer(mux)
		} else {
			ts = httptest.NewServer(mux)
		}
		url := ts.URL
		defer ts.Close()

		// client := ts.Client()
		// rs, err := client.Get(ts.URL + "/get")
		// rs, err := client.Post(ts.URL+"/post", "application/json", bytes.NewBuffer([]byte(`{"id": 123, "name": "John Doe"}`)))

		// if err != nil {
		// 	t.Fatalf("failed to get index.html: %v", err)
		// }
		// if rs.StatusCode != http.StatusOK {
		// 	t.Errorf("expected status code %d, got %d", http.StatusOK, rs.StatusCode)
		// }
		// fmt.Println("pre-test-launch rs", rs)

		req, err := http.NewRequest("POST", url+"/post", strings.NewReader(`{"id": 123, "name": "John Doe"}`))
		if err != nil {
			t.Fatalf("failed to get index.html: %v", err)
		}
		// req.Header.Set("Content-Type", "application/json")
		// w := httptest.NewRecorder()
		// handler(w, req)

		// resp := w.Result()
		// body, _ := io.ReadAll(resp.Body)

		// fmt.Println(resp.StatusCode)
		// fmt.Println(resp.Header.Get("Content-Type"))
		// fmt.Println(string(body))
		resp, err := ts.Client().Do(req)
		if err != nil {
			t.Fatalf("failed to get index.html: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
		greeting, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("greeting: %s\n", greeting)
		defer resp.Body.Close()

		// req.Header.Set("Content-Type", "application/json")
		// unmarshal req.Body
		// fmt.Println("req", req)

		t.Run(tt.name, func(t *testing.T) {
			var resp ResponseData
			var response_type *ResponseData
			c := &CollectHTTP{
				Collector: tt.Collector,
			}

			switch {
			case c.Collector.Get != nil:
				response_type = sample_get_response
				c.Collector.Get.URL = fmt.Sprintf("%s%s", url, "/get")
				fmt.Println("Requested: ", c.Collector.Get)
			case c.Collector.Post != nil:
				response_type = sample_post_response
				c.Collector.Post.URL = fmt.Sprintf("%s%s", url, "/post")
				fmt.Println("Requested: ", c.Collector.Post)
			case c.Collector.Put != nil:
				response_type = sample_put_response
				c.Collector.Put.URL = fmt.Sprintf("%s%s", url, "/put")
				fmt.Println("Requested: ", c.Collector.Put)
			}
			got, err := c.Collect(tt.args.progressChan)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectHTTP.Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// fmt.Println(c.Collector.CollectorName)
			var expected_filename string
			if c.Collector.CollectorName == "" {
				expected_filename = "result.json"
			} else {
				expected_filename = c.Collector.CollectorName + ".json"
			}

			if err := json.Unmarshal(got[expected_filename], &resp); err != nil {
				t.Errorf("CollectHTTP.Collect() error = %v, wantErr %v", err, tt.wantErr)
				// t.Errorf("CollectHTTP.Collect() JSON Response Unmarshal error = %v", err)
				return
			}
			gotString := make(map[string]string)
			// wantString := make(map[string]string)
			for key, value := range got {
				gotString[key] = string(value)
			}
			fmt.Println("gotString: ", gotString)
			fmt.Println("Unmarshalled resp: ", resp)

			// Correct format of the collected data (JSON data)
			// assert.Equal(t, response_type.Response.Status, resp.Response.Status)
			// assert.Equal(t, response_type.Response.Body, resp.Response.Body)
			assert.Equal(t, response_type.Response.Headers.ContentLength, resp.Response.Headers.ContentLength)
			// assert.Equal(t, response_type.Response.Headers.ContentType, resp.Response.Headers.ContentType)

			// // make another request to the server at c.Collector.Get.URL to check if the server is still running
			// resp_new, err := http.Get(url)
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// 	return
			// }
			// defer resp_new.Body.Close()
			// fmt.Println("NEW Status Code:", resp_new.StatusCode)
			// fmt.Println("NEW Headers:", resp_new)

			// for key, value := range tt.want {
			// 	wantString[key] = string(value)
			// }
			// t.Errorf("CollectHTTP.Collect() = %v, want %v", got, tt.want)

			// t.Errorf("CollectHTTP.Collect() = %v, want %v", gotString, wantString)

			// assert.JSONEq(t, string(tt.want["example-com.json"]), string(got["example-com.json"]))

			// JSONIgnoreFieldsEqual(
			// 	t,
			// 	string(tt.want["example-com.json"]),
			// 	string(got["example-com.json"]),
			// 	[]string{"response.headers.Date"},
			// )
			// Compare JSON structures while ignoring the "Date" field
			// expectedJSON, err := jason.NewObjectFromBytes([]byte(tt.want["example-com.json"]))
			// assert.NoError(t, err)

			// gotJSON, err := jason.NewObjectFromBytes([]byte(got["example-com.json"]))
			// assert.NoError(t, err)

			// rc_expected, err := expectedJSON.GetString("response", "status")
			// rc_got, err := gotJSON.GetString("response", "status")
			// if rc_expected != rc_got {
			// 	t.Errorf("CollectHTTP.Collect() error = %v, want %v", rc_got, rc_expected)
			// 	return
			// }
			// assert.Equal(t, rc_expected, rc_got)

			// // Remove the "Date" field from both JSON objects
			// expectedJSON.Delete("response", "headers", "Date")
			// gotJSON.Delete("response", "headers", "Date")

			// // Compare the modified JSON objects
			// assert.True(t, expectedJSON.Equal(gotJSON))

			// reset got, err
			got = nil
			err = nil
			response_type = nil
			// time.Sleep(2 * time.Second)
		})
	}
}

// timeout, get requests s3:// http://
// add invalid fields collector schema (out of bounds ints, etc)
// change schema indentation
// If the collectorName field is unset it will be named result.json.
// https://troubleshoot.sh/docs/collect/http/#example-collector-definition
func TestCollectHTTP_Timeouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second) // Simulate a slow server
	}))
	defer server.Close()
	url := server.URL

	type args struct {
		progressChan chan<- interface{}
	}

	// sample_get_response := &ResponseData{
	// 	Response: Response{
	// 		Status: 200,
	// 		Body:   "{\"status\": \"healthy\"}",
	// 		Headers: Headers{
	// 			ContentLength: "11",
	// 			ContentType:   "application/json; charset=utf-8",
	// 		},
	// 	},
	// }
	// sample_get_bytes, _ := sample_get_response.ToJSON()

	tests := []struct {
		name      string
		Collector *troubleshootv1beta2.HTTP
		args      args
		want      ErrorResponse
		wantErr   bool
	}{
		{
			// check valid file path when CollectorName is supplied
			name: "GET: check request timeout < server delay (exit early)",
			Collector: &troubleshootv1beta2.HTTP{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "example-com",
				},
				Get: &troubleshootv1beta2.Get{
					URL:     url,
					Timeout: 1,
				},
			},
			args: args{
				progressChan: nil,
			},
			want: ErrorResponse{
				Error: HTTPError{
					Message: "context deadline exceeded",
				},
			},
			wantErr: true,
		},
		// {
		// 	// check valid file path when CollectorName is supplied
		// 	name: "GET: check request timeout > server delay",
		// 	Collector: &troubleshootv1beta2.HTTP{
		// 		CollectorMeta: troubleshootv1beta2.CollectorMeta{
		// 			CollectorName: "example-com",
		// 		},
		// 		Get: &troubleshootv1beta2.Get{
		// 			URL:     url,
		// 			Timeout: 3,
		// 		},
		// 	},
		// 	args: args{
		// 		progressChan: nil,
		// 	},
		// 	want: CollectorResult{
		// 		"result.json": sample_get_bytes,
		// 	},
		// 	wantErr: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CollectHTTP{
				Collector: tt.Collector,
			}

			got, err := c.Collect(tt.args.progressChan)
			if err != nil {
				t.Errorf("CollectHTTP.Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			expected_filename := c.Collector.CollectorName + ".json"
			var resp ErrorResponse

			if err := json.Unmarshal(got[expected_filename], &resp); err != nil {
				t.Errorf("CollectHTTP.Collect() error = %v, wantErr %v", err, tt.wantErr)
				// t.Errorf("CollectHTTP.Collect() JSON Response Unmarshal error = %v", err)
				return
			}

			if strings.Contains(resp.Error.Message, "context deadline exceeded") {
				fmt.Println("Error message:", resp.Error.Message)
			} else {
				t.Errorf("Error message [%v] does not indicate [%v]", resp.Error.Message, tt.want.Error.Message)
			}

			// check resp type is ErrorResponse and not ResponseData

			// switch v := resp.(type) {
			// case ErrorResponse:
			// 	// Response is ErrorResponse
			// 	// check if v.error.message contains want

			// case ResponseData:
			// 	// Response is ResponseData
			// 	fmt.Println("Response data:", v.Response.Status)
			// 	t.Errorf("CollectHTTP.Collect() return type = %v, wantErr %v", v, tt.wantErr)
			// 	// default:
			// 	// 	// Unknown type
			// 	// 	fmt.Println("Unknown type returned: ", v)
			// }

			// var response_type *ResponseData
			// response_type = sample_get_response
			// assert.Equal(t, response_type.Response.Status, resp.Response.Status)
			// assert.Equal(t, response_type.Response.Body, resp.Response.Body)
			// assert.Equal(t, response_type.Response.Headers.ContentLength, resp.Response.Headers.ContentLength)
			// assert.Equal(t, response_type.Response.Headers.ContentType, resp.Response.Headers.ContentType)
			// var resp ResponseData
			// fmt.Println(c.Collector.CollectorName)

			gotString := make(map[string]string)
			// wantString := make(map[string]string)
			for key, value := range got {
				gotString[key] = string(value)
			}
			fmt.Println("got", gotString)
		})
	}
}

func isTimeoutError(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

// func TestGetCollector(t *testing.T) {
// 	type args struct {
// 		collector    *troubleshootv1beta2.Collect
// 		bundlePath   string
// 		namespace    string
// 		clientConfig *rest.Config
// 		client       kubernetes.Interface
// 		sinceTime    *time.Time
// 	}
// 	tests := []struct {
// 		name  string
// 		args  args
// 		want  interface{}
// 		want1 bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1 := GetCollector(tt.args.collector, tt.args.bundlePath, tt.args.namespace, tt.args.clientConfig, tt.args.client, tt.args.sinceTime)
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetCollector() got = %v, want %v", got, tt.want)
// 			}
// 			if got1 != tt.want1 {
// 				t.Errorf("GetCollector() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }
