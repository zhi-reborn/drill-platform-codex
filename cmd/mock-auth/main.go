package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type mockUser struct {
	Username   string   `json:"username"`
	RealName   string   `json:"real_name"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	Department string   `json:"department"`
	Groups     []string `json:"groups"`
}

var users = map[string]mockUser{
	"admin": {
		Username: "admin", RealName: "系统管理员", Email: "admin@example.local",
		Phone: "13800000001", Department: "技术部", Groups: []string{"drill-admin"},
	},
	"director": {
		Username: "director", RealName: "演练指挥员", Email: "director@example.local",
		Phone: "13800000002", Department: "指挥中心", Groups: []string{"drill-director"},
	},
	"executor": {
		Username: "executor", RealName: "任务执行员", Email: "executor@example.local",
		Phone: "13800000003", Department: "运维部", Groups: []string{"drill-executor"},
	},
	"viewer": {
		Username: "viewer", RealName: "观察员", Email: "viewer@example.local",
		Phone: "13800000004", Department: "审计部", Groups: []string{"drill-viewer"},
	},
}

var tickets = struct {
	sync.Mutex
	data map[string]string
}{data: make(map[string]string)}

var loginTpl = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <title>Mock CAS Login</title>
  <style>
    body { margin:0; min-height:100vh; display:grid; place-items:center; background:#0f172a; color:#e5e7eb; font-family: system-ui, sans-serif; }
    main { width: 420px; padding: 28px; border: 1px solid #334155; border-radius: 12px; background:#111827; }
    h1 { margin:0 0 8px; font-size:26px; }
    p { color:#94a3b8; margin:0 0 20px; }
    a { display:flex; justify-content:space-between; align-items:center; margin:10px 0; padding:12px 14px; border-radius:8px; color:#e5e7eb; text-decoration:none; background:#1f2937; border:1px solid #374151; }
    a:hover { border-color:#38bdf8; background:#0f2742; }
    small { color:#38bdf8; }
  </style>
</head>
<body>
  <main>
    <h1>Mock CAS Login</h1>
    <p>选择一个模拟企业账号完成 CAS 登录。</p>
    {{range .Users}}
      <a href="/cas/mock-login?service={{$.Service}}&user={{.Username}}">
        <span>{{.RealName}} / {{.Username}}</span><small>{{index .Groups 0}}</small>
      </a>
    {{end}}
  </main>
</body>
</html>`))

func main() {
	http.HandleFunc("/cas/login", handleCASLogin)
	http.HandleFunc("/cas/mock-login", handleMockLogin)
	http.HandleFunc("/cas/serviceValidate", handleServiceValidate)
	http.HandleFunc("/ldap/users/", handleLDAPUser)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":9091"
	log.Printf("mock CAS/LDAP service listening on http://127.0.0.1%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleCASLogin(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service == "" {
		http.Error(w, "missing service", http.StatusBadRequest)
		return
	}
	list := make([]mockUser, 0, len(users))
	for _, user := range users {
		list = append(list, user)
	}
	_ = loginTpl.Execute(w, map[string]any{
		"Service": service,
		"Users":   list,
	})
}

func handleMockLogin(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	username := r.URL.Query().Get("user")
	if service == "" || username == "" {
		http.Error(w, "missing service or user", http.StatusBadRequest)
		return
	}
	if _, ok := users[username]; !ok {
		http.Error(w, "unknown user", http.StatusBadRequest)
		return
	}
	ticket := newTicket()
	tickets.Lock()
	tickets.data[ticket] = username
	tickets.Unlock()

	u, err := url.Parse(service)
	if err != nil {
		http.Error(w, "bad service", http.StatusBadRequest)
		return
	}
	q := u.Query()
	q.Set("ticket", ticket)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func handleServiceValidate(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	tickets.Lock()
	username, ok := tickets.data[ticket]
	if ok {
		delete(tickets.data, ticket)
	}
	tickets.Unlock()

	w.Header().Set("Content-Type", "application/xml;charset=utf-8")
	if !ok {
		_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas"><cas:authenticationFailure code="INVALID_TICKET">ticket not recognized</cas:authenticationFailure></cas:serviceResponse>`)
		return
	}
	resp := struct {
		XMLName xml.Name `xml:"cas:serviceResponse"`
		Xmlns   string   `xml:"xmlns:cas,attr"`
		Success struct {
			User string `xml:"cas:user"`
		} `xml:"cas:authenticationSuccess"`
	}{Xmlns: "http://www.yale.edu/tp/cas"}
	resp.Success.User = username
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(resp)
}

func handleLDAPUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/ldap/users/")
	username, _ = url.PathUnescape(username)
	user, ok := users[username]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func newTicket() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return "ST-" + hex.EncodeToString(buf)
}
