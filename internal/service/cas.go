package service

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CASConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ServerURL  string `yaml:"serverURL"`
	PublicURL  string `yaml:"publicURL"`
	ServiceURL string `yaml:"serviceURL"`
}

type CASClient struct {
	serverURL  string
	httpClient *http.Client
}

func NewCASClient(serverURL string) *CASClient {
	return &CASClient{
		serverURL: strings.TrimRight(serverURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func BuildCASLoginURL(serverURL, serviceURL string) (string, error) {
	if serverURL == "" {
		return "", errors.New("CAS 地址未配置")
	}
	u, err := url.Parse(strings.TrimRight(serverURL, "/") + "/login")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("service", serviceURL)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *CASClient) ValidateTicket(ticket, serviceURL string) (string, error) {
	if ticket == "" {
		return "", errors.New("CAS ticket 为空")
	}
	u, err := url.Parse(c.serverURL + "/serviceValidate")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("ticket", ticket)
	q.Set("service", serviceURL)
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CAS 校验失败: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return ParseCASServiceResponse(body)
}

func ParseCASServiceResponse(body []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(body))
	var inSuccess bool
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "authenticationSuccess":
				inSuccess = true
			case "authenticationFailure":
				var msg string
				if err := decoder.DecodeElement(&msg, &t); err != nil {
					return "", err
				}
				return "", fmt.Errorf("CAS 认证失败: %s", strings.TrimSpace(msg))
			case "user":
				if inSuccess {
					var username string
					if err := decoder.DecodeElement(&username, &t); err != nil {
						return "", err
					}
					username = strings.TrimSpace(username)
					if username == "" {
						return "", errors.New("CAS 响应用户名为空")
					}
					return username, nil
				}
			}
		case xml.EndElement:
			if t.Name.Local == "authenticationSuccess" {
				inSuccess = false
			}
		}
	}
	return "", errors.New("CAS 响应未包含认证成功信息")
}
