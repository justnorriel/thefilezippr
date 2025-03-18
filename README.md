# thefilezippr - A Go-based In-Memory File Zipping Tool

`thefilezippr` is a simple web application built in Go that allows users to upload multiple files, zip them, and download the resulting `.zip` file. It is designed to be lightweight, easy to use, and deployable on modern hosting platforms like Vercel with no file system dependencies.

---

## Features

- **File Upload**: Users can upload multiple files simultaneously.
- **Zip Creation**: Uploaded files are automatically zipped into a single `.zip` file.
- **File Download**: Users can download the zipped file directly from the browser.
- **In-Memory Processing**: All files are processed in memory without saving to disk, making it ideal for serverless environments.
- **Automatic Cleanup**: Zipped archives are stored temporarily in memory and cleaned up periodically.
- **Responsive Design**: The web interface is mobile-friendly and works on all devices.

---

## Prerequisites

Before running the application, ensure you have the following installed:

- **Go** (version 1.20 or higher): [Download Go](https://golang.org/dl/)
- **Git**: [Download Git](https://git-scm.com/)

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/thefilezippr.git
cd thefilezippr
```

### 2. Build the Application

Run the following command to build the Go application:

```bash
go build -o thefilezippr
```

### 3. Run the Application

Start the server:

```bash
./thefilezippr
```

The application will start on [http://localhost:8080](http://localhost:8080).

---

## Configuration

The application uses the following environment variables for configuration:

| Variable      | Default Value | Description                       |
|----------------|---------------|-----------------------------------|
| `PORT`        | `8080`        | Port on which the server will run.|

To set environment variables, create a `.env` file in the root directory:

```
PORT=8080
```

---

## Deployment on Vercel

This application is designed to work perfectly with serverless platforms like Vercel:

1. Fork or clone this repository
2. Connect your GitHub repository to Vercel
3. Deploy as a Go application
4. No additional configuration is needed as the app doesn't require file system access

---

## Memory Management

The application stores all zip files in memory and automatically removes files older than 24 hours. This cleanup process runs every hour in the background to prevent memory leaks.

---

## Technical Details

- Uses Go's built-in `archive/zip` package for zip file creation
- Implements thread-safe in-memory storage with mutex locks for concurrent access
- All file operations happen in memory without writing to disk
- Responsive UI with mobile-first design principles

---

## Contributing

Contributions are welcome! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Commit your changes and push to your fork.
4. Submit a pull request.

---

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

## Acknowledgments

Built with ❤️ using Go.

Inspired by the need for a simple file zipping tool for serverless environments.
