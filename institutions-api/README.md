# Institution Search API

A high-performance REST API built with Go for searching educational institutions by name. Features full-text search with SQLite FTS5, synonym expansion, BM25 ranking, and trigram tokenization for fuzzy matching.

## Features

- 🔍 **Full-Text Search** - Powered by SQLite FTS5 with trigram tokenization
- 🎯 **BM25 Ranking** - Intelligent relevance scoring for search results
- 📚 **Synonym Support** - Automatic expansion of common abbreviations (MIT → Massachusetts Institute of Technology)
- 🌍 **Global Coverage** - Comprehensive database of universities and colleges worldwide
- ⚡ **High Performance** - Optimized with WAL mode and connection pooling
- 🔒 **Rate Limiting** - 100 requests per minute per IP
- 🌐 **CORS Enabled** - Ready for cross-origin requests
- 📖 **Interactive Docs** - Built-in HTML documentation interface


## Installation

### Prerequisites

- Go 1.24 or higher
- SQLite database file (`institutions.db`)

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd Go
```

2. Install dependencies:
```bash
go mod download
```

3. Ensure you have the `institutions.db` file in the project root

4. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### 🏠 Documentation

**GET** `/`

Returns an interactive HTML documentation page with API details and examples.

**Example:**
```bash
curl http://localhost:8080/
```

### 🔍 Search Institutions

**GET** `/api/institutions`

Search for educational institutions by name with support for synonyms, abbreviations, and fuzzy matching.

**Query Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Search query (minimum 2 characters) |
| `limit` | integer | No | 20 | Number of results (max: 100) |

**Example Requests:**
```bash
# Search for MIT
curl "http://localhost:8080/api/institutions?name=MIT&limit=5"

# Search for IIT institutions
curl "http://localhost:8080/api/institutions?name=IIT&limit=10"

# Partial match search
curl "http://localhost:8080/api/institutions?name=University%20of%20California"
```

**Response:**
```json
[
  {
    "id": 104,
    "name": "Indian Institute of Technology Bombay",
    "state": "Maharashtra"
  },
  {
    "id": 105,
    "name": "Indian Institute of Technology Delhi",
    "state": "Delhi"
  }
]
```

**Status Codes:**
- `200 OK` - Request successful
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### 💓 Health Check

**GET** `/health`

Health check endpoint for monitoring and load balancers.

**Example:**
```bash
curl http://localhost:8080/health
```

## Synonym Expansion

The API automatically expands common university abbreviations to their full names. When you search for an abbreviation, the API searches for both the abbreviation and the full name.

### Supported Abbreviations

#### 🇺🇸 United States
- **MIT** → Massachusetts Institute of Technology
- **Caltech** → California Institute of Technology
- **CMU** → Carnegie Mellon University
- **NYU** → New York University
- **UCLA** → University of California Los Angeles
- **UCSD** → University of California San Diego
- **UCB** → University of California Berkeley
- **USC** → University of Southern California
- **UNC** → University of North Carolina
- **UIUC** → University of Illinois Urbana-Champaign
- **Georgia Tech / GT** → Georgia Institute of Technology
- **RIT** → Rochester Institute of Technology
- **RPI** → Rensselaer Polytechnic Institute
- **TAMU** → Texas A&M University
- **LSU** → Louisiana State University
- **ASU** → Arizona State University
- **PSU** → Pennsylvania State University
- **BYU** → Brigham Young University
- **SMU** → Southern Methodist University
- **NEU** → Northeastern University
- **BU** → Boston University
- **BC** → Boston College
- **UPenn / Penn** → University of Pennsylvania
- **WashU** → Washington University in St. Louis
- **UMBC** → University of Maryland Baltimore County
- **VA Tech / VT** → Virginia Polytechnic Institute and State University
- **SUNY** → State University of New York
- **CUNY** → City University of New York

#### 🇮🇳 India
- **IIT** → Indian Institute of Technology
- **NIT** → National Institute of Technology
- **IIIT** → Indian Institute of Information Technology
- **BITS** → Birla Institute of Technology
- **DU** → University of Delhi
- **JNU** → Jawaharlal Nehru University
- **BHU** → Banaras Hindu University
- **AMU** → Aligarh Muslim University
- **AIIMS** → All India Institute of Medical Sciences
- **ISI** → Indian Statistical Institute
- **IISc** → Indian Institute of Science
- **IIM** → Indian Institute of Management
- **VIT** → Vellore Institute of Technology
- **SRM** → SRM Institute of Science and Technology
- **Manipal** → Manipal Academy of Higher Education
- **LPU** → Lovely Professional University
- **IGNOU** → Indira Gandhi National Open University

#### 🇬🇧🇪🇺 UK & Europe
- **LSE** → London School of Economics
- **UCL** → University College London
- **ICL** → Imperial College London
- **Oxbridge** → University of Oxford
- **ETH** → ETH Zurich
- **EPFL** → École Polytechnique Fédérale de Lausanne
- **TUM** → Technical University of Munich

#### 🌏 Asia & Oceania
- **NUS** → National University of Singapore
- **NTU** → Nanyang Technological University
- **HKU** → University of Hong Kong
- **HKUST** → Hong Kong University of Science and Technology
- **ANU** → Australian National University
- **UNSW** → University of New South Wales
- **KAIST** → Korea Advanced Institute of Science and Technology
- **SNU** → Seoul National University

## How It Works

### Search Algorithm

1. **Query Sanitization** - Removes non-graphic characters and escapes quotes
2. **Synonym Expansion** - Expands known abbreviations to full names
3. **FTS5 Matching** - Uses SQLite's full-text search with trigram tokenization
4. **BM25 Ranking** - Ranks results by relevance with custom weights (10.0 for name, 5.0 for state)
5. **Length Sorting** - Secondary sort by name length for similar relevance scores
6. **Result Limiting** - Returns top N results (configurable, max 100)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please open an issue on GitHub.

---


