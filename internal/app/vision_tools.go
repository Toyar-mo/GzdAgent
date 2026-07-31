package app

import (
	"github.com/Toyar-mo/GzdAgent/internal/config"
	"github.com/Toyar-mo/GzdAgent/internal/mcp"
	"github.com/Toyar-mo/GzdAgent/internal/vision"

	"go.uber.org/zap"
)

func registerVisionTools(mcpServer *mcp.Server, cfg *config.Config, logger *zap.Logger) {
	vision.RegisterAnalyzeImageTool(mcpServer, cfg, logger)
}
