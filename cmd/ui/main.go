package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	uiPort := getEnv("UI_PORT", "8090")
	apiBase := getEnv("API_BASE_URL", "http://localhost:8080")

	target, err := url.Parse(apiBase)
	if err != nil {
		log.Fatalf("invalid API_BASE_URL: %v", err)
	}

	proxy := newAPIProxy(target)

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveUI)
	mux.Handle("/api/", proxy)

	server := &http.Server{
		Addr:              ":" + uiPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
	}

	log.Printf("UI started on http://localhost:%s", uiPort)
	log.Printf("Proxying /api/* to %s", apiBase)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ui server failed: %v", err)
	}
}

func newAPIProxy(target *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		req.Host = target.Host
	}

	return proxy
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

const htmlPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Withdrawal UI</title>
  <style>
    :root {
      --bg: #f3f5f7;
      --card: #ffffff;
      --text: #111827;
      --muted: #4b5563;
      --primary: #0f766e;
      --secondary: #1f2937;
      --danger: #b91c1c;
      --border: #e5e7eb;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 20px;
      font-family: "Segoe UI", Tahoma, sans-serif;
      color: var(--text);
      background: radial-gradient(circle at top left, #dff5ef 0%, var(--bg) 58%);
      min-height: 100vh;
    }
    .wrap { max-width: 980px; margin: 0 auto; display: grid; gap: 14px; }
    h1 { margin: 0; font-size: 26px; }
    .hint { color: var(--muted); }
    .grid { display: grid; gap: 14px; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
    .card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 14px;
      box-shadow: 0 8px 20px rgba(0,0,0,.05);
    }
    .card h2 { margin: 0 0 10px; font-size: 18px; }
    .row { display: grid; gap: 8px; margin-bottom: 10px; }
    label { font-size: 13px; color: var(--muted); }
    input, textarea {
      width: 100%;
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 10px 12px;
      font-size: 14px;
      font-family: inherit;
      color: var(--text);
      background: #fff;
    }
    textarea { min-height: 90px; resize: vertical; font-family: Consolas, "Courier New", monospace; }
    button {
      border: 0;
      border-radius: 10px;
      background: var(--primary);
      color: #fff;
      font-weight: 600;
      padding: 10px 12px;
      cursor: pointer;
    }
    button.secondary { background: var(--secondary); }
    button.danger { background: var(--danger); }
    pre {
      margin: 0;
      border-radius: 10px;
      background: #0b1220;
      color: #dbeafe;
      padding: 12px;
      min-height: 160px;
      overflow: auto;
      font-size: 12px;
      line-height: 1.4;
    }
    .status {
      display: inline-block;
      margin-top: 8px;
      padding: 4px 8px;
      border-radius: 8px;
      background: #ecfeff;
      color: #155e75;
      font-size: 12px;
      font-weight: 600;
    }
  </style>
</head>
<body>
  <div class="wrap">
    <div>
      <h1>Withdrawal UI</h1>
      <div class="hint">JWT хранится в браузере и автоматически отправляется во все защищённые запросы.</div>
    </div>

    <div class="card">
      <h2>Вход</h2>
      <div class="row"><label>username</label><input id="username" value="admin"></div>
      <div class="row"><label>password</label><input id="password" type="password" value="admin123"></div>
      <div class="row" style="grid-template-columns: 1fr 1fr;">
        <button onclick="login()">Войти</button>
        <button class="danger" onclick="logout()">Выйти</button>
      </div>
      <div id="authStatus" class="status">Не авторизован</div>
    </div>

    <div class="grid">
      <div class="card">
        <h2>POST /v1/withdrawals</h2>
        <div class="row"><label>balance_id</label><input id="balanceId" type="number" value="1"></div>
        <div class="row"><label>amount</label><input id="amount" type="number" value="10"></div>
        <div class="row"><label>destination</label><input id="destination" value="wallet_ui"></div>
        <div class="row"><label>idempotency_key</label><input id="idempotencyKey" placeholder="auto if empty"></div>
        <button onclick="createWithdrawal()">Создать списание</button>
      </div>

      <div class="card">
        <h2>GET /v1/withdrawals</h2>
        <button onclick="getWithdrawals()">Обновить список</button>
      </div>

      <div class="card">
        <h2>GET /v1/withdrawals/{id}</h2>
        <div class="row"><label>withdrawal id</label><input id="withdrawalId" placeholder="uuid"></div>
        <button onclick="getWithdrawal()">Получить по id</button>
      </div>
    </div>

    <div class="card">
      <h2>Response</h2>
      <pre id="result">No requests yet.</pre>
    </div>
  </div>

<script>
    const API = "/api";
    const LS_ACCESS = "ws_access_token";
    const LS_REFRESH = "ws_refresh_token";
    const LS_USERNAME = "ws_username";
    const LS_PASSWORD = "ws_password";

    function render(title, status, data) {
      document.getElementById("result").textContent = JSON.stringify({
        at: new Date().toISOString(),
        request: title,
        status: status,
        response: data
      }, null, 2);
    }

    function nowKey() {
      return "ui-" + Date.now() + "-" + Math.floor(Math.random() * 100000);
    }

    function access() { return (localStorage.getItem(LS_ACCESS) || "").trim(); }
    function refresh() { return (localStorage.getItem(LS_REFRESH) || "").trim(); }

    function authHeader(token) {
      return token ? { Authorization: "Bearer " + token } : {};
    }

    async function api(method, path, body, token) {
      const res = await fetch(API + path, {
        method,
        headers: { "Content-Type": "application/json", ...authHeader(token) },
        body: body ? JSON.stringify(body) : undefined
      });

      let data;
      try { data = await res.json(); }
      catch (_) { data = { raw: await res.text() }; }

      return { status: res.status, data };
    }

    function setAuthStatus() {
      const token = access();
      const username = (localStorage.getItem(LS_USERNAME) || "").trim();
      const node = document.getElementById("authStatus");
      if (!token) {
        node.textContent = "Не авторизован";
        return;
      }
      node.textContent = "Авторизован: " + (username || "unknown");
    }

    function setTokens(a, r) {
      if (a !== undefined) localStorage.setItem(LS_ACCESS, a);
      if (r !== undefined) localStorage.setItem(LS_REFRESH, r);
      setAuthStatus();
    }

    async function login() {
      const username = document.getElementById("username").value.trim();
      const password = document.getElementById("password").value;
      const res = await api("POST", "/v1/auth/token", { username, password });
      if (res.status === 200 && res.data) {
        setTokens(res.data.access_token, res.data.refresh_token);
        localStorage.setItem(LS_USERNAME, username);
        localStorage.setItem(LS_PASSWORD, password);
        await getWithdrawals();
      }
      render("POST /v1/auth/token", res.status, res.data);
    }

    async function tryRefresh() {
      if (!refresh()) return false;
      const res = await api("POST", "/v1/auth/refresh", { refresh_token: refresh() });
      if (res.status === 200 && res.data) {
        setTokens(res.data.access_token, res.data.refresh_token);
        return true;
      }
      return false;
    }

    function logout() {
      localStorage.removeItem(LS_ACCESS);
      localStorage.removeItem(LS_REFRESH);
      localStorage.removeItem(LS_USERNAME);
      localStorage.removeItem(LS_PASSWORD);
      setAuthStatus();
      render("logout", 200, { ok: true });
    }

    async function createWithdrawal() {
      const idemInput = document.getElementById("idempotencyKey");
      if (!idemInput.value.trim()) idemInput.value = nowKey();
      const res = await api("POST", "/v1/withdrawals", {
        balance_id: Number(document.getElementById("balanceId").value),
        amount: Number(document.getElementById("amount").value),
        destination: document.getElementById("destination").value.trim(),
        idempotency_key: idemInput.value.trim()
      }, access());
      if (res.status === 401 && await tryRefresh()) return createWithdrawal();
      if (res.status === 200 && res.data && res.data.id) document.getElementById("withdrawalId").value = res.data.id;
      render("POST /v1/withdrawals", res.status, res.data);
    }

    async function getWithdrawals() {
      const res = await api("GET", "/v1/withdrawals", null, access());
      if (res.status === 401 && await tryRefresh()) return getWithdrawals();
      render("GET /v1/withdrawals", res.status, res.data);
    }

    async function getWithdrawal() {
      const id = document.getElementById("withdrawalId").value.trim();
      if (!id) { render("GET /v1/withdrawals/{id}", 400, { error: "withdrawal id is required" }); return; }
      const res = await api("GET", "/v1/withdrawals/" + encodeURIComponent(id), null, access());
      if (res.status === 401 && await tryRefresh()) return getWithdrawal();
      render("GET /v1/withdrawals/{id}", res.status, res.data);
    }

    (async function init() {
      document.getElementById("username").value = localStorage.getItem(LS_USERNAME) || "admin";
      document.getElementById("password").value = localStorage.getItem(LS_PASSWORD) || "admin123";
      setAuthStatus();

      if (!access() && localStorage.getItem(LS_USERNAME) && localStorage.getItem(LS_PASSWORD)) {
        await login();
        return;
      }

      if (access()) {
        await getWithdrawals();
      }
    })();
  </script>
</body>
</html>`
