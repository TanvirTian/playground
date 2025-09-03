package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path"
    "strings"
    "sync"

    "playground/interpreter/lexer"
    "playground/interpreter/parser"
    myinterp "playground/interpreter/interpreter"
)

type server struct {
    mu       sync.Mutex
    sessions map[string]*myinterp.Interpreter
}

func newServer() *server {
    return &server{sessions: map[string]*myinterp.Interpreter{}}
}


func (s *server) getSession(r *http.Request) (string, *myinterp.Interpreter) {
    token := ""
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        token = strings.TrimPrefix(auth, "Bearer ")
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    if token == "" || s.sessions[token] == nil {
        token = newToken()
        s.sessions[token] = myinterp.NewInterpreter()
    }

    return token, s.sessions[token]
}

func newToken() string {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        panic(err)
    }
    return hex.EncodeToString(b)
}

type evalReq struct {
    Line string `json:"line"`
}

type evalResp struct {
    OK     bool        `json:"ok"`
    Result interface{} `json:"result,omitempty"`
    Error  string      `json:"error,omitempty"`
}

func (s *server) handleEval(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    token, interp := s.getSession(r)

    var req evalReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, evalResp{OK: false, Error: "bad request"})
        return
    }

    line := strings.TrimSpace(req.Line)
    if line == "" {
        writeJSON(w, 200, evalResp{OK: true, Result: ""})
        return
    }

    defer func() {
        if rec := recover(); rec != nil {
            errMsg := fmt.Sprint(rec)
            firstLine := strings.Split(errMsg, "\n")[0]
            fmt.Println(firstLine)
            writeJSON(w, 200, evalResp{OK: false, Error: firstLine})
        }
    }()

    lex := lexer.NewLexer(line)
    par := parser.NewParser(lex)
    res := interp.Interpret(par)
    w.Header().Set("X-Session-Token", token) 
    writeJSON(w, 200, evalResp{OK: true, Result: res})
}

func (s *server) handleSession(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    token, _ := s.getSession(r)
    writeJSON(w, 200, map[string]string{"token": token})
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, path.Join("public", "index.html"))
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if origin != "" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Vary", "Origin")
        }
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    s := newServer()

    mux := http.NewServeMux()
    mux.HandleFunc("/", s.handleIndex)
    mux.HandleFunc("/eval", s.handleEval)
    mux.HandleFunc("/session", s.handleSession)
    mux.HandleFunc("/health", s.handleHealth)

    handler := withCORS(mux)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Println("Playground running on port:", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}
