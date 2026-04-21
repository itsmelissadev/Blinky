package system

import (
	"blinky/internal/api"
	"blinky/internal/api/public"
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	publicApp     *fiber.App
	publicAppLock sync.Mutex
	dbPool        *pgxpool.Pool
	cfg           *config.Config
	isRunning     bool
)

func RegisterRoutes(router fiber.Router, db *pgxpool.Pool, c *config.Config) {
	dbPool = db
	cfg = c
	isRunning = true

	g := router.Group("/system/engine")
	g.Post("/restart", restartEngine)
	g.Post("/stop", stopEngine)
	g.Post("/start", startEngine)
}

func RegisterSetupRoutes(router fiber.Router, c *config.Config) {
	cfg = c
	g := router.Group("/setup")
	g.Get("/status", getSetupStatus)
	g.Post("/test-db", testDBConnection)
	g.Post("/env", saveSetupEnv)
}

func getSetupStatus(c *fiber.Ctx) error {
	return api.Success(c, fiber.Map{
		"is_env_exist": cfg.IsEnvExist,
	})
}

func testDBConnection(c *fiber.Ctx) error {
	type DBReq struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	var req DBReq
	if err := c.BodyParser(&req); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		req.User, req.Password, req.Host, req.Port, req.Name)

	db, err := pgxpool.New(c.Context(), connStr)
	if err != nil {
		return api.SendError(c, api.ErrCoreInternalServer, 500)
	}
	defer db.Close()

	if err := db.Ping(c.Context()); err != nil {
		return api.SendError(c, api.ErrCoreInternalServer, 500)
	}

	logger.Success("[SYSTEM/SETUP] Database connection test successful for %s:%s", req.Host, req.Port)
	return api.Success(c, api.SuccessSetupDBTestOK)
}

func saveSetupEnv(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	if err := config.SaveEnv(data); err != nil {
		return api.SendError(c, api.ErrCoreInternalServer, 500)
	}

	return restartEngine(c)
}

func SetPublicApp(app *fiber.App) {
	publicAppLock.Lock()
	defer publicAppLock.Unlock()
	publicApp = app
}

func restartEngine(c *fiber.Ctx) error {
	logger.Warn("[SYSTEM/ENGINE] API Route Restart requested via Admin Panel")

	go func() {
		time.Sleep(500 * time.Millisecond)
		logger.Info("[SYSTEM/ENGINE] Restarting whole engine to refresh API routes and configurations...")

		execPath, err := os.Executable()
		if err != nil {
			logger.Error("[SYSTEM/ENGINE] Failed to get executable: %v", err)
			os.Exit(1)
		}

		attr := &os.ProcAttr{
			Dir:   ".",
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		}

		_, err = os.StartProcess(execPath, os.Args, attr)
		if err != nil {
			logger.Error("[SYSTEM/ENGINE] Failed to spawn process: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	return api.Success(c, api.SuccessEngineRestarting)
}

func stopEngine(c *fiber.Ctx) error {
	publicAppLock.Lock()
	defer publicAppLock.Unlock()

	if !isRunning {
		return api.Success(c, api.SuccessEngineAlreadyStopped)
	}

	logger.Warn("[SYSTEM/ENGINE] Stopping Public API service...")

	if publicApp != nil {
		go func() {
			err := publicApp.Shutdown()
			if err != nil {
				logger.Error("[SYSTEM/ENGINE] Error during Public API shutdown: %v", err)
			}
			isRunning = false
			logger.Success("[SYSTEM/ENGINE] Public API service stopped")
		}()
	}

	return api.Success(c, api.SuccessEngineStopping)
}

func startEngine(c *fiber.Ctx) error {
	publicAppLock.Lock()
	defer publicAppLock.Unlock()

	if isRunning {
		return api.Success(c, api.SuccessEngineAlreadyRunning)
	}

	logger.Info("[SYSTEM/ENGINE] Starting Public API service...")

	publicApp = public.NewPublicApp(dbPool)
	isRunning = true

	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.PublicAPIHost, cfg.PublicAPIPort)
		logger.Info("[SYSTEM/ENGINE] Public API listening on %s", addr)
		if err := publicApp.Listen(addr); err != nil {
			logger.Error("[SYSTEM/ENGINE] Public API failed to start: %v", err)
			isRunning = false
		}
	}()

	return api.Success(c, api.SuccessEngineStarting)
}
