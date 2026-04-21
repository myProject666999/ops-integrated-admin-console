package runtime

import (
	"fmt"
	"io"
	"net/http"
	"ops-admin-backend/internal/project"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type ginServer struct {
	*server
	router *gin.Engine
}

func newGinServer(s *server) *ginServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsGinMiddleware())

	gs := &ginServer{
		server: s,
		router: r,
	}

	gs.setupRoutes()
	return gs
}

func (gs *ginServer) setupRoutes() {
	gs.router.GET("/health", gs.handleHealth)
	gs.router.GET("/api/health", gs.handleHealth)

	api := gs.router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", gs.handleLogin)
			auth.POST("/register", gs.handleRegister)
			auth.GET("/me", gs.authMiddleware(), gs.handleMe)
			auth.POST("/window-close-start", gs.authMiddleware(), gs.handleWindowCloseStart)
			auth.POST("/window-close-cancel", gs.authMiddleware(), gs.handleWindowCloseCancel)
			auth.POST("/logout", gs.authMiddleware(), gs.handleLogout)
			auth.POST("/change-password", gs.authMiddleware(), gs.handleChangePassword)
		}

		projects := api.Group("/projects", gs.authMiddleware())
		{
			projects.GET("/credentials", gs.handleProjectCredentials)
			projects.PUT("/credentials/:type", gs.handleProjectCredentialByType)
			projects.POST("/relogin", gs.handleProjectsRelogin)
			projects.POST("/operate-async", gs.handleProjectOperateAsyncStart)
			projects.GET("/operate-async/:job_id", gs.handleProjectOperateAsyncStatus)
			projects.POST("/:type/load", gs.handleProjectLoad)
			projects.GET("/:type/batch-template", gs.handleProjectBatchTemplate)
			projects.POST("/:type/batch-upload", gs.handleProjectBatchUpload)
			projects.GET("/:type/batch-files", gs.handleProjectBatchFiles)
			projects.POST("/:type/operate", gs.handleProjectOperate)
		}

		api.GET("/logs", gs.authMiddleware(), gs.handleLogs)
	}

	gs.router.NoRoute(gs.serveStaticGin)
}

func (gs *ginServer) Run(addr string) error {
	return gs.router.Run(addr)
}

func (gs *ginServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (gs *ginServer) handleLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "用户名和密码不能为空"})
		return
	}

	var adminCount int
	if err := gs.db.QueryRow(`SELECT COUNT(1) FROM admins`).Scan(&adminCount); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "查询管理员失败"})
		return
	}
	if adminCount == 0 {
		c.JSON(http.StatusBadRequest, apiError{Error: "暂无管理员账号，请先注册"})
		return
	}

	var userID int64
	var username, hash string
	err := gs.db.QueryRow(`SELECT id,username,password_hash FROM admins WHERE username=?`, req.Username).Scan(&userID, &username, &hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apiError{Error: "账号或密码错误"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, apiError{Error: "账号或密码错误"})
		return
	}

	token, err := randomToken(48)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "生成令牌失败"})
		return
	}
	exp := time.Now().Add(gs.tokenTTL)
	now := nowStr()
	if _, err = gs.db.Exec(`INSERT INTO auth_tokens(token,user_id,expires_at,created_at) VALUES(?,?,?,?)`, token, userID, exp.Format(time.RFC3339), now); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "创建登录会话失败"})
		return
	}
	if err = ensureDefaultProjectCredentialsForUser(gs.db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "初始化项目凭据失败"})
		return
	}

	gs.logAction(userID, username, "login", "", "用户登录成功")
	c.JSON(http.StatusOK, loginResp{
		Token:                token,
		Username:             username,
		ExpireAt:             exp.Format(time.RFC3339),
		DefaultPwd:           false,
		ProjectCacheTTLInSec: int(gs.cfg.ProjectCacheTTL.Seconds()),
		SessionIdleTTLInSec:  int(gs.cfg.SessionIdleTTL.Seconds()),
	})
}

func (gs *ginServer) handleRegister(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "用户名和密码不能为空"})
		return
	}
	if len(username) < 3 || len(username) > 32 {
		c.JSON(http.StatusBadRequest, apiError{Error: "用户名长度必须为3-32位"})
		return
	}
	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, apiError{Error: "密码长度至少8位"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "密码加密失败"})
		return
	}
	res, err := gs.db.Exec(`INSERT INTO admins(username,password_hash,created_at,updated_at) VALUES(?,?,?,?)`, username, string(hash), nowStr(), nowStr())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, apiError{Error: "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, apiError{Error: "创建用户失败"})
		return
	}
	userID, _ := res.LastInsertId()
	if err = ensureDefaultProjectCredentialsForUser(gs.db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "初始化项目凭据失败"})
		return
	}
	gs.logAction(userID, username, "register", "", "管理员注册成功")
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func (gs *ginServer) handleMe(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	c.JSON(http.StatusOK, gin.H{
		"id":                        u.ID,
		"username":                  u.Username,
		"project_cache_ttl_seconds": int(gs.cfg.ProjectCacheTTL.Seconds()),
		"session_idle_ttl_seconds":  int(gs.cfg.SessionIdleTTL.Seconds()),
	})
}

func (gs *ginServer) handleWindowCloseStart(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	var req browserCloseEventReq
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	gs.scheduleBrowserCloseLifecycle(u, req)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (gs *ginServer) handleWindowCloseCancel(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	var req browserCloseEventReq
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	found, started := gs.cancelBrowserCloseState(u.Token)
	if !found || !started {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	detail := formatBrowserCloseCancelDetail(req)
	gs.logAction(u.ID, u.Username, "browser_close_timer_canceled", "", detail)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (gs *ginServer) handleLogout(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	var req browserCloseEventReq
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	gs.cancelBrowserCloseStatesByUser(u.ID)
	gs.cleanupUserAuthTokens(u.ID)
	if strings.TrimSpace(req.Reason) == "reopen_timeout" {
		detail := formatBrowserCloseEventDetail("页面关闭超时，已清理该账号全部 Token 与项目会话缓存", req)
		gs.logAction(u.ID, u.Username, "logout", "", detail)
	} else {
		gs.logAction(u.ID, u.Username, "logout", "", "管理员退出登录")
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (gs *ginServer) handleChangePassword(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	if strings.TrimSpace(req.OldPassword) == "" || strings.TrimSpace(req.NewPassword) == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "原密码和新密码不能为空"})
		return
	}
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, apiError{Error: "新密码长度至少8位"})
		return
	}

	var hash string
	if err := gs.db.QueryRow(`SELECT password_hash FROM admins WHERE id=?`, u.ID).Scan(&hash); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "查询当前密码失败"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)) != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "原密码错误"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "密码加密失败"})
		return
	}
	if _, err = gs.db.Exec(`UPDATE admins SET password_hash=?,updated_at=? WHERE id=?`, string(newHash), nowStr(), u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "更新密码失败"})
		return
	}
	gs.logAction(u.ID, u.Username, "change_password", "", "管理员修改密码")
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

func (gs *ginServer) handleProjectCredentials(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	if err := ensureDefaultProjectCredentialsForUser(gs.db, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "初始化项目凭据失败"})
		return
	}
	rows, err := gs.db.Query(`SELECT project_type,account,password,updated_at FROM project_credentials WHERE user_id=? ORDER BY project_type`, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "查询项目凭据失败"})
		return
	}
	defer rows.Close()

	items := make([]map[string]string, 0)
	for rows.Next() {
		var t, account, password, updated string
		if err = rows.Scan(&t, &account, &password, &updated); err != nil {
			c.JSON(http.StatusInternalServerError, apiError{Error: "读取项目凭据失败"})
			return
		}
		plainPwd, decErr := decryptCredentialPassword(password, gs.cfg.CredentialKey)
		if decErr != nil {
			c.JSON(http.StatusInternalServerError, apiError{Error: "项目凭据解密失败"})
			return
		}
		items = append(items, map[string]string{"project_type": t, "account": account, "password": plainPwd, "updated_at": updated})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (gs *ginServer) handleProjectCredentialByType(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if !validCredentialProjectType(projectType) {
		c.JSON(http.StatusBadRequest, apiError{Error: "无效的项目类型"})
		return
	}
	var req projectCredentialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	if strings.TrimSpace(req.Account) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "账号和密码不能为空"})
		return
	}
	encryptedPwd, err := encryptCredentialPassword(req.Password, gs.cfg.CredentialKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "凭据加密失败"})
		return
	}
	if _, err = gs.db.Exec(`INSERT INTO project_credentials(user_id,project_type,account,password,updated_at) VALUES(?,?,?,?,?)
	ON CONFLICT(user_id,project_type) DO UPDATE SET account=excluded.account,password=excluded.password,updated_at=excluded.updated_at`,
		u.ID, projectType, req.Account, encryptedPwd, nowStr(),
	); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "更新项目凭据失败"})
		return
	}
	gs.projectSessions.clearUserProject(u.ID, projectType)
	gs.logAction(u.ID, u.Username, "update_project_credential", projectType, "更新项目凭据")
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (gs *ginServer) handleProjectsRelogin(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	gs.projectSessions.clearToken(u.Token)
	reloginItems := make([]map[string]interface{}, 0, 3)
	for _, projectType := range []string{"ad", "print", "vpn"} {
		_, _, message, err := gs.ensureProjectSession(u, projectType, true)
		if err != nil {
			reloginItems = append(reloginItems, map[string]interface{}{
				"project_type": projectType,
				"ok":           false,
				"message":      err.Error(),
			})
			continue
		}
		reloginItems = append(reloginItems, map[string]interface{}{
			"project_type":  projectType,
			"ok":            true,
			"message":       message,
			"session_state": "countdown_relogin",
		})
	}
	gs.logAction(u.ID, u.Username, "project_relogin", "", "手动触发项目重新登录")
	c.JSON(http.StatusOK, gin.H{
		"items":           reloginItems,
		"next_cleanup_at": time.Now().Add(gs.cfg.ProjectCacheTTL).Format(time.RFC3339),
	})
}

func (gs *ginServer) handleProjectLoad(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if !validProjectType(projectType) {
		c.JSON(http.StatusBadRequest, apiError{Error: "无效的项目类型"})
		return
	}
	_, didLogin, message, err := gs.ensureProjectSession(u, projectType, false)
	if err != nil {
		gs.logAction(u.ID, u.Username, "project_load_failed", projectType, truncate(err.Error(), 600))
		c.JSON(http.StatusBadGateway, apiError{Error: err.Error()})
		return
	}
	if !didLogin {
		c.JSON(http.StatusOK, gin.H{"loaded": true, "first_load": false, "session_state": "reused"})
		return
	}
	gs.logAction(u.ID, u.Username, "project_load", projectType, "首次加载完成")
	c.JSON(http.StatusOK, gin.H{"loaded": true, "first_load": true, "message": message, "session_state": "first_login"})
}

func (gs *ginServer) handleProjectBatchFiles(c *gin.Context) {
	_ = c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, apiError{Error: "批量文件仅支持AD项目"})
		return
	}
	files, err := project.BatchExcelFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: err.Error()})
		return
	}
	items := make([]map[string]string, 0, len(files))
	for _, name := range files {
		items = append(items, map[string]string{
			"name": name,
			"path": filepath.Join(project.BatchUploadDir(), name),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"dir":   project.BatchUploadDir(),
	})
}

func (gs *ginServer) handleProjectBatchTemplate(c *gin.Context) {
	_ = c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, apiError{Error: "批量模板仅支持AD项目"})
		return
	}
	path := project.BatchTemplatePath()
	c.FileAttachment(path, filepath.Base(path))
}

func (gs *ginServer) handleProjectBatchUpload(c *gin.Context) {
	_ = c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, apiError{Error: "批量上传仅支持AD项目"})
		return
	}
	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "无效的表单数据"})
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "文件不能为空"})
		return
	}
	defer file.Close()

	oldFile := filepath.Base(strings.TrimSpace(c.Request.FormValue("old_file")))

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".xlsx" && ext != ".xls" {
		c.JSON(http.StatusBadRequest, apiError{Error: "仅支持上传 xlsx/.xls 文件"})
		return
	}

	storedName := "ad_batch_" + time.Now().Format("20060102150405") + ext
	outPath := filepath.Join(project.BatchUploadDir(), storedName)
	outFile, err := os.Create(outPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: err.Error()})
		return
	}
	defer outFile.Close()
	if _, err = io.Copy(outFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "保存文件失败"})
		return
	}

	if oldFile != "" {
		_ = os.Remove(filepath.Join(project.BatchUploadDir(), oldFile))
	}

	c.JSON(http.StatusOK, gin.H{
		"name":          storedName,
		"original_name": header.Filename,
		"path":          outPath,
	})
}

func (gs *ginServer) handleProjectOperate(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	projectType := c.Param("type")
	if !validProjectType(projectType) {
		c.JSON(http.StatusBadRequest, apiError{Error: "无效的项目类型"})
		return
	}
	var req operateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	if strings.TrimSpace(req.Action) == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "操作类型不能为空"})
		return
	}
	if req.Params == nil {
		req.Params = map[string]interface{}{}
	}
	if projectType == "vpn" && req.Action == "delete_users" && toBoolDefault(req.Params["remote_firewall"], false) {
		fwAccount, fwPassword, fwErr := gs.getProjectCredential(u.ID, "vpn_firewall")
		if fwErr != nil {
			req.Params["__vpn_fw_configured"] = false
			req.Params["__vpn_fw_error"] = fwErr.Error()
		} else {
			req.Params["__vpn_fw_configured"] = true
			req.Params["__vpn_fw_account"] = fwAccount
			req.Params["__vpn_fw_password"] = fwPassword
		}
	}

	entry, didLogin, _, err := gs.ensureProjectSession(u, projectType, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	result, err := gs.operateWithProjectSession(entry, req.Action, req.Params)
	if err != nil {
		gs.logAction(u.ID, u.Username, "project_operate_failed", projectType, fmt.Sprintf("action=%s, err=%v", req.Action, err))
		c.JSON(http.StatusBadGateway, apiError{Error: err.Error()})
		return
	}
	if !result.OK {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = result.Message
		}
		gs.logAction(u.ID, u.Username, "project_operate_failed", projectType, fmt.Sprintf("action=%s, err=%s", req.Action, errMsg))
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": errMsg, "message": result.Message, "data": result.Data})
		return
	}
	gs.logAction(u.ID, u.Username, "project_operate", projectType, fmt.Sprintf("action=%s", req.Action))
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": result.Message, "data": result.Data, "session_state": projectSessionStateFromDidLogin(didLogin)})
}

func (gs *ginServer) handleProjectOperateAsyncStart(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	var req asyncOperateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: "请求体格式错误"})
		return
	}
	req.ProjectType = strings.TrimSpace(req.ProjectType)
	req.Action = strings.TrimSpace(req.Action)
	if !validProjectType(req.ProjectType) {
		c.JSON(http.StatusBadRequest, apiError{Error: "无效的项目类型"})
		return
	}
	if req.Action == "" {
		c.JSON(http.StatusBadRequest, apiError{Error: "操作类型不能为空"})
		return
	}

	params := cloneInterfaceMap(req.Params)
	if params == nil {
		params = map[string]interface{}{}
	}
	if req.ProjectType == "vpn" && req.Action == "delete_users" && toBoolDefault(params["remote_firewall"], false) {
		fwAccount, fwPassword, fwErr := gs.getProjectCredential(u.ID, "vpn_firewall")
		if fwErr != nil {
			params["__vpn_fw_configured"] = false
			params["__vpn_fw_error"] = fwErr.Error()
		} else {
			params["__vpn_fw_configured"] = true
			params["__vpn_fw_account"] = fwAccount
			params["__vpn_fw_password"] = fwPassword
		}
	}

	_, didLogin, _, err := gs.ensureProjectSession(u, req.ProjectType, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	job, createErr := gs.createAsyncOperateJob(u, req.ProjectType, req.Action)
	if createErr != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "创建异步任务失败"})
		return
	}
	go gs.runAsyncOperate(job.ID, u, req.ProjectType, req.Action, params)
	c.JSON(http.StatusOK, gin.H{
		"job_id":        job.ID,
		"status":        job.Status,
		"created_at":    job.CreatedAt.Format(time.RFC3339),
		"project_type":  req.ProjectType,
		"action":        req.Action,
		"session_state": projectSessionStateFromDidLogin(didLogin),
	})
}

func (gs *ginServer) handleProjectOperateAsyncStatus(c *gin.Context) {
	u := c.MustGet("user").(authedUser)
	jobID := c.Param("job_id")
	view, ok := gs.getAsyncOperateJobView(jobID, u.ID)
	if !ok {
		c.JSON(http.StatusNotFound, apiError{Error: "任务不存在或已过期"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (gs *ginServer) handleLogs(c *gin.Context) {
	_ = c.MustGet("user").(authedUser)
	page := 1
	pageSize := 20

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 && n <= 1000 {
			pageSize = n
		}
	}
	if v := c.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			pageSize = n
		}
	}
	if pageSize > 200 {
		pageSize = 200
	}

	projectType := strings.TrimSpace(c.Query("project_type"))
	where := ""
	countArgs := make([]interface{}, 0, 1)
	if projectType != "" {
		where = ` WHERE project_type=?`
		countArgs = append(countArgs, projectType)
	}

	var total int
	countQuery := `SELECT COUNT(1) FROM operation_logs` + where
	if err := gs.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "查询日志失败"})
		return
	}

	offset := (page - 1) * pageSize
	query := `SELECT id,COALESCE(user_id,0),COALESCE(username,''),COALESCE(action,''),COALESCE(project_type,''),COALESCE(detail,''),created_at FROM operation_logs` +
		where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args := make([]interface{}, 0, len(countArgs)+2)
	args = append(args, countArgs...)
	args = append(args, pageSize, offset)

	rows, err := gs.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, apiError{Error: "查询日志失败"})
		return
	}
	defer rows.Close()
	items := make([]logRow, 0)
	for rows.Next() {
		var row logRow
		if err = rows.Scan(&row.ID, &row.UserID, &row.Username, &row.Action, &row.ProjectType, &row.Detail, &row.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, apiError{Error: "读取日志失败"})
			return
		}
		row.Detail = normalizeGarbledText(row.Detail)
		items = append(items, row)
	}
	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (gs *ginServer) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiError{Error: "缺少认证令牌"})
			return
		}
		now := time.Now().Format(time.RFC3339)
		u, err := gs.loadAuthedUser(token, now)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apiError{Error: "令牌无效或已过期"})
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func corsGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (gs *ginServer) serveStaticGin(c *gin.Context) {
	staticDir := gs.cfg.StaticDir
	if staticDir == "" {
		staticDir = "./static"
	}
	staticDir = filepath.Clean(staticDir)

	requestPath := c.Request.URL.Path
	if requestPath == "/" {
		requestPath = "/index.html"
	}

	filePath := filepath.Join(staticDir, requestPath)
	filePath = filepath.Clean(filePath)

	if !strings.HasPrefix(filePath, staticDir+string(filepath.Separator)) && filePath != staticDir {
		c.File(filepath.Join(staticDir, "index.html"))
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		c.File(filepath.Join(staticDir, "index.html"))
		return
	}

	if info.IsDir() {
		c.File(filepath.Join(staticDir, "index.html"))
		return
	}

	c.File(filePath)
}
