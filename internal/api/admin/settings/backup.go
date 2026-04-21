package settings

import (
	"archive/zip"
	"blinky/internal/api"
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/pathutil"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	BackupDir = "system/backups"
)

type BackupInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *backupHandler) list(c *fiber.Ctx) error {
	if _, err := os.Stat(BackupDir); os.IsNotExist(err) {
		return api.SuccessList(c, []BackupInfo{}, api.ListMeta{Total: 0, Limit: 0, Offset: 0})
	}

	files, err := os.ReadDir(BackupDir)
	if err != nil {
		logger.Error("[SETTINGS/BACKUP] Failed to read backup directory: %v", err)
		return api.SendError(c, api.ErrBackupDirReadFailed)
	}

	backups := []BackupInfo{}
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".zip" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:      f.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return api.SuccessList(c, backups, api.ListMeta{
		Total:  len(backups),
		Limit:  len(backups),
		Offset: 0,
	})
}

func (h *backupHandler) create(c *fiber.Ctx) error {
	filename, err := CreateBackup(h.cfg)
	if err != nil {
		logger.Error("[SETTINGS/BACKUP] Backup failed: %v", err)
		return api.SendError(c, api.ErrBackupPgDumpFailed)
	}

	logger.Success("[SETTINGS/BACKUP] Backup created successfully: %s", filename)
	return api.Success(c, fiber.Map{
		"message":  api.SuccessBackupCreated.Message,
		"filename": filename,
	})
}

func CreateBackup(cfg *config.Config) (string, error) {
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	sqlFileName := fmt.Sprintf("temp_%s.sql", timestamp)
	sqlFile := filepath.Join(BackupDir, sqlFileName)
	zipFileName := fmt.Sprintf("backup_%s.zip", timestamp)
	zipFile := filepath.Join(BackupDir, zipFileName)

	pgDumpPath := getPgDumpPath(cfg)

	cmd := exec.Command(pgDumpPath,
		"-h", cfg.DBHost,
		"-p", cfg.DBPort,
		"-U", cfg.DBUser,
		"-f", sqlFile,
		cfg.DBName,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.DBPass))

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %v, Output: %s", err, string(output))
	}
	defer os.Remove(sqlFile)

	if err := zipFiles(zipFile, []string{sqlFile}); err != nil {
		return "", err
	}

	return zipFileName, nil
}

func getPgDumpPath(cfg *config.Config) string {
	if cfg.PostgresPath != "" {
		binFolder := pathutil.Join(cfg.PostgresPath, "bin")
		p := pathutil.Join(binFolder, "pg_dump.exe")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		p = pathutil.Join(binFolder, "pg_dump")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	if path, err := exec.LookPath("pg_dump"); err == nil {
		return path
	}

	return "pg_dump"
}

func (h *backupHandler) delete(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return api.SendError(c, api.ErrBackupMissingFilename)
	}

	if filepath.Ext(filename) != ".zip" || filepath.Base(filename) != filename {
		return api.SendError(c, api.ErrBackupInvalidFilename)
	}

	filePath := filepath.Join(BackupDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return api.SendError(c, api.ErrBackupNotFound)
	}

	if err := os.Remove(filePath); err != nil {
		logger.Error("[SETTINGS/BACKUP] Failed to delete backup file %s: %v", filename, err)
		return api.SendError(c, api.ErrBackupDeleteFailed)
	}

	logger.Info("[SETTINGS/BACKUP] Backup deleted: %s", filename)
	return api.Success(c, api.SuccessBackupDeleted)
}

func (h *backupHandler) download(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filepath.Ext(filename) != ".zip" || filepath.Base(filename) != filename {
		return api.SendError(c, api.ErrBackupInvalidFilename)
	}

	filePath := filepath.Join(BackupDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return api.SendError(c, api.ErrBackupNotFound)
	}

	return c.Download(filePath)
}

func (h *backupHandler) getConfig(c *fiber.Ctx) error {
	return api.Success(c, fiber.Map{
		"postgresPath": h.cfg.PostgresPath,
	})
}

func (h *backupHandler) updateConfig(c *fiber.Ctx) error {
	var body struct {
		PostgresPath string `json:"postgresPath"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	h.cfg.PostgresPath = body.PostgresPath

	if err := updateEnvVariable("POSTGRESQL_FOLDER_PATH", body.PostgresPath); err != nil {
		logger.Error("[SETTINGS/BACKUP] Failed to update .env: %v", err)
		return api.SendError(c, api.ErrEnvUpdateFailed)
	}

	logger.Success("[SETTINGS/BACKUP] Backup configuration updated")
	return api.Success(c, api.SuccessConfigUpdated)
}

func zipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
