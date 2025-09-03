// main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type APIResponse struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// Prometheus metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

// Mock database
var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()},
    {ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()},
}

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
        next.ServeHTTP(w, r)
        duration := time.Since(start)
        log.Printf("Request completed in %v", duration)
    })
}

func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        duration := time.Since(start).Seconds()
        
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := APIResponse{
        Status: "healthy",
        Data: map[string]interface{}{
            "timestamp": time.Now(),
            "service":   "User API",
            "version":   "1.0.0",
        },
    }
    json.NewEncoder(w).Encode(response)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    response := APIResponse{
        Status: "success",
        Data:   users,
    }
    json.NewEncoder(w).Encode(response)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        response := APIResponse{
            Status:  "error",
            Message: "Invalid user ID",
        }
        json.NewEncoder(w).Encode(response)
        return
    }

    for _, user := range users {
        if user.ID == id {
            w.Header().Set("Content-Type", "application/json")
            response := APIResponse{
                Status: "success",
                Data:   user,
            }
            json.NewEncoder(w).Encode(response)
            return
        }
    }

    w.WriteHeader(http.StatusNotFound)
    response := APIResponse{
        Status:  "error",
        Message: "User not found",
    }
    json.NewEncoder(w).Encode(response)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        response := APIResponse{
            Status:  "error",
            Message: "Invalid JSON",
        }
        json.NewEncoder(w).Encode(response)
        return
    }

    user.ID = len(users) + 1
    user.CreatedAt = time.Now()
    users = append(users, user)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    response := APIResponse{
        Status: "success",
        Data:   user,
    }
    json.NewEncoder(w).Encode(response)
}

func main() {
    r := mux.NewRouter()
    
    // Middleware
    r.Use(loggingMiddleware)
    r.Use(metricsMiddleware)
    
    // Routes
    r.HandleFunc("/health", healthHandler).Methods("GET")
    r.HandleFunc("/users", getUsersHandler).Methods("GET")
    r.HandleFunc("/users/{id:[0-9]+}", getUserHandler).Methods("GET")
    r.HandleFunc("/users", createUserHandler).Methods("POST")
    r.Handle("/metrics", promhttp.Handler())

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
