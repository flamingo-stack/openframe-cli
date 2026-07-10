package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

// FileBackup represents a file backup operation
type FileBackup struct {
	OriginalPath    string
	BackupPath      string
	ContentOnly     bool // If true, we only store content in memory
	OriginalContent []byte
	FileExisted     bool
}

// FileCleanup handles backup and restore of files modified by CLI
type FileCleanup struct {
	backups          []FileBackup
	cleanupOnSuccess bool // If true, clean up temporary files only on successful completion
}

// NewFileCleanup creates a new file cleanup manager
func NewFileCleanup() *FileCleanup {
	return &FileCleanup{
		backups:          make([]FileBackup, 0),
		cleanupOnSuccess: false, // By default, cleanup on any completion
	}
}

// RestoreFiles restores all backed up files to their original state
// This method is used for error conditions and interruptions, so it always cleans up
func (fc *FileCleanup) RestoreFiles(verbose bool) error {
	// Force cleanup regardless of success-only mode for error/interruption scenarios
	return fc.restoreFilesForced(verbose, false)
}

// RestoreFilesOnSuccess restores files only after successful completion
func (fc *FileCleanup) RestoreFilesOnSuccess(verbose bool) error {
	return fc.RestoreFilesWithResult(verbose, true)
}

// restoreFilesForced restores files ignoring the success-only setting (used for errors/interruptions)
func (fc *FileCleanup) restoreFilesForced(verbose bool, success bool) error {
	if len(fc.backups) == 0 {
		if verbose {
			pterm.Info.Println("No files to restore")
		}
		return nil
	}

	restoredCount := 0
	for _, backup := range fc.backups {
		if err := fc.restoreFile(backup, verbose); err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to restore %s: %v\n", backup.OriginalPath, err)
			}
			continue
		}
		restoredCount++
	}

	if restoredCount > 0 && verbose {
		if success {
			pterm.Success.Printf("✅ Cleaned up %d temporary file(s) after successful operation\n", restoredCount)
		} else {
			pterm.Success.Printf("✅ Cleaned up %d file(s) after interruption\n", restoredCount)
		}
	}

	// Clean up physical backup files
	fc.cleanupBackupFiles(verbose)

	return nil
}

// RestoreFilesWithResult restores files based on success state
func (fc *FileCleanup) RestoreFilesWithResult(verbose bool, success bool) error {
	if len(fc.backups) == 0 {
		if verbose {
			pterm.Info.Println("No files to restore")
		}
		return nil
	}

	restoredCount := 0
	for _, backup := range fc.backups {
		// For temporary files registered for success-only cleanup
		if fc.cleanupOnSuccess && !backup.FileExisted && !success {
			if verbose {
				pterm.Info.Printf("⏳ Keeping temporary file until successful completion: %s\n", backup.OriginalPath)
			}
			continue // Skip cleanup of temp files unless successful
		}

		if err := fc.restoreFile(backup, verbose); err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to restore %s: %v\n", backup.OriginalPath, err)
			}
			continue
		}
		restoredCount++
	}

	if restoredCount > 0 && verbose {
		if success {
			pterm.Success.Printf("✅ Cleaned up %d temporary file(s) after successful operation\n", restoredCount)
		} else {
			pterm.Success.Printf("✅ Cleaned up %d file(s) after failed operation\n", restoredCount)
		}
	}

	// Clean up physical backup files
	fc.cleanupBackupFiles(verbose)

	return nil
}

// restoreFile restores a single file from backup
func (fc *FileCleanup) restoreFile(backup FileBackup, verbose bool) error {
	if backup.FileExisted {
		if backup.ContentOnly {
			// Restore from memory
			if err := os.WriteFile(backup.OriginalPath, backup.OriginalContent, 0600); err != nil {
				return fmt.Errorf("failed to restore file from memory: %w", err)
			}
		} else {
			// Restore from backup file
			if err := fc.copyFile(backup.BackupPath, backup.OriginalPath); err != nil {
				return fmt.Errorf("failed to restore from backup file: %w", err)
			}
		}
		if verbose {
			pterm.Success.Printf("✓ Restored original: %s\n", backup.OriginalPath)
		}
	} else {
		// File didn't exist originally, remove it
		if err := os.Remove(backup.OriginalPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove file that didn't exist originally: %w", err)
		}
		if verbose {
			pterm.Success.Printf("✓ Removed generated file: %s\n", backup.OriginalPath)
		}
	}

	return nil
}

// cleanupBackupFiles removes physical backup files
func (fc *FileCleanup) cleanupBackupFiles(verbose bool) {
	for _, backup := range fc.backups {
		if !backup.ContentOnly && backup.BackupPath != "" {
			if err := os.Remove(backup.BackupPath); err != nil && !os.IsNotExist(err) {
				if verbose {
					pterm.Warning.Printf("Failed to remove backup file %s: %v\n", backup.BackupPath, err)
				}
			}
		}
	}
}

// RegisterTempFile registers a temporary file for cleanup (will be deleted on restore)
func (fc *FileCleanup) RegisterTempFile(filePath string) error {
	// Register a temporary file that we know didn't exist before
	// No need to back it up since we created it - just mark it for deletion
	backup := FileBackup{
		OriginalPath:    filePath,
		BackupPath:      "",    // No backup needed
		ContentOnly:     false, // Not storing content
		OriginalContent: nil,   // No original content
		FileExisted:     false, // File didn't exist originally
	}

	fc.backups = append(fc.backups, backup)
	return nil
}

// SetCleanupOnSuccessOnly configures temporary files to be cleaned only on successful completion
func (fc *FileCleanup) SetCleanupOnSuccessOnly(enabled bool) {
	fc.cleanupOnSuccess = enabled
}

// copyFile copies a file from src to dst
func (fc *FileCleanup) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src) // #nosec G304 -- copies a program-tracked backup file
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	destFile, err := os.Create(dst) // #nosec G304 -- restores to a program-tracked backup path
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
