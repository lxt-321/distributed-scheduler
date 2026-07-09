package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"dscheduler/common"
	"dscheduler/model"
	"dscheduler/scheduler"
	"dscheduler/store"
)

// Handler 调度中心 HTTP 接口
type Handler struct {
	Engine *scheduler.Engine
}

func NewHandler(e *scheduler.Engine) *Handler { return &Handler{Engine: e} }

// RegisterRoutes 注册所有 API 路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 任务管理
		api.GET("/job/page", h.jobPage)
		api.POST("/job/add", h.jobAdd)
		api.POST("/job/update", h.jobUpdate)
		api.POST("/job/remove", h.jobRemove)
		api.POST("/job/start", h.jobStart)
		api.POST("/job/stop", h.jobStop)
		api.POST("/job/trigger", h.jobTrigger)

		// 执行器分组
		api.GET("/group/page", h.groupPage)
		api.POST("/group/add", h.groupAdd)
		api.POST("/group/update", h.groupUpdate)
		api.POST("/group/remove", h.groupRemove)

		// 执行器回调 & 日志
		api.POST("/callback", h.callback)
		api.GET("/log/page", h.logPage)
	}
}

// ---------- 任务 ----------

func (h *Handler) jobPage(c *gin.Context) {
	jobs, err := store.ListJobs(50)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, jobs)
}

func (h *Handler) jobAdd(c *gin.Context) {
	var j model.JobInfo
	if err := c.ShouldBindJSON(&j); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if j.TriggerStatus == 0 {
		j.TriggerStatus = common.TriggerStatusStop
	}
	id, err := store.InsertJob(j)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	if j.TriggerStatus == common.TriggerStatusStart {
		if job, e := store.GetJob(id); e == nil {
			h.Engine.AddJob(job)
		}
	}
	common.Success(c, gin.H{"id": id})
}

func (h *Handler) jobUpdate(c *gin.Context) {
	var j model.JobInfo
	if err := c.ShouldBindJSON(&j); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := store.UpdateJob(j); err != nil {
		common.Fail(c, err.Error())
		return
	}
	if job, err := store.GetJob(j.ID); err == nil {
		h.Engine.AddJob(job)
	}
	common.Success(c, nil)
}

func (h *Handler) jobRemove(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if err := store.DeleteJob(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	h.Engine.RemoveJob(id)
	common.Success(c, nil)
}

func (h *Handler) jobStart(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if err := store.SetTriggerStatus(id, common.TriggerStatusStart); err != nil {
		common.Fail(c, err.Error())
		return
	}
	if job, err := store.GetJob(id); err == nil {
		h.Engine.AddJob(job)
	}
	common.Success(c, nil)
}

func (h *Handler) jobStop(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if err := store.SetTriggerStatus(id, common.TriggerStatusStop); err != nil {
		common.Fail(c, err.Error())
		return
	}
	h.Engine.RemoveJob(id)
	common.Success(c, nil)
}

func (h *Handler) jobTrigger(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if err := h.Engine.TriggerJob(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, nil)
}

// ---------- 执行器分组 ----------

func (h *Handler) groupPage(c *gin.Context) {
	groups, err := store.ListGroups()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, groups)
}

func (h *Handler) groupAdd(c *gin.Context) {
	var g model.JobGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	id, err := store.InsertGroup(g)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, gin.H{"id": id})
}

func (h *Handler) groupUpdate(c *gin.Context) {
	var g model.JobGroup
	if err := c.ShouldBindJSON(&g); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	if err := store.UpdateGroup(g); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, nil)
}

func (h *Handler) groupRemove(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	if err := store.DeleteGroup(id); err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, nil)
}

// ---------- 执行器回调 & 日志 ----------

// callback 接收执行器执行完成后的回调，更新执行日志
func (h *Handler) callback(c *gin.Context) {
	var cb model.CallbackParam
	if err := c.ShouldBindJSON(&cb); err != nil {
		common.Fail(c, "参数错误")
		return
	}
	for _, item := range cb.Callbacks {
		store.UpdateLogHandle(item.LogID, item.HandleCode, item.HandleMsg)
	}
	common.Success(c, nil)
}

func (h *Handler) logPage(c *gin.Context) {
	logs, err := store.ListLogs(50)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.Success(c, logs)
}
