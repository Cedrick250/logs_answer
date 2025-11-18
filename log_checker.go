
package main

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: log-archive <log-directory>")
        os.Exit(1)
    }

    logDir := os.Args[1]

    // ✅ Hardcode archive directory to vboxuser's Desktop
    archiveDir := "/home/vboxuser/Desktop/log_archives"
    logFile := filepath.Join(archiveDir, "archive_log.txt")

    // Ensure archive directory exists
    if err := os.MkdirAll(archiveDir, 0755); err != nil {
        fmt.Printf("Error creating archive directory: %v\n", err)
        os.Exit(1)
    }

    // Timestamp for archive name
    timestamp := time.Now().Format("20060102_150405")
    archiveName := fmt.Sprintf("logs_archive_%s.tar.gz", timestamp)
    archivePath := filepath.Join(archiveDir, archiveName)

    // Create archive file
    archiveFile, err := os.Create(archivePath)
    if err != nil {
        fmt.Printf("Error creating archive file: %v\n", err)
        os.Exit(1)
    }

    // Setup gzip writer
    gw := gzip.NewWriter(archiveFile)

    // Setup tar writer
    tw := tar.NewWriter(gw)

    // Walk through log directory
    err = filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Skip directories and non-regular files (symlinks, sockets, devices)
        if !info.Mode().IsRegular() {
            return nil
        }

        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()

        // Create tar header
        header, err := tar.FileInfoHeader(info, "")
        if err != nil {
            return err
        }

        // Preserve relative path inside archive
        relPath, err := filepath.Rel(logDir, path)
        if err != nil {
            return err
        }
        header.Name = relPath
        header.Size = info.Size()

        // Write header
        if err := tw.WriteHeader(header); err != nil {
            return err
        }

        // Copy file data
        if _, err := io.Copy(tw, file); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        fmt.Printf("Error archiving logs: %v\n", err)
        os.Exit(1)
    }

    // Explicitly close writers to flush buffers
    if err := tw.Close(); err != nil {
        fmt.Printf("Error closing tar writer: %v\n", err)
        os.Exit(1)
    }
    if err := gw.Close(); err != nil {
        fmt.Printf("Error closing gzip writer: %v\n", err)
        os.Exit(1)
    }
    if err := archiveFile.Close(); err != nil {
        fmt.Printf("Error closing archive file: %v\n", err)
        os.Exit(1)
    }

    // Record archive event
    entry := fmt.Sprintf("[%s] Archived %s -> %s\n",
        time.Now().Format(time.RFC3339), logDir, archivePath)

    f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Error writing log file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()

    if _, err := f.WriteString(entry); err != nil {
        fmt.Printf("Error writing log entry: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✅ Archive created: %s\n", archivePath)
    fmt.Printf("📄 Log updated: %s\n", logFile)
}
