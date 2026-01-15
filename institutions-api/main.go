package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	// "github.com/xuri/excelize/v2"
		lru "github.com/hashicorp/golang-lru/v2"
	_ "modernc.org/sqlite"
)

type Institution struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"country"` 
}

var (
	db *sql.DB
	cache *lru.Cache[string, []Institution]
)

var synonyms = map[string]string{
	// --- USA (Tech & Ivy) ---
	"mit":     "Massachusetts Institute of Technology",
	"caltech": "California Institute of Technology",
	"cmu":     "Carnegie Mellon University",
	"nyu":     "New York University",
	"ucla":    "University of California Los Angeles",
	"ucsd":    "University of California San Diego",
	"ucb":     "University of California Berkeley",
	"usc":     "University of Southern California",
	"unc":     "University of North Carolina",
	"uiuc":    "University of Illinois Urbana-Champaign",
	"gatech":  "Georgia Institute of Technology",
	"gt":      "Georgia Institute of Technology",
	"rit":     "Rochester Institute of Technology",
	"rpi":     "Rensselaer Polytechnic Institute",
	"tamu":    "Texas A&M University",
	"lsu":     "Louisiana State University",
	"asu":     "Arizona State University",
	"psu":     "Pennsylvania State University",
	"byu":     "Brigham Young University",
	"smu":     "Southern Methodist University",
	"neu":     "Northeastern University",
	"bu":      "Boston University",
	"bc":      "Boston College",
	"upenn":   "University of Pennsylvania",
	"penn":    "University of Pennsylvania",
	"washu":   "Washington University in St. Louis",
	"umbc":    "University of Maryland Baltimore County",
	"va tech": "Virginia Polytechnic Institute and State University",
	"vt":      "Virginia Polytechnic Institute and State University",
	"suny":    "State University of New York",
	"cuny":    "City University of New York",

	// --- INDIA (Common Prefixes) ---
	"iit":    "Indian Institute of Technology",
	"nit":    "National Institute of Technology",
	"iiit":   "Indian Institute of Information Technology",
	"bit":    "Birla Institute of Technology",
	"du":     "University of Delhi",
	"jnu":    "Jawaharlal Nehru University",
	"bhu":    "Banaras Hindu University",
	"amu":    "Aligarh Muslim University",
	"aiims":  "All India Institute of Medical Sciences",
	"isi":    "Indian Statistical Institute",
	"iisc":   "Indian Institute of Science",
	"iim":    "Indian Institute of Management",
	"vit":    "Vellore Institute of Technology",
	"srm":    "SRM Institute of Science and Technology",
	"manipal": "Manipal Academy of Higher Education",
	"lpu":    "Lovely Professional University",
	"ignou":  "Indira Gandhi National Open University",
	

	// --- UK & EUROPE ---
	"lse":    "London School of Economics",
	"ucl":    "University College London",
	"icl":    "Imperial College London",
	"oxbridge": "University of Oxford",
	"eth":    "ETH Zurich",
	"epfl":   "École Polytechnique Fédérale de Lausanne",
	"tum":    "Technical University of Munich",

	// --- ASIA / OCEANIA ---
	"nus":    "National University of Singapore",
	"ntu":    "Nanyang Technological University",
	"hku":    "University of Hong Kong",
	"hkust":  "Hong Kong University of Science and Technology",
	"anu":    "Australian National University",
	"unsw":   "University of New South Wales",
	"kaist":  "Korea Advanced Institute of Science and Technology",
	"snu":    "Seoul National University",
}


func main() {
	var err error
	db, err = sql.Open("sqlite", "./institutions.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		log.Println("⚠️ Failed to enable WAL mode:", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Compress(5))
	r.Use(httprate.Limit(
		100,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}))
	
	r.Get("/", documentationHandler)
	r.Get("/api/institutions", searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}
	log.Printf("🚀 Server running on port %s", port)
	http.ListenAndServe(":"+port, r)
}
func sanitizeQuery(input string) string {
		clean := strings.Map(func(r rune) rune {
				if unicode.IsGraphic(r) {
						return r
				}
				return -1
		}, input)
		clean = strings.ReplaceAll(clean, "\"", "\"\"")
		return strings.TrimSpace(clean)
}
func expandQuery(input string) string {
	lower := strings.ToLower(input)
	if expansion, ok := synonyms[lower]; ok {
		return fmt.Sprintf("\"%s\" OR \"%s\"", input, expansion)
	}
	return fmt.Sprintf("\"%s\"", input)
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
	
	query := r.URL.Query().Get("name")
	limitStr := r.URL.Query().Get("limit")
	cacheKey := fmt.Sprintf("institutions:%s:limit:%s", strings.ToLower(query), limitStr)

    if institutions, found := cache.Get(cacheKey); found {
        w.Header().Set("X-Cache", "HIT")
        w.Header().Set("Cache-Control", "public, max-age=3600")
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(institutions)
        return
    }
	if limitStr == "" {
		limitStr = "20"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query = sanitizeQuery(query)
	query = expandQuery(query)
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]Institution{})
		return
	}

	sqlQuery := fmt.Sprintf(`
		SELECT rowid, name, state
		FROM institutions
		WHERE institutions MATCH ?
		ORDER BY
			-- Priority 1: Exact Start Matches ("Cambridge..." is better than "The Cambridge...")
			(CASE WHEN name LIKE '%s%%' THEN 0 ELSE 1 END) ASC,

			-- Priority 2: Institution Hierarchy (University > College > School)
			(CASE 
				WHEN name LIKE '%%University%%' THEN 0 
				WHEN name LIKE '%%Institute of Technology%%' THEN 1
				WHEN name LIKE '%%Institute%%' THEN 2
				WHEN name LIKE '%%College%%' THEN 3
				ELSE 4 
			END) ASC,

			-- Priority 3: Shorter names are usually the main institution
			length(name) ASC,

			-- Priority 4: Standard Text Relevance (BM25)
			bm25(institutions, 10.0, 5.0) ASC
		LIMIT %d
	`, query, limit)

	rows, err := db.Query(sqlQuery, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var results []Institution
	for rows.Next() {
		var i Institution
		rows.Scan(&i.ID, &i.Name, &i.State)
		results = append(results, i)
	}

	if results == nil {
		results = []Institution{}
	}
	cache.Add(cacheKey, results)

    w.Header().Set("X-Cache", "MISS")
    w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
func documentationHandler(w http.ResponseWriter, r *http.Request){
	const apiDocsHTML = `
		<!DOCTYPE html>
		<html lang="en">
		<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Institution Search API Documentation</title>
				<style>
						:root {
								--bg-color: #f8f9fa;
								--text-color: #212529;
								--accent-color: #007bff;
								--code-bg: #e9ecef;
								--card-bg: #ffffff;
								--border-color: #dee2e6;
						}
						@media (prefers-color-scheme: dark) {
								:root {
										--bg-color: #121212;
										--text-color: #e0e0e0;
										--accent-color: #64b5f6;
										--code-bg: #1e1e1e;
										--card-bg: #1e1e1e;
										--border-color: #333;
								}
						}
						body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: var(--text-color); background: var(--bg-color); margin: 0; padding: 20px; }
						.container { max-width: 800px; margin: 0 auto; background: var(--card-bg); padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); border: 1px solid var(--border-color); }
						h1, h2, h3 { color: var(--text-color); }
						h1 { border-bottom: 2px solid var(--accent-color); padding-bottom: 10px; }
						.endpoint { background: var(--code-bg); padding: 10px 15px; border-radius: 5px; font-family: monospace; font-size: 1.1em; display: inline-block; border: 1px solid var(--border-color); }
						.method { color: #fff; background: #28a745; padding: 2px 8px; border-radius: 4px; margin-right: 10px; font-weight: bold; font-size: 0.9em; }
						code { background: var(--code-bg); padding: 2px 5px; border-radius: 4px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; font-size: 0.9em; }
						pre { background: var(--code-bg); padding: 15px; border-radius: 5px; overflow-x: auto; border: 1px solid var(--border-color); }
						table { width: 100%; border-collapse: collapse; margin: 20px 0; }
						th, td { text-align: left; padding: 12px; border-bottom: 1px solid var(--border-color); }
						th { background: var(--code-bg); }
						.badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.8em; font-weight: bold; }
						.required { background: #dc3545; color: white; }
						.optional { background: #6c757d; color: white; }
						.footer { margin-top: 40px; font-size: 0.9em; color: #6c757d; text-align: center; }
				</style>
		</head>
		<body>

		<div class="container">
				<h1>🏛️ Institution Search API</h1>
				<p>A high-performance JSON API to search for Universities and Higher Education Institutions. It supports full-text search, trigram matching, and synonym expansion (e.g., searching "MIT" finds "Massachusetts Institute of Technology").</p>

				<h2>Endpoint</h2>
				<div class="endpoint">
						<span class="method">GET</span> /api/institutions
				</div>

				<h3>Query Parameters</h3>
				<table>
						<thead>
								<tr>
										<th>Parameter</th>
										<th>Type</th>
										<th>Required</th>
										<th>Description</th>
								</tr>
						</thead>
						<tbody>
								<tr>
										<td><code>name</code></td>
										<td>String</td>
										<td><span class="badge required">Yes</span></td>
										<td>The search term. Supports partial matches and acronyms. Min length: 2 chars.</td>
								</tr>
								<tr>
										<td><code>limit</code></td>
										<td>Integer</td>
										<td><span class="badge optional">No</span></td>
										<td>Number of results to return. Default: <code>20</code>. Max: <code>100</code>.</td>
								</tr>
						</tbody>
				</table>

				<h3>Example Request</h3>
				<pre><code>GET /api/institutions?name=IIT&limit=3</code></pre>

				<h3>Example Response</h3>
				<pre><code>[
			{
				"id": 104,
				"name": "Indian Institute of Technology Bombay",
				"state": "Maharashtra"
			},
			{
				"id": 105,
				"name": "Indian Institute of Technology Delhi",
				"state": "Delhi"
			},
			{
				"id": 142,
				"name": "Indian Institute of Technology Madras",
				"state": "Tamil Nadu"
			}
		]</code></pre>

				<h3>Status Codes</h3>
				<ul>
						<li><code>200 OK</code> - Request successful.</li>
						<li><code>429 Too Many Requests</code> - Rate limit exceeded (100 req/min).</li>
						<li><code>500 Internal Server Error</code> - Server-side issue.</li>
				</ul>

				<div class="footer">
						API Version 1.0 &bull; Powered by Go & SQLite
				</div>
		</div>

		</body>
		</html>
		`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(apiDocsHTML))
}