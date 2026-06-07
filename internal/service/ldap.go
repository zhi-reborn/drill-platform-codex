package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type LDAPConfig struct {
	Enabled             bool   `yaml:"enabled"`
	URL                 string `yaml:"url"`
	BindDN              string `yaml:"bindDN"`
	BindPassword        string `yaml:"bindPassword"`
	BaseDN              string `yaml:"baseDN"`
	UserFilter          string `yaml:"userFilter"`
	UsernameAttribute   string `yaml:"usernameAttribute"`
	RealNameAttribute   string `yaml:"realNameAttribute"`
	EmailAttribute      string `yaml:"emailAttribute"`
	PhoneAttribute      string `yaml:"phoneAttribute"`
	DepartmentAttribute string `yaml:"departmentAttribute"`
	GroupBaseDN         string `yaml:"groupBaseDN"`
	GroupFilter         string `yaml:"groupFilter"`
	GroupNameAttribute  string `yaml:"groupNameAttribute"`
	TimeoutSeconds      int    `yaml:"timeoutSeconds"`
}

type LDAPClient struct {
	cfg LDAPConfig
}

func NewLDAPClient(cfg LDAPConfig) *LDAPClient {
	return &LDAPClient{cfg: normalizeLDAPConfig(cfg)}
}

func normalizeLDAPConfig(cfg LDAPConfig) LDAPConfig {
	if cfg.UserFilter == "" {
		cfg.UserFilter = "(uid=%s)"
	}
	if cfg.UsernameAttribute == "" {
		cfg.UsernameAttribute = "uid"
	}
	if cfg.RealNameAttribute == "" {
		cfg.RealNameAttribute = "cn"
	}
	if cfg.EmailAttribute == "" {
		cfg.EmailAttribute = "mail"
	}
	if cfg.PhoneAttribute == "" {
		cfg.PhoneAttribute = "mobile"
	}
	if cfg.DepartmentAttribute == "" {
		cfg.DepartmentAttribute = "department"
	}
	if cfg.GroupFilter == "" {
		cfg.GroupFilter = "(memberUid=%s)"
	}
	if cfg.GroupNameAttribute == "" {
		cfg.GroupNameAttribute = "cn"
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 10
	}
	return cfg
}

func (c *LDAPClient) LookupUser(username string) (ExternalUser, error) {
	if c.cfg.URL == "" || c.cfg.BaseDN == "" {
		return ExternalUser{}, errors.New("LDAP 地址或 BaseDN 未配置")
	}
	if strings.HasPrefix(c.cfg.URL, "http://") || strings.HasPrefix(c.cfg.URL, "https://") {
		return c.lookupHTTPMockUser(username)
	}
	conn, err := ldap.DialURL(c.cfg.URL)
	if err != nil {
		return ExternalUser{}, err
	}
	defer conn.Close()
	conn.SetTimeout(time.Duration(c.cfg.TimeoutSeconds) * time.Second)

	if c.cfg.BindDN != "" {
		if err := conn.Bind(c.cfg.BindDN, c.cfg.BindPassword); err != nil {
			return ExternalUser{}, err
		}
	}

	attrs := []string{
		c.cfg.UsernameAttribute,
		c.cfg.RealNameAttribute,
		c.cfg.EmailAttribute,
		c.cfg.PhoneAttribute,
		c.cfg.DepartmentAttribute,
	}
	req := ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		c.cfg.TimeoutSeconds,
		false,
		fmt.Sprintf(c.cfg.UserFilter, ldap.EscapeFilter(username)),
		attrs,
		nil,
	)
	res, err := conn.Search(req)
	if err != nil {
		return ExternalUser{}, err
	}
	if len(res.Entries) == 0 {
		return ExternalUser{}, errors.New("LDAP 用户不存在")
	}

	entry := res.Entries[0]
	ext := ExternalUser{
		Username:   firstNonEmpty(entry.GetAttributeValue(c.cfg.UsernameAttribute), username),
		RealName:   entry.GetAttributeValue(c.cfg.RealNameAttribute),
		Email:      entry.GetAttributeValue(c.cfg.EmailAttribute),
		Phone:      entry.GetAttributeValue(c.cfg.PhoneAttribute),
		Department: entry.GetAttributeValue(c.cfg.DepartmentAttribute),
	}
	ext.Groups = c.lookupGroups(conn, username)
	return ext, nil
}

func (c *LDAPClient) lookupHTTPMockUser(username string) (ExternalUser, error) {
	base := strings.TrimRight(c.cfg.URL, "/")
	reqURL := base + "/users/" + url.PathEscape(username)
	client := &http.Client{Timeout: time.Duration(c.cfg.TimeoutSeconds) * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return ExternalUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ExternalUser{}, errors.New("LDAP 用户不存在")
	}
	if resp.StatusCode != http.StatusOK {
		return ExternalUser{}, fmt.Errorf("LDAP mock 查询失败: HTTP %d", resp.StatusCode)
	}
	var ext ExternalUser
	if err := json.NewDecoder(resp.Body).Decode(&ext); err != nil {
		return ExternalUser{}, err
	}
	if ext.Username == "" {
		ext.Username = username
	}
	return ext, nil
}

func (c *LDAPClient) lookupGroups(conn *ldap.Conn, username string) []string {
	if c.cfg.GroupBaseDN == "" {
		return nil
	}
	req := ldap.NewSearchRequest(
		c.cfg.GroupBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		c.cfg.TimeoutSeconds,
		false,
		fmt.Sprintf(c.cfg.GroupFilter, ldap.EscapeFilter(username)),
		[]string{c.cfg.GroupNameAttribute},
		nil,
	)
	res, err := conn.Search(req)
	if err != nil {
		return nil
	}
	groups := make([]string, 0, len(res.Entries))
	for _, entry := range res.Entries {
		group := strings.TrimSpace(entry.GetAttributeValue(c.cfg.GroupNameAttribute))
		if group != "" {
			groups = append(groups, group)
		}
	}
	return groups
}
