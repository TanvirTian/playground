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

func (s *server) getSession(w http.ResponseWriter, r *http.Request) (string, *myinterp.Interpreter) {
    cookie, err := r.Cookie("sid")
    if err != nil || cookie.Value == "" {
        sid := newSID()

        http.SetCookie(w, &http.Cookie{
            Name:     "sid",
            Value:    sid,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,                   
            SameSite: http.SameSiteNoneMode,  
        })
        s.mu.Lock()
        s.sessions[sid] = myinterp.NewInterpreter()
        s.mu.Unlock()
        return sid, s.sessions[sid]
    }
    sid := cookie.Value
    s.mu.Lock()
    interp := s.sessions[sid]
    if interp == nil {
        interp = myinterp.NewInterpreter()
        s.sessions[sid] = interp
    }
    s.mu.Unlock()
    return sid, interp
}

func newSID() string {
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

    _, interp := s.getSession(w, r)

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
    writeJSON(w, 200, evalResp{OK: true, Result: res})
}

func (s *server) handleSession(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    sid, _ := s.getSession(w, r)
    writeJSON(w, 200, map[string]string{"session": sid})
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    http.ServeFile(w, r, path.Join("public", "index.html"))
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
            w.Header().Set("Access-Control-Allow-Origin", origin) // cannot be *
            w.Header().Set("Vary", "Origin")
        }
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Credentials", "true") // allow cookies

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

    handler := withCORS(mux)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Println("Playground running on port:", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}
