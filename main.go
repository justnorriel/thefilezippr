package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	uploadsDir = "./uploads" // Default uploads directory
	zipsDir    = "./zips"    // Default zips directory
)

func main() {
	// Set uploads and zips directories from environment variables
	if dir := os.Getenv("UPLOADS_DIR"); dir != "" {
		uploadsDir = dir
	}
	if dir := os.Getenv("ZIPS_DIR"); dir != "" {
		zipsDir = dir
	}

	// Create uploads and zips directories with error handling
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	if err := os.MkdirAll(zipsDir, 0755); err != nil {
		log.Fatalf("Failed to create zips directory: %v", err)
	}

	// Start a background goroutine to clean up old files every hour
	go func() {
		for {
			cleanupOldFiles(uploadsDir, 24*time.Hour) // Cleanup files older than 24 hours
			cleanupOldFiles(zipsDir, 24*time.Hour)    // Cleanup files older than 24 hours
			time.Sleep(1 * time.Hour)                 // Run cleanup every hour
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

// Cleanup old files in a directory
func cleanupOldFiles(dir string, maxAge time.Duration) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dir, err)
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", file.Name(), err)
			continue
		}

		if time.Since(info.ModTime()) > maxAge {
			err := os.Remove(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Printf("Error deleting file %s: %v", file.Name(), err)
			} else {
				log.Printf("Deleted old file: %s", file.Name())
			}
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

	// Create a unique folder name based on timestamp
	timestamp := time.Now().Unix()
	uploadDir := filepath.Join(uploadsDir, fmt.Sprintf("%d", timestamp))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save all uploaded files
	filePaths := []string{}
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Create the destination file
		destPath := filepath.Join(uploadDir, fileHeader.Filename)
		dest, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dest.Close()

		// Copy the uploaded file to the destination file
		_, err = io.Copy(dest, file)
		if err != nil {
			http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filePaths = append(filePaths, destPath)
	}

	// Create the zip file
	zipFilename := fmt.Sprintf("%d.zip", timestamp)
	zipPath := filepath.Join(zipsDir, zipFilename)
	
	err = zipFiles(zipPath, filePaths)
	if err != nil {
		http.Error(w, "Error creating zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to the download page
	http.Redirect(w, r, "/download/"+zipFilename, http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the filename from the URL
	filename := filepath.Base(r.URL.Path)
	zipPath := filepath.Join(zipsDir, filename)

	// Check if the file exists
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		http.Error(w, "Zip file not found", http.StatusNotFound)
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

	// If direct download requested
	if r.URL.Query().Get("dl") == "1" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Type", "application/zip")
		http.ServeFile(w, r, zipPath)
		return
	}

	// Otherwise show the success page
	fmt.Fprint(w, html)
}

func zipFiles(zipPath string, filePaths []string) error {
	// Create a new zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add files to the zip
	for _, filePath := range filePaths {
		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get file info
		info, err := file.Stat()
		if err != nil {
			return err
		}

		// Create a zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set the name to just the filename without the full path
		header.Name = filepath.Base(filePath)

		// Create the file in the zip
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy the file to the zip
		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}