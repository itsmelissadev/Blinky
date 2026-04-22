package postgresql

import (
	"archive/tar"
	"archive/zip"
	"blinky/internal/pkg/logger"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"blinky/internal"
	"blinky/internal/pkg/pathutil"
)

type ProgressReader struct {
	io.Reader
	Total   int64
	Current int64
	Label   string
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	pr.print()
	return n, err
}

func (pr *ProgressReader) print() {
	percent := float64(pr.Current) / float64(pr.Total) * 100
	bars := int(percent / 5)

	prog := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)

	fmt.Printf("\r\033[K[POSTGRES/PROGRESS] %s: [%s] %.1f%%", pr.Label, prog, percent)
	if pr.Current >= pr.Total {
		fmt.Println()
	}
}

func (m *Manager) EnsureBinaries() error {
	binPath := m.getBinPath("postgres")
	if _, err := os.Stat(binPath); err == nil {
		return nil
	}

	logger.Warn("[POSTGRES/DOWNLOAD] Full binaries not found. Downloading PostgreSQL %s...", internal.AppPostgresVersion)

	url := getDownloadURL()
	if url == "" {
		return fmt.Errorf("unsupported platform for EDB binaries: %s %s", runtime.GOOS, runtime.GOARCH)
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("postgres-%s.archive", internal.AppPostgresVersion))
	if err := downloadFile(url, tmpFile); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	logger.Info("[POSTGRES/DOWNLOAD] Extracting full distribution to %s...", pathutil.GetPostgresPath())

	var err error
	if strings.HasSuffix(url, ".zip") {
		err = unzipWithProgress(tmpFile, pathutil.GetPostgresPath())
	} else {
		err = untar(tmpFile, pathutil.GetPostgresPath())
	}

	if err != nil {
		return err
	}

	logger.Success("[POSTGRES/DOWNLOAD] PostgreSQL full distribution ready")
	return nil
}

func getDownloadURL() string {
	version := "16.3-1"

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("https://get.enterprisedb.com/postgresql/postgresql-%s-windows-x64-binaries.zip", version)
	case "linux":
		if runtime.GOARCH == "amd64" {
			return fmt.Sprintf("https://get.enterprisedb.com/postgresql/postgresql-%s-linux-x64-binaries.tar.gz", version)
		}
	case "darwin":
		return fmt.Sprintf("https://get.enterprisedb.com/postgresql/postgresql-%s-osx-binaries.zip", version)
	}
	return ""
}

func downloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	pr := &ProgressReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
		Label:  "Downloading",
	}

	_, err = io.Copy(out, pr)
	return err
}

func unzipWithProgress(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	totalFiles := int64(len(r.File))
	var currentFile int64

	for _, f := range r.File {
		currentFile++

		percent := float64(currentFile) / float64(totalFiles) * 100
		bars := int(percent / 5)
		prog := strings.Repeat("█", bars) + strings.Repeat("░", 20-bars)
		fmt.Printf("\r\033[K[POSTGRES/PROGRESS] Extracting: [%s] %.1f%% (%d/%d)", prog, percent, currentFile, totalFiles)

		name := f.Name
		if after, ok := strings.CutPrefix(name, "pgsql/"); ok {
			name = after
		}
		if name == "" {
			continue
		}

		fpath := filepath.Join(dest, name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}

func untar(src string, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := header.Name
		if after, ok := strings.CutPrefix(name, "pgsql/"); ok {
			name = after
		}
		if name == "" {
			continue
		}

		target := filepath.Join(dest, name)
		fmt.Printf("\r\033[K[POSTGRES/PROGRESS] Extracting: %s", name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	fmt.Println()
	return nil
}
