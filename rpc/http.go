package rpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type MsHttpClient struct {
	client     http.Client
	serviceMap map[string]MsService
}

func NewHttpClient() *MsHttpClient {
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return &MsHttpClient{
		client:     client,
		serviceMap: map[string]MsService{},
	}
}

func (c *MsHttpClient) Get(url string) ([]byte, error) {

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *MsHttpClient) responseHandle(request *http.Request) ([]byte, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status is %d", response.StatusCode))
	}
	reader := bufio.NewReader(response.Body)
	buffLen := 79
	buff := make([]byte, buffLen)
	body := make([]byte, 0)
	for {
		n, err := reader.Read(buff)
		if err == io.EOF || n == 0 {
			break
		}
		body = append(body, buff[:n]...)
		if n < buffLen {
			break
		}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type HttpConfig struct {
	Host string
	Port int
}
type MsService interface {
	Env() HttpConfig
}

func (c *MsHttpClient) RegisterHttpService(serviceName string, service MsService) {
	c.serviceMap[serviceName] = service
}

func (c *MsHttpClient) Do(service string, method string) MsService {
	msService, ok := c.serviceMap[service]
	if !ok {
		panic(errors.New("service not found"))
	}
	//找到service里面的Field, 给其中调用的方法赋值
	t := reflect.TypeOf(msService)
	v := reflect.ValueOf(msService)
	if t.Kind() != reflect.Pointer {

	}
}
