<!-- source-hash: a3afb49ad59313334bccf9d0b57753cd -->
This file contains comprehensive test cases for the FileCleanup system, validating temporary file management, backup/restore operations, and cleanup behavior under different scenarios.

## Key Components

- **Temporary File Tests**: `TestRegisterTempFile_NoBackupCreated`, `TestRegisterTempFile_MultipleFiles` - Verify temp files are registered and cleaned without creating backups
- **Cleanup Mode Tests**: `TestFileCleanup_CleanupOnSuccessOnly_Behavior`, `TestFileCleanup_RegularMode_AlwaysCleanup` - Test conditional vs unconditional cleanup
- **Signal Handling Test**: `TestFileCleanup_SignalInterruption_Scenario` - Validates behavior during CTRL-C interruption
- **Backup Tests**: `TestFileCleanup_BackupFile_MemoryOnly`, `TestFileCleanup_BackupFile_PhysicalBackup` - Test both memory-based and disk-based backup strategies

## Usage Example

```go
func TestMyCleanupScenario(t *testing.T) {
    cleanup := NewFileCleanup()
    
    // Create and register a temporary file
    tmpDir := t.TempDir()
    tempFile := filepath.Join(tmpDir, "temp.yaml")
    os.WriteFile(tempFile, []byte("content"), 0644)
    
    cleanup.RegisterTempFile(tempFile)
    
    // Test cleanup behavior
    cleanup.RestoreFiles(false)
    assert.NoFileExists(t, tempFile)
}
```

The tests ensure robust file management across success/failure scenarios, interruption handling, and proper backup/restore functionality.