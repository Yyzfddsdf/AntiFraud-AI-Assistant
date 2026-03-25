package scam_simulation

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GeneratePackRequest struct {
	CaseType      string `json:"case_type"`
	TargetPersona string `json:"target_persona"`
	Difficulty    string `json:"difficulty"`
	Locale        string `json:"locale"`
}

type AnswerSessionRequest struct {
	PackID    string `json:"pack_id"`
	StepID    string `json:"step_id"`
	OptionKey string `json:"option_key"`
}

func RegisterRoutes(api *gin.RouterGroup, service *Service) {
	if api == nil || service == nil {
		return
	}

	group := api.Group("/scam/simulation")
	group.POST("/packs/generate", generatePackHandle(service))
	group.GET("/packs", listPacksHandle(service))
	group.GET("/packs/:packId/ongoing", getPackOngoingStatusHandle(service))
	group.GET("/sessions", listSessionsHandle(service))
	group.POST("/sessions/answer", answerSessionHandle(service))
	group.DELETE("/sessions/:sessionId", deleteSessionHandle(service))
}

func generatePackHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GeneratePackRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		userID := simulationCurrentUserID(c)
		_, err := service.StartGeneratePackTask(userID, GeneratePackInput{
			CaseType:      req.CaseType,
			TargetPersona: req.TargetPersona,
			Difficulty:    req.Difficulty,
			Locale:        req.Locale,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrUnfinishedSessionExists), errors.Is(err, ErrGeneratingTaskExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "action": "请刷新题目列表，若长时间无结果可重试生成"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "生成模拟题包失败: " + err.Error()})
			}
			return
		}

		go func() {
			for {
				task, taskErr := service.GetLatestPendingTask(userID)
				if taskErr != nil {
					break
				}
				if strings.TrimSpace(task.TaskID) == "" {
					break
				}
				_ = service.ProcessGeneratePackTask(task.TaskID)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"status": "submitted", "message": "题目生成任务已提交，请轮询题目列表"})
	}
}

func listPacksHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := simulationCurrentUserID(c)
		limit, offset := simulationListPagination(c)
		items, err := service.ListPacks(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询题包列表失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"packs": items, "limit": limit, "offset": offset})
	}
}

func getPackOngoingStatusHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := simulationCurrentUserID(c)
		packID := strings.TrimSpace(c.Param("packId"))
		if packID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "packId 不能为空"})
			return
		}

		session, pack, err := service.StartSession(userID, packID)
		if err != nil {
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
			case errors.Is(err, ErrPackAlreadyAttempted), errors.Is(err, ErrUnfinishedSessionExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询题目状态失败: " + err.Error()})
			}
			return
		}
		reportRows, listErr := service.ListSessions(userID, 200, 0)
		reportLevel := ""
		for i := range reportRows {
			if strings.TrimSpace(reportRows[i].PackID) == packID {
				reportLevel = strings.TrimSpace(reportRows[i].Level)
				break
			}
		}
		if listErr != nil {
			reportLevel = ""
		}

		nextStep := interface{}(nil)
		if session.CurrentStep >= 0 && session.CurrentStep < len(pack.Steps) {
			nextStep = pack.Steps[session.CurrentStep]
		}
		c.JSON(http.StatusOK, gin.H{
			"status":        session.Status,
			"pack_id":       packID,
			"current_step":  session.CurrentStep,
			"current_score": session.Score,
			"next_step":     nextStep,
			"result": gin.H{
				"total_score": session.Score,
				"level":       reportLevel,
			},
			"pack": pack,
		})
	}
}

func listSessionsHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := simulationCurrentUserID(c)
		limit, offset := simulationListPagination(c)
		items, err := service.ListSessions(userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询报告列表失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sessions": items, "limit": limit, "offset": offset})
	}
}

func answerSessionHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := simulationCurrentUserID(c)
		var req AnswerSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		if strings.TrimSpace(req.PackID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pack_id 不能为空"})
			return
		}

		if strings.TrimSpace(req.StepID) == "" && strings.TrimSpace(req.OptionKey) == "" {
			session, pack, err := service.StartSession(userID, req.PackID)
			if err != nil {
				switch {
				case errors.Is(err, gorm.ErrRecordNotFound):
					c.JSON(http.StatusNotFound, gin.H{"error": "题包不存在"})
				case errors.Is(err, ErrPackAlreadyAttempted), errors.Is(err, ErrUnfinishedSessionExists):
					c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"error": "开始答题失败: " + err.Error()})
				}
				return
			}
			nextStep := interface{}(nil)
			if session.CurrentStep >= 0 && session.CurrentStep < len(pack.Steps) {
				nextStep = pack.Steps[session.CurrentStep]
			}
			c.JSON(http.StatusOK, gin.H{"status": session.Status, "pack_id": session.PackID, "current_step": session.CurrentStep, "current_score": session.Score, "next_step": nextStep, "pack": pack})
			return
		}

		if strings.TrimSpace(req.StepID) == "" || strings.TrimSpace(req.OptionKey) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提交答案时 step_id/option_key 均不能为空"})
			return
		}

		session, pack, result, err := service.SubmitAnswerByPack(userID, req.PackID, req.StepID, req.OptionKey)
		if err != nil {
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "题包会话不存在"})
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			return
		}
		nextStep := interface{}(nil)
		if session.CurrentStep >= 0 && session.CurrentStep < len(pack.Steps) {
			nextStep = pack.Steps[session.CurrentStep]
		}
		c.JSON(http.StatusOK, gin.H{"status": session.Status, "pack_id": session.PackID, "current_step": session.CurrentStep, "current_score": session.Score, "next_step": nextStep, "result": result, "pack": pack})
	}
}

func deleteSessionHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := simulationCurrentUserID(c)
		packID := strings.TrimSpace(c.Param("sessionId"))
		if packID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "packId 不能为空"})
			return
		}
		err := service.DeleteSessionByPack(userID, packID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除报告失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "报告删除成功"})
	}
}

func simulationListPagination(c *gin.Context) (int, int) {
	if c == nil {
		return 20, 0
	}
	limit := 20
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			offset = parsed
		}
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func simulationCurrentUserID(c *gin.Context) string {
	if c == nil {
		return "demo-user"
	}
	userIDValue, exists := c.Get("userID")
	if !exists {
		return "demo-user"
	}
	if value, ok := userIDValue.(uint); ok {
		return fmt.Sprintf("%d", value)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", userIDValue))
}
