package main

import (
        "bytes"
        "database/sql"
        "encoding/json"
        "fmt"
        "io"

        // "encoding/json"
        "log"
        "net/http"

        // "os"

        // "github.com/go-chi/chi/v5"
        // "github.com/go-chi/chi/v5/middleware"
        // "github.com/go-chi/cors"
        _ "modernc.org/sqlite"
)

type Institution struct {
        ID   int    `json:"S. No."`
        Name string `json:"College Name"`
}
func autoseed(){
        res,err:= http.Get("https://gist.githubusercontent.com/rvsp/45a64b307193107cfe6e2d737aa98803/raw/501ee0ea30046ef77552c80aca923f35c39a8dd8/all_engg.json")
        if err != nil{
                log.Fatal(err)
        }
        defer res.Body.Close()
        if res.StatusCode != http.StatusOK{
                log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
        }
        body, err := io.ReadAll(res.Body)
        if err != nil{
                log.Fatal(err)
        }
        body = bytes.ReplaceAll(body, []byte("\n"), []byte(" "))
        log.Printf("First 500 chars: %s", string(body[:500]))
        var institutions []Institution
        if err := json.Unmarshal(body, &institutions); err != nil {
                log.Fatal("Error parsing JSON:", err)
        }

        fmt.Printf("Found %d institutions. Inserting into DB...", len(institutions))
}
var db *sql.DB
func main() {
        var err error
        db, err = sql.Open("sqlite", "./institutions.db")
        if err != nil {
                log.Fatal(err)
        }else{
                log.Println("Connected to database")
        }
        defer db.Close()
        autoseed()
}
