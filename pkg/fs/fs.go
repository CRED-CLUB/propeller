package fs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem interface
type FileSystem interface {
	CreateDir(dirPath string, permission os.FileMode, forceCreate bool) error
	DeleteDir(dirPath string) error
	CopyDir(srcPath, destPath string) error

	CreateFile(filePath string, forceCreate bool) error
	WriteFile(filePath string, permission os.FileMode, contents []byte) error
	ReadFile(filePath string) ([]byte, error)
	DeleteFile(filePath string) error
	CopyFile(sourcePath, destPath string) error

	SearchFiles(sourceDir string, excludeDirs []string, searchFile string) ([]string, error)
	SearchFileExtensions(sourceDir string, excludeDirs []string, extension string) ([]string, error)

	CreateSymLink(sourcePath string, destPath string) error

	Pwd() (string, error)
	Exists(filePath string) (bool, error)
	Cd(dir string) error
}

// LocalFileSystem ...
type LocalFileSystem struct {
}

// CreateDir ...
func (l *LocalFileSystem) CreateDir(dirPath string, permission os.FileMode, forceCreate bool) error {
	_, err := os.Stat(dirPath)
	if !forceCreate && !os.IsNotExist(err) {
		err = fmt.Errorf("error in creating directory %s", dirPath)
		return err
	}

	err = os.MkdirAll(dirPath, permission)
	if err != nil {
		err = fmt.Errorf("error in creating all directories %s", dirPath)
		return err
	}

	return os.Chmod(dirPath, 0755)
}

// DeleteDir ...
func (l *LocalFileSystem) DeleteDir(dirPath string) error {
	return os.RemoveAll(dirPath)
}

// CopyDir ...
func (l *LocalFileSystem) CopyDir(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source path : %s is not a directory", src)
	}

	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		destPath := filepath.Join(dest, path[len(src)+1:])
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(srcFile *os.File) {
			err := srcFile.Close()
			if err != nil {

			}
		}(srcFile)

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer func(destFile *os.File) {
			err := destFile.Close()
			if err != nil {

			}
		}(destFile)

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// CreateFile ...
func (l *LocalFileSystem) CreateFile(filePath string, forceCreate bool) error {
	_, err := os.Stat(filePath)
	if !forceCreate && os.IsExist(err) {
		return fmt.Errorf("error in stating file %s because of %v", filePath, err)
	}
	_, err = os.Create(filePath)
	return err
}

// WriteFile ...
func (l *LocalFileSystem) WriteFile(filePath string, permission os.FileMode, contents []byte) error {
	return os.WriteFile(filePath, contents, permission)
}

// DeleteFile ...
func (l *LocalFileSystem) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

// CopyFile ...
func (l *LocalFileSystem) CopyFile(sourcePath, destPath string) error {
	input, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(destPath, input, 0644)
}

// ReadFile ...
func (l *LocalFileSystem) ReadFile(filePath string) ([]byte, error) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return contents, err
}

// SearchFiles ...
func (l *LocalFileSystem) SearchFiles(sourceDir string, excludeDirs []string, searchFile string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(sourceDir, fs.WalkDirFunc(func(path string, ds fs.DirEntry, err error) error {
		if !ds.IsDir() && !isAncestor(excludeDirs, path) && filepath.Base(path) == searchFile {
			files = append(files, path)
		}
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// SearchFileExtensions ...
func (l *LocalFileSystem) SearchFileExtensions(sourceDir string, excludeDirs []string, extension string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(sourceDir, fs.WalkDirFunc(func(path string, ds fs.DirEntry, err error) error {
		if !ds.IsDir() && !isAncestor(excludeDirs, path) && filepath.Ext(path) == extension {
			files = append(files, path)
		}
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return files, nil
}

// CreateSymLink ...
func (l *LocalFileSystem) CreateSymLink(sourcePath, destPath string) error {
	return os.Symlink(sourcePath, destPath)
}

// Pwd ...
func (l *LocalFileSystem) Pwd() (string, error) {
	return os.Getwd()
}

// Exists ...
func (l *LocalFileSystem) Exists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Cd ...
func (l *LocalFileSystem) Cd(dir string) error {
	return os.Chdir(dir)
}

// isAncestor
func isAncestor(excludeDirs []string, path string) bool {
	for _, excludeDir := range excludeDirs {
		if strings.Contains(path, excludeDir) {
			return true
		}
	}
	return false
}
