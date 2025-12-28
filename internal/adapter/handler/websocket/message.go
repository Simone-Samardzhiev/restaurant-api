package websocket

import "github.com/gin-gonic/gin"

var notWebSocketUpgradeErrorMessage = gin.H{
	"code":     "websocket_upgrade_required",
	"messages": []string{"Request should be a websocket upgrade to connect"},
}
