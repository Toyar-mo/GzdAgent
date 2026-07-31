package app

import (
	"github.com/Toyar-mo/GzdAgent/internal/ctfknowledge"
	"github.com/Toyar-mo/GzdAgent/internal/mcp"

	"go.uber.org/zap"
)

// registerCTFKnowledgeTools 注册 CTF 知识库相关工具到 MCP 服务器
func registerCTFKnowledgeTools(mcpServer *mcp.Server, wpManager *ctfknowledge.Manager, logger *zap.Logger) {
	if wpManager == nil {
		logger.Warn("跳过 CTF 知识库工具注册：wpManager 为空")
		return
	}
	ctfknowledge.RegisterWPKnowledgeTool(mcpServer, wpManager, logger)
}
