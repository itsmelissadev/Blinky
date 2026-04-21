package system

import (
	"blinky/internal/api"
	"blinky/internal/pkg/pathutil"
	"os"

	"github.com/gofiber/fiber/v2"
)

func RegisterFileRoutes(router fiber.Router) {
	g := router.Group("/system/files")
	g.Get("/browse", browse)
}

func browse(c *fiber.Ctx) error {
	path := pathutil.Normalize(c.Query("path"))

	if path == "" || path == string(os.PathSeparator) {
		var roots []string

		if os.PathSeparator == '\\' {
			for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
				driveStr := string(drive) + ":\\"
				if _, err := os.Stat(driveStr); err == nil {
					roots = append(roots, string(drive)+":/")
				}
			}
		} else {
			roots = append(roots, "/")
		}

		if len(roots) > 0 {
			return api.Success(c, fiber.Map{
				"path":   "",
				"dirs":   roots,
				"parent": "",
			})
		}
	}

	entries, err := os.ReadDir(pathutil.FromSlash(path))
	if err != nil {
		return api.SendError(c, api.ErrBackupBrowseFailed)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return api.Success(c, fiber.Map{
		"path":   path,
		"dirs":   dirs,
		"parent": pathutil.GetParent(path),
	})
}
