package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                 string
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	JWTExpiresHours        int
	CORSAllowedOrigin      string
	HistoryRetentionDays   int
	UploadDir              string
	UploadMaxMB            int
	SkillsConfigPath       string
	ShutdownTimeoutSeconds int
}

func Load() Config {
	loadDotEnv(".env")
	return Config{
		AppEnv:                 get("APP_ENV", "development"),
		Port:                   get("PORT", "8080"),
		DatabaseURL:            get("DATABASE_URL", "postgres://employee:employee@localhost:5432/employee_portal?sslmode=disable"),
		JWTSecret:              get("JWT_SECRET", "dev-secret"),
		JWTExpiresHours:        getInt("JWT_EXPIRES_HOURS", 8),
		CORSAllowedOrigin:      get("CORS_ALLOWED_ORIGIN", "http://localhost:5173,http://127.0.0.1:5173"),
		HistoryRetentionDays:   getInt("HISTORY_RETENTION_DAYS", 7),
		UploadDir:              get("UPLOAD_DIR", "uploads"),
		UploadMaxMB:            getInt("UPLOAD_MAX_MB", 3),
		SkillsConfigPath:       get("SKILLS_CONFIG_PATH", "config/skills.json"),
		ShutdownTimeoutSeconds: getInt("SHUTDOWN_TIMEOUT_SECONDS", 10),
	}
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
