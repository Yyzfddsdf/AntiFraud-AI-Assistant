package family_system

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	appcfg "antifraud/config"
	"antifraud/login_system/smscode"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

// RegisterRoutes 注册家庭系统路由。
func RegisterRoutes(router *gin.RouterGroup, service *Service) {
	if router == nil || service == nil {
		return
	}

	router.POST("/families", createFamilyHandle(service))
	router.GET("/families/me", getMyFamilyHandle(service))
	router.POST("/families/invitations", createInvitationHandle(service))
	router.GET("/families/invitations", listInvitationsHandle(service))
	router.GET("/families/invitations/received", listReceivedInvitationsHandle(service))
	router.POST("/families/invitations/accept", acceptInvitationHandle(service))
	router.GET("/families/members", listMembersHandle(service))
	router.PATCH("/families/members/:memberId", updateMemberHandle(service))
	router.DELETE("/families/members/:memberId", deleteMemberHandle(service))
	router.POST("/families/guardian-links", createGuardianLinkHandle(service))
	router.GET("/families/guardian-links", listGuardianLinksHandle(service))
	router.DELETE("/families/guardian-links/:linkId", deleteGuardianLinkHandle(service))
	router.GET("/families/notifications/ws", notificationsWebSocketHandle(service))
	router.POST("/families/notifications/:notificationId/read", markNotificationReadHandle(service))
}

const (
	defaultFamilyNotificationPollInterval = 30 * time.Second
	defaultFamilyNotificationRecentWindow = 1 * time.Hour
)

type familyNotificationRuntimeConfig struct {
	pollInterval time.Duration
	recentWindow time.Duration
}

type familyNotificationWSMessage struct {
	Type           string `json:"type"`
	NotificationID uint   `json:"notification_id"`
	FamilyID       uint   `json:"family_id"`
	TargetUserID   uint   `json:"target_user_id"`
	TargetName     string `json:"target_name"`
	EventType      string `json:"event_type"`
	RecordID       string `json:"record_id"`
	Title          string `json:"title"`
	CaseSummary    string `json:"case_summary,omitempty"`
	ScamType       string `json:"scam_type,omitempty"`
	Summary        string `json:"summary"`
	RiskLevel      string `json:"risk_level"`
	EventAt        string `json:"event_at"`
	ReadAt         string `json:"read_at,omitempty"`
}

func createFamilyHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		var input CreateFamilyInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		result, err := service.CreateFamily(c.Request.Context(), userID, input)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

func getMyFamilyHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		result, err := service.GetMyFamily(c.Request.Context(), userID)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func createInvitationHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		var input CreateFamilyInvitationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		invitation, err := service.CreateInvitation(c.Request.Context(), userID, input)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"invitation": invitation})
	}
}

func listInvitationsHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		result, err := service.ListInvitations(c.Request.Context(), userID)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"invitations": result})
	}
}

func listReceivedInvitationsHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		result, err := service.ListReceivedInvitations(c.Request.Context(), userID)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"invitations": result})
	}
}

func acceptInvitationHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		var input AcceptFamilyInvitationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		result, err := service.AcceptInvitation(c.Request.Context(), userID, input)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func listMembersHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		result, err := service.ListMembers(c.Request.Context(), userID)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"members": result})
	}
}

func updateMemberHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		memberID, err := parseUintParam(c.Param("memberId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "memberId 无效"})
			return
		}
		var input UpdateFamilyMemberInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		result, err := service.UpdateMember(c.Request.Context(), userID, memberID, input)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"member": result})
	}
}

func deleteMemberHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		memberID, err := parseUintParam(c.Param("memberId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "memberId 无效"})
			return
		}
		if err := service.RemoveMember(c.Request.Context(), userID, memberID); err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "家庭成员已移除"})
	}
}

func createGuardianLinkHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		var input CreateGuardianLinkInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		result, err := service.CreateGuardianLink(c.Request.Context(), userID, input)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"guardian_link": result})
	}
}

func listGuardianLinksHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		result, err := service.GetMyFamily(c.Request.Context(), userID)
		if err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"guardian_links": result.GuardianLinks})
	}
}

func deleteGuardianLinkHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		linkID, err := parseUintParam(c.Param("linkId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "linkId 无效"})
			return
		}
		if err := service.DeleteGuardianLink(c.Request.Context(), userID, linkID); err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "守护关系已移除"})
	}
}

func notificationsWebSocketHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		runtimeCfg := loadFamilyNotificationRuntimeConfig()
		wsServer := websocket.Server{
			Handshake: func(cfg *websocket.Config, req *http.Request) error {
				_ = cfg
				_ = req
				return nil
			},
			Handler: func(ws *websocket.Conn) {
				runFamilyNotificationSession(ws, userID, service, runtimeCfg)
			},
		}
		wsServer.ServeHTTP(c.Writer, c.Request)
	}
}

func markNotificationReadHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := resolveCurrentUserID(c)
		if !ok {
			return
		}
		notificationID, err := parseUintParam(c.Param("notificationId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "notificationId 无效"})
			return
		}
		if err := service.MarkNotificationRead(c.Request.Context(), userID, notificationID); err != nil {
			writeFamilyError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "通知已标记为已读"})
	}
}

func resolveCurrentUserID(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return 0, false
	}
	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户标识无效"})
		return 0, false
	}
	return userID, true
}

func parseUintParam(raw string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func writeFamilyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNoFamily):
		c.JSON(http.StatusNotFound, gin.H{"error": "当前用户未加入家庭"})
	case errors.Is(err, ErrFamilyAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "当前用户已加入家庭"})
	case errors.Is(err, ErrFamilyPermissionDenied):
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作当前家庭"})
	case errors.Is(err, ErrInvalidInvitationCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请码无效"})
	case errors.Is(err, ErrInvitationExpired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请已过期"})
	case errors.Is(err, ErrInvitationTargetMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前账号与邀请目标不匹配"})
	case errors.Is(err, ErrInvitationProcessed):
		c.JSON(http.StatusConflict, gin.H{"error": "邀请已处理"})
	case errors.Is(err, ErrInvalidFamilyRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的家庭角色，仅支持 guardian 或 member"})
	case errors.Is(err, ErrInvalidInvitationTarget):
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少填写受邀人的邮箱或手机号"})
	case errors.Is(err, ErrInvalidGuardianConfig):
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的守护关系配置"})
	case errors.Is(err, ErrFamilyMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "家庭成员不存在"})
	case errors.Is(err, ErrGuardianLinkNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "守护关系不存在"})
	case errors.Is(err, ErrFamilyOwnerImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "家庭创建者不可移除或降级"})
	case errors.Is(err, smscode.ErrInvalidPhoneFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "家庭系统处理失败"})
	}
}

func runFamilyNotificationSession(ws *websocket.Conn, userID uint, service *Service, runtimeCfg familyNotificationRuntimeConfig) {
	if ws == nil || service == nil {
		return
	}
	defer ws.Close()

	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer stop()
		for {
			var incoming string
			if err := websocket.Message.Receive(ws, &incoming); err != nil {
				return
			}
		}
	}()

	sentNotificationIDs := make(map[uint]struct{})
	if err := pushFamilyNotifications(ws, service, userID, sentNotificationIDs, runtimeCfg.recentWindow); err != nil {
		stop()
		return
	}

	ticker := time.NewTicker(runtimeCfg.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := pushFamilyNotifications(ws, service, userID, sentNotificationIDs, runtimeCfg.recentWindow); err != nil {
				stop()
				return
			}
		}
	}
}

func pushFamilyNotifications(ws *websocket.Conn, service *Service, userID uint, sentNotificationIDs map[uint]struct{}, recentWindow time.Duration) error {
	if ws == nil {
		return fmt.Errorf("family notification websocket is nil")
	}
	notifications, err := service.ListRecentUnreadNotifications(context.Background(), userID, recentWindow)
	if err != nil {
		return err
	}
	for _, item := range notifications {
		if _, exists := sentNotificationIDs[item.ID]; exists {
			continue
		}
		msg := familyNotificationWSMessage{
			Type:           "family_high_risk_alert",
			NotificationID: item.ID,
			FamilyID:       item.FamilyID,
			TargetUserID:   item.TargetUserID,
			TargetName:     strings.TrimSpace(item.TargetName),
			EventType:      strings.TrimSpace(item.EventType),
			RecordID:       strings.TrimSpace(item.RecordID),
			Title:          strings.TrimSpace(item.Title),
			CaseSummary:    strings.TrimSpace(item.CaseSummary),
			ScamType:       strings.TrimSpace(item.ScamType),
			Summary:        strings.TrimSpace(item.Summary),
			RiskLevel:      strings.TrimSpace(item.RiskLevel),
			EventAt:        strings.TrimSpace(item.EventAt),
			ReadAt:         strings.TrimSpace(item.ReadAt),
		}
		if err := websocket.JSON.Send(ws, msg); err != nil {
			return fmt.Errorf("send family notification failed: %w", err)
		}
		sentNotificationIDs[item.ID] = struct{}{}
	}
	return nil
}

func loadFamilyNotificationRuntimeConfig() familyNotificationRuntimeConfig {
	result := familyNotificationRuntimeConfig{
		pollInterval: defaultFamilyNotificationPollInterval,
		recentWindow: defaultFamilyNotificationRecentWindow,
	}
	cfg, err := appcfg.LoadConfig("config/config.json")
	if err != nil || cfg == nil {
		return result
	}
	pollInterval := time.Duration(cfg.FamilyAlertWS.PollIntervalSeconds) * time.Second
	if pollInterval > 0 {
		result.pollInterval = pollInterval
	}
	recentWindow := time.Duration(cfg.FamilyAlertWS.RecentWindowMinutes) * time.Minute
	if recentWindow > 0 {
		result.recentWindow = recentWindow
	}
	return result
}
