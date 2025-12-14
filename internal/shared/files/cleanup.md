<!-- source-hash: 1ba729d3e7f90810d0838b31d866f6a9 -->
This package provides file backup and cleanup utilities for managing temporary files and ensuring proper restoration of original files during CLI operations.

## Key Components

**Core Types:**
- `FileBackup`: Represents a backup operation with original path, backup location, and content storage options
- `FileCleanup`: Main manager for handling file backups and restoration

**Primary Methods:**
- `NewFileCleanup()`: Creates a new cleanup manager instance
- `BackupFile(filePath, useMemoryOnly)`: Creates backup before file modification (disk or memory-based)
- `RestoreFiles(verbose)`: Restores all backed up files (for error scenarios)
- `RestoreFilesOnSuccess(verbose)`: Restores files only after successful operations
- `RegisterTempFile(filePath)`: Registers temporary files for automatic cleanup

**Utility Functions:**
- `AddCLIMarkerToFile(filePath, content)`: Adds generation markers to files
- `GetSafeFileName(baseName)`: Creates timestamped temporary filenames

## Usage Example

```go
// Create cleanup manager
cleanup := NewFileCleanup()

// Backup existing file before modification
err := cleanup.BackupFile("config.yaml", false) // Use disk backup
if err != nil {
    log.Fatal(err)
}

// Register new temporary file
cleanup.RegisterTempFile("temp-script.sh")

// On success, clean up temporary files
defer cleanup.RestoreFilesOnSuccess(true)

// On error, restore everything
defer func() {
    if r := recover(); r != nil {
        cleanup.RestoreFiles(true)
    }
}()
```