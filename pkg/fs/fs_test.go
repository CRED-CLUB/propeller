package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileSystem_CreateDir(t *testing.T) {
	fs := &LocalFileSystem{}

	// Create temp directory for tests
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		dirPath     string
		permission  os.FileMode
		forceCreate bool
		wantErr     bool
	}{
		{
			name:        "create new directory",
			dirPath:     filepath.Join(tempDir, "newdir"),
			permission:  0755,
			forceCreate: false,
			wantErr:     false,
		},
		{
			name:        "create existing directory without force",
			dirPath:     filepath.Join(tempDir, "existingdir"),
			permission:  0755,
			forceCreate: false,
			wantErr:     true,
		},
		{
			name:        "force create existing directory",
			dirPath:     filepath.Join(tempDir, "forcedir"),
			permission:  0755,
			forceCreate: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pre-create directory for "existing" test cases
			if tt.name != "create new directory" {
				err := os.MkdirAll(tt.dirPath, 0755)
				if err != nil {
					t.Fatalf("failed to setup test: %v", err)
				}
			}

			err := fs.CreateDir(tt.dirPath, tt.permission, tt.forceCreate)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we don't expect an error, verify the directory exists
			if !tt.wantErr {
				info, err := os.Stat(tt.dirPath)
				if err != nil {
					t.Errorf("directory was not created: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error("created path is not a directory")
				}
				if info.Mode().Perm() != 0755 {
					t.Errorf("directory has wrong permissions: got %v, want %v", info.Mode().Perm(), 0755)
				}
			}
		})
	}
}

func TestLocalFileSystem_DeleteDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_delete_dir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Create a test file inside the temporary directory
	testFilePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create an instance of LocalFileSystem
	fs := &LocalFileSystem{}

	tests := []struct {
		name    string
		dirPath string
		wantErr bool
	}{
		{
			name:    "Delete existing directory",
			dirPath: tempDir,
			wantErr: false,
		},
		{
			name:    "Delete non-existing directory",
			dirPath: filepath.Join(tempDir, "non_existing"),
			wantErr: false, // RemoveAll doesn't return error for non-existing paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.DeleteDir(tt.dirPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LocalFileSystem.DeleteDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the directory no longer exists
			if _, err := os.Stat(tt.dirPath); !os.IsNotExist(err) {
				t.Errorf("Directory still exists after deletion: %v", tt.dirPath)
			}
		})
	}
}

func TestLocalFileSystem_CreateAndDeleteFile(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Test CreateFile
	err := fs.CreateFile(testFile, false)
	if err != nil {
		t.Errorf("CreateFile() error = %v", err)
	}

	// Test file exists
	exists, err := fs.Exists(testFile)
	if err != nil || !exists {
		t.Errorf("File should exist after creation")
	}

	// Test DeleteFile
	err = fs.DeleteFile(testFile)
	if err != nil {
		t.Errorf("DeleteFile() error = %v", err)
	}

	// Verify file is deleted
	exists, err = fs.Exists(testFile)
	if err != nil || exists {
		t.Errorf("File should not exist after deletion")
	}
}

func TestLocalFileSystem_WriteAndReadFile(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")

	// Test WriteFile
	err := fs.WriteFile(testFile, 0644, content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// Test ReadFile
	readContent, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile() error = %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("ReadFile() got = %v, want %v", string(readContent), string(content))
	}
}

func TestLocalFileSystem_CopyFile(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "src.txt")
	destFile := filepath.Join(tempDir, "dest.txt")
	content := []byte("test content")

	// Create and write to source file
	err := fs.WriteFile(srcFile, 0644, content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// Test CopyFile
	err = fs.CopyFile(srcFile, destFile)
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
	}

	// Verify content
	destContent, err := fs.ReadFile(destFile)
	if err != nil {
		t.Errorf("ReadFile() error = %v", err)
	}
	if string(destContent) != string(content) {
		t.Errorf("Copied file content mismatch, got = %v, want %v", string(destContent), string(content))
	}
}

func TestLocalFileSystem_CopyDir(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destDir := filepath.Join(tempDir, "dest")

	// Create source directory with files
	err := fs.CreateDir(srcDir, 0755, false)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	err = fs.WriteFile(filepath.Join(srcDir, "test.txt"), 0644, []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test CopyDir
	err = fs.CopyDir(srcDir, destDir)
	if err != nil {
		t.Errorf("CopyDir() error = %v", err)
	}

	// Verify copied file exists
	exists, err := fs.Exists(filepath.Join(destDir, "test.txt"))
	if err != nil || !exists {
		t.Errorf("Copied file should exist in destination directory")
	}
}

func TestLocalFileSystem_SearchFiles(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()

	// Create test directory structure
	testFiles := []string{
		filepath.Join(tempDir, "test1.txt"),
		filepath.Join(tempDir, "subdir", "test2.txt"),
	}

	err := fs.CreateDir(filepath.Join(tempDir, "subdir"), 0755, false)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	for _, file := range testFiles {
		err := fs.WriteFile(file, 0644, []byte("test"))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test SearchFiles
	found, err := fs.SearchFiles(tempDir, []string{}, "test1.txt")
	if err != nil {
		t.Errorf("SearchFiles() error = %v", err)
	}
	if len(found) != 1 {
		t.Errorf("SearchFiles() found %d files, want 1", len(found))
	}
}

func TestLocalFileSystem_SearchFileExtensions(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()

	// Create test files with different extensions
	files := []string{
		filepath.Join(tempDir, "test1.txt"),
		filepath.Join(tempDir, "test2.txt"),
		filepath.Join(tempDir, "test3.doc"),
	}

	for _, file := range files {
		err := fs.WriteFile(file, 0644, []byte("test"))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test SearchFileExtensions
	found, err := fs.SearchFileExtensions(tempDir, []string{}, ".txt")
	if err != nil {
		t.Errorf("SearchFileExtensions() error = %v", err)
	}
	if len(found) != 2 {
		t.Errorf("SearchFileExtensions() found %d files, want 2", len(found))
	}
}

func TestLocalFileSystem_CreateSymLink(t *testing.T) {
	fs := &LocalFileSystem{}
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "src.txt")
	linkFile := filepath.Join(tempDir, "link.txt")

	// Create source file
	err := fs.WriteFile(srcFile, 0644, []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test CreateSymLink
	err = fs.CreateSymLink(srcFile, linkFile)
	if err != nil {
		t.Errorf("CreateSymLink() error = %v", err)
	}

	// Verify symlink exists
	linkInfo, err := os.Lstat(linkFile)
	if err != nil {
		t.Errorf("Failed to stat symlink: %v", err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Error("Created file is not a symlink")
	}
}
