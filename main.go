package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// In-memory storage for zip files with mutex for concurrent access
type ZipStorage struct {
	zips  map[string][]byte
	mutex sync.RWMutex
}

var zipStorage = ZipStorage{
	zips: make(map[string][]byte),
}

func main() {
	// Start a background goroutine to clean up old zip files every hour
	go func() {
		for {
			cleanupOldZips(24 * time.Hour) // Cleanup zips older than 24 hours
			time.Sleep(1 * time.Hour)      // Run cleanup every hour
		}
	}()

	// Set up routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)

	// Start server
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Cleanup old zips from memory
func cleanupOldZips(maxAge time.Duration) {
	currentTime := time.Now()
	
	zipStorage.mutex.Lock()
	defer zipStorage.mutex.Unlock()
	
	for key := range zipStorage.zips {
		// Extract timestamp from key (assuming key format is "timestamp.zip")
		timestampStr := key[:len(key)-4] // Remove ".zip"
        var timestamp int64
        _, err := fmt.Sscanf(timestampStr, "%d", &timestamp)
        if err != nil {
            continue
        }
        
        zipTime := time.Unix(timestamp, 0)
		if currentTime.Sub(zipTime) > maxAge {
			delete(zipStorage.zips, key)
			log.Printf("Deleted old zip: %s", key)
		}
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>[ thefilezippr ]</title>
    <style>
        /* Base styles */
        body {
            font-family: Tahoma, Verdana, Arial, sans-serif;
            max-width: 1000px;
            margin: 0 auto;
            padding: 0;
            background-color: #fff;
            color: #333;
            font-size: 11px;
        }
        #header {
            background: linear-gradient(to right, #3b5998, #6d84b4);
            color: white;
            padding: 4px 8px;
            border-bottom: 1px solid #3b5998;
            font-size: 20px;
            margin-bottom: 0;
            height: 30px;
            line-height: 30px;
            text-align: left;
        }
        .navbar {
            background-color: #3b5998;
            color: white;
            padding: 4px 8px;
            text-align: right;
        }
        .navbar a {
            color: white;
            text-decoration: none;
            padding: 0 5px;
            font-size: 12px;
        }
        #main-container {
            display: flex;
            flex-direction: row;
            width: 100%;
        }
        #sidebar {
            width: 200px;
            padding: 10px;
            flex-shrink: 0;
        }
        #content {
            flex-grow: 1;
            padding: 10px;
        }
        .box {
            border: 1px solid #3b5998;
            margin-bottom: 10px;
            background-color: #f7f7f7;
            padding: 0;
        }
        .box-header {
            background-color: #3b5998;
            color: white;
            padding: 4px 8px;
            font-weight: bold;
        }
        .box-content {
            padding: 8px;
        }
        .menu-item {
            padding: 3px 0;
            border-bottom: 1px solid #ddd;
        }
        .menu-item a {
            color: #3b5998;
            text-decoration: none;
        }
        input[type="file"] {
            font-size: 11px;
            margin-bottom: 10px;
            max-width: 100%;
        }
        .btn {
            background-color: #3b5998;
            color: white;
            padding: 3px 8px;
            border: 1px solid #29447e;
            cursor: pointer;
            font-size: 11px;
        }
        .search {
            border: 1px solid #ccc;
            padding: 5px;
            margin-bottom: 10px;
        }
        .search input[type="text"] {
            width: 110px;
            border: 1px solid #3b5998;
            font-size: 11px;
            padding: 2px;
        }
        .search input[type="submit"] {
            background-color: #3b5998;
            color: white;
            border: none;
            font-size: 11px;
            padding: 2px 6px;
            cursor: pointer;
        }
        .ad-box {
            border: 1px dashed #ccc;
            padding: 5px;
            margin-top: 20px;
            background-color: #f7f7f7;
            text-align: center;
            height: 250px;
        }
        
        /* Mobile Menu Toggle */
        #mobile-menu-toggle {
            display: none;
            background-color: #3b5998;
            color: white;
            border: none;
            padding: 8px 12px;
            font-size: 14px;
            cursor: pointer;
            width: 100%;
            text-align: left;
        }
        
        /* Responsive styles */
        @media (max-width: 768px) {
            #main-container {
                flex-direction: column;
            }
            #sidebar {
                width: auto;
                padding: 5px;
                order: 2;
                display: none; /* Hide sidebar by default on mobile */
            }
            #sidebar.show {
                display: block;
            }
            #content {
                order: 3;
                padding: 5px;
            }
            .box {
                margin-bottom: 8px;
            }
            #mobile-menu-toggle {
                display: block;
                order: 1;
            }
            .navbar {
                text-align: center;
                padding: 2px 4px;
            }
            .navbar a {
                padding: 0 3px;
                font-size: 11px;
            }
            #header {
                text-align: center;
                height: auto;
                padding: 8px;
            }
            .ad-box {
                height: auto;
                padding: 8px;
            }
            input[type="file"] {
                width: 100%;
            }
        }
        
        /* Tablet styles */
        @media (min-width: 769px) and (max-width: 1024px) {
            #sidebar {
                width: 180px;
                padding: 8px;
            }
            #content {
                padding: 8px;
            }
        }
    </style>
</head>
<body>
    <div id="header">[ thefilezippr ]</div>
    <div class="navbar">
        <a href="#">home</a>
        <a href="#">search</a>
        <a href="#">about</a>
        <a href="#">faq</a>
        <a href="#">logout</a>
    </div>
    
    <button id="mobile-menu-toggle" onclick="toggleSidebar()">☰ Menu</button>
    
    <div id="main-container">
        <div id="sidebar">
            <div class="search">
                <form>
                    <div>quick search</div>
                    <input type="text" name="q">
                    <input type="submit" value="go">
                </form>
            </div>
            
            <div class="box">
                <div class="box-header">My Files</div>
                <div class="box-content">
                    <div class="menu-item"><a href="#">My Uploads</a></div>
                    <div class="menu-item"><a href="#">My Zips</a></div>
                    <div class="menu-item"><a href="#">Recent Files</a></div>
                    <div class="menu-item"><a href="#">Shared Files</a></div>
                    <div class="menu-item"><a href="#">My Account</a></div>
                    <div class="menu-item"><a href="#">My Privacy</a></div>
                </div>
            </div>
            
            <div class="ad-box">
                <h3>File Storage Pro</h3>
                <p>Get unlimited zip space!</p>
                <p>Only $19.99/month</p>
                <button class="btn">Learn More</button>
            </div>
        </div>
        
        <div id="content">
            <div class="box">
                <div class="box-header">File Zipper</div>
                <div class="box-content">
                    <form action="/upload" method="post" enctype="multipart/form-data">
                        <p>Select files to zip:</p>
                        <input type="file" name="files" multiple><br>
                        <input type="submit" value="Zip Files" class="btn">
                    </form>
                </div>
            </div>
            
            <div class="box">
                <div class="box-header">Recent Activity</div>
                <div class="box-content">
                    <p>No recent zip activity.</p>
                </div>
            </div>
            
            <div class="box">
                <div class="box-header">File Network</div>
                <div class="box-content">
                    <p>You have 0 friends using thefilezippr.</p>
                    <p>Invite your friends!</p>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        function toggleSidebar() {
            const sidebar = document.getElementById('sidebar');
            sidebar.classList.toggle('show');
        }
    </script>
</body>
</html>
`
	fmt.Fprint(w, html)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the files from the form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// Create a unique filename based on timestamp
	timestamp := time.Now().Unix()
	zipFilename := fmt.Sprintf("%d.zip", timestamp)

	// Create an in-memory buffer to hold the zip
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// Process all uploaded files
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Read the file data into memory
		fileData, err := io.ReadAll(file)
		file.Close() // Close immediately as we've read all data
		if err != nil {
			http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a file in the zip
		zipFile, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			http.Error(w, "Error creating zip entry: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the file data to the zip
		if _, err := zipFile.Write(fileData); err != nil {
			http.Error(w, "Error writing to zip: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Close the zip writer to flush all data
	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Error finalizing zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the zip in memory
	zipStorage.mutex.Lock()
	zipStorage.zips[zipFilename] = zipBuffer.Bytes()
	zipStorage.mutex.Unlock()

	// Redirect to the download page
	http.Redirect(w, r, "/download/"+zipFilename, http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the filename from the URL
	filename := filepath.Base(r.URL.Path)

	// Check if the zip exists in memory
	zipStorage.mutex.RLock()
	zipData, exists := zipStorage.zips[filename]
	zipStorage.mutex.RUnlock()

	if !exists {
		http.Error(w, "Zip file not found", http.StatusNotFound)
		return
	}

	// If direct download requested
	if r.URL.Query().Get("dl") == "1" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipData)))
		w.Write(zipData)
		return
	}

	// Set success HTML
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>[ thefilezippr ]</title>
    <style>
        /* Base styles */
        body {
            font-family: Tahoma, Verdana, Arial, sans-serif;
            max-width: 1000px;
            margin: 0 auto;
            padding: 0;
            background-color: #fff;
            color: #333;
            font-size: 11px;
        }
        #header {
            background: linear-gradient(to right, #3b5998, #6d84b4);
            color: white;
            padding: 4px 8px;
            border-bottom: 1px solid #3b5998;
            font-size: 20px;
            margin-bottom: 0;
            height: 30px;
            line-height: 30px;
            text-align: left;
        }
        .navbar {
            background-color: #3b5998;
            color: white;
            padding: 4px 8px;
            text-align: right;
        }
        .navbar a {
            color: white;
            text-decoration: none;
            padding: 0 5px;
            font-size: 12px;
        }
        .content {
            padding: 20px;
            text-align: center;
        }
        .box {
            border: 1px solid #3b5998;
            margin: 20px auto;
            width: 400px;
            max-width: 90%;
            background-color: #f7f7f7;
        }
        .box-header {
            background-color: #3b5998;
            color: white;
            padding: 4px 8px;
            font-weight: bold;
        }
        .box-content {
            padding: 20px;
        }
        .btn {
            background-color: #3b5998;
            color: white;
            padding: 3px 8px;
            border: 1px solid #29447e;
            cursor: pointer;
            font-size: 11px;
            text-decoration: none;
            display: inline-block;
            margin-top: 10px;
        }
        
        /* Responsive styles */
        @media (max-width: 768px) {
            .content {
                padding: 10px;
            }
            .box {
                width: 90%;
                margin: 10px auto;
            }
            .box-content {
                padding: 10px;
            }
            .navbar {
                text-align: center;
                padding: 2px 4px;
            }
            .navbar a {
                padding: 0 3px;
                font-size: 11px;
            }
            #header {
                text-align: center;
                height: auto;
                padding: 8px;
            }
        }
    </style>
</head>
<body>
    <div id="header">[ thefilezippr ]</div>
    <div class="navbar">
        <a href="/">home</a>
        <a href="#">search</a>
        <a href="#">about</a>
        <a href="#">faq</a>
        <a href="#">logout</a>
    </div>
    
    <div class="content">
        <div class="box">
            <div class="box-header">Your Files Are Ready!</div>
            <div class="box-content">
                <p>Your files have been successfully zipped.</p>
                <a href="/download/` + filename + `?dl=1" class="btn">Download Zip File</a>
                <p><a href="/">Back to Home</a></p>
            </div>
        </div>
    </div>
</body>
</html>
	`

	// Show the success page
	fmt.Fprint(w, html)
}