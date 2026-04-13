package runtime

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginReqGin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerReqGin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type changePasswordReqGin struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type browserCloseEventReqGin struct {
	Reason         string `json:"reason"`
	ClosedAtMS     int64  `json:"closed_at_ms"`
	TimeoutAtMS    int64  `json:"timeout_at_ms"`
	ReopenedAtMS   int64  `json:"reopened_at_ms"`
	IdleTTLSeconds int    `json:"idle_ttl_seconds"`
}

type projectCredentialReqGin struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type operateReqGin struct {
	Action string                 `json:"action" binding:"required"`
	Params map[string]interface{} `json:"params"`
}

type asyncOperateReqGin struct {
	ProjectType string                 `json:"project_type" binding:"required"`
	Action      string                 `json:"action" binding:"required"`
	Params      map[string]interface{} `json:"params"`
}

func (s *server) handleLoginGin(c *gin.Context) {
	var req loginReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	var adminCount int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM admins`).Scan(&adminCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询管理员失败"})
		return
	}
	if adminCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "暂无管理员账号，请先注册"})
		return
	}

	var userID int64
	var dbUsername, hash string
	err := s.db.QueryRow(`SELECT id,username,password_hash FROM admins WHERE username=?`, username).Scan(&userID, &dbUsername, &hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	token, err := randomToken(48)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}
	exp := time.Now().Add(s.tokenTTL)
	now := nowStr()
	if _, err = s.db.Exec(`INSERT INTO auth_tokens(token,user_id,expires_at,created_at) VALUES(?,?,?,?)`, token, userID, exp.Format(time.RFC3339), now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建登录会话失败"})
		return
	}
	if err = ensureDefaultProjectCredentialsForUser(s.db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化项目凭据失败"})
		return
	}

	s.logAction(userID, dbUsername, "login", "", "用户登录成功")
	c.JSON(http.StatusOK, gin.H{
		"token":                     token,
		"username":                  dbUsername,
		"expire_at":                 exp.Format(time.RFC3339),
		"default_pwd":               false,
		"project_cache_ttl_seconds": int(s.cfg.ProjectCacheTTL.Seconds()),
		"session_idle_ttl_seconds":  int(s.cfg.SessionIdleTTL.Seconds()),
	})
}

func (s *server) handleRegisterGin(c *gin.Context) {
	var req registerReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}
	if len(username) < 3 || len(username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名长度必须为3-32位"})
		return
	}
	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度至少8位"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	res, err := s.db.Exec(`INSERT INTO admins(username,password_hash,created_at,updated_at) VALUES(?,?,?,?)`, username, string(hash), nowStr(), nowStr())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	userID, _ := res.LastInsertId()
	if err = ensureDefaultProjectCredentialsForUser(s.db, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化项目凭据失败"})
		return
	}

	s.logAction(userID, username, "register", "", "管理员注册成功")
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func (s *server) handleMeGin(c *gin.Context) {
	u := getAuthedUser(c)
	c.JSON(http.StatusOK, gin.H{
		"id":                        u.ID,
		"username":                  u.Username,
		"project_cache_ttl_seconds": int(s.cfg.ProjectCacheTTL.Seconds()),
		"session_idle_ttl_seconds":  int(s.cfg.SessionIdleTTL.Seconds()),
	})
}

func (s *server) handleLogoutGin(c *gin.Context) {
	u := getAuthedUser(c)
	var req browserCloseEventReqGin
	_ = c.ShouldBindJSON(&req)

	s.cancelBrowserCloseStatesByUser(u.ID)
	s.cleanupUserAuthTokens(u.ID)

	if strings.TrimSpace(req.Reason) == "reopen_timeout" {
		detail := formatBrowserCloseEventDetail("页面关闭超时，已清理该账号全部 Token 与项目会话缓存", browserCloseEventReq(req))
		s.logAction(u.ID, u.Username, "logout", "", detail)
		fmt.Printf("[browser-close-timeout] user=%s detail=%s\n", u.Username, detail)
	} else {
		s.logAction(u.ID, u.Username, "logout", "", "管理员退出登录")
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *server) handleChangePasswordGin(c *gin.Context) {
	u := getAuthedUser(c)
	var req changePasswordReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPassword == "" || newPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码和新密码不能为空"})
		return
	}
	if len(newPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度至少8位"})
		return
	}

	var hash string
	if err := s.db.QueryRow(`SELECT password_hash FROM admins WHERE id=?`, u.ID).Scan(&hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询当前密码失败"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPassword)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if _, err = s.db.Exec(`UPDATE admins SET password_hash=?,updated_at=? WHERE id=?`, string(newHash), nowStr(), u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	s.logAction(u.ID, u.Username, "change_password", "", "管理员修改密码")
	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

func (s *server) handleWindowCloseStartGin(c *gin.Context) {
	u := getAuthedUser(c)
	var req browserCloseEventReqGin
	_ = c.ShouldBindJSON(&req)
	s.scheduleBrowserCloseLifecycle(u, browserCloseEventReq(req))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *server) handleWindowCloseCancelGin(c *gin.Context) {
	u := getAuthedUser(c)
	var req browserCloseEventReqGin
	_ = c.ShouldBindJSON(&req)

	found, started := s.cancelBrowserCloseState(u.Token)
	if !found || !started {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	detail := formatBrowserCloseCancelDetail(browserCloseEventReq(req))
	s.logAction(u.ID, u.Username, "browser_close_timer_canceled", "", detail)
	fmt.Printf("[browser-close-cancel] user=%s detail=%s\n", u.Username, detail)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *server) handleProjectCredentialsGin(c *gin.Context) {
	u := getAuthedUser(c)
	if err := ensureDefaultProjectCredentialsForUser(s.db, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "初始化项目凭据失败"})
		return
	}

	rows, err := s.db.Query(`SELECT project_type,account,password,updated_at FROM project_credentials WHERE user_id=? ORDER BY project_type`, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询项目凭据失败"})
		return
	}
	defer rows.Close()

	items := make([]map[string]string, 0)
	for rows.Next() {
		var t, account, password, updated string
		if err = rows.Scan(&t, &account, &password, &updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取项目凭据失败"})
			return
		}
		plainPwd, decErr := decryptCredentialPassword(password, s.cfg.CredentialKey)
		if decErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "项目凭据解密失败"})
			return
		}
		items = append(items, map[string]string{
			"project_type": t,
			"account":      account,
			"password":     plainPwd,
			"updated_at":   updated,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *server) handleProjectCredentialByTypeGin(c *gin.Context) {
	u := getAuthedUser(c)
	projectType := c.Param("type")

	if !validCredentialProjectType(projectType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目类型"})
		return
	}

	var req projectCredentialReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	encryptedPwd, err := encryptCredentialPassword(req.Password, s.cfg.CredentialKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "凭据加密失败"})
		return
	}

	if _, err = s.db.Exec(`INSERT INTO project_credentials(user_id,project_type,account,password,updated_at) VALUES(?,?,?,?,?)
	ON CONFLICT(user_id,project_type) DO UPDATE SET account=excluded.account,password=excluded.password,updated_at=excluded.updated_at`,
		u.ID, projectType, req.Account, encryptedPwd, nowStr(),
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新项目凭据失败"})
		return
	}

	s.projectSessions.clearUserProject(u.ID, projectType)
	s.logAction(u.ID, u.Username, "update_project_credential", projectType, "更新项目凭据")
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (s *server) handleProjectsReloginGin(c *gin.Context) {
	u := getAuthedUser(c)
	s.projectSessions.clearToken(u.Token)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *server) handleProjectLoadGin(c *gin.Context) {
	u := getAuthedUser(c)
	projectType := c.Param("type")

	if !validProjectType(projectType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目类型"})
		return
	}

	_, didLogin, message, err := s.ensureProjectSession(u, projectType, false)
	if err != nil {
		s.logAction(u.ID, u.Username, "project_load_failed", projectType, truncate(err.Error(), 600))
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if !didLogin {
		c.JSON(http.StatusOK, gin.H{"loaded": true, "first_load": false, "session_state": "reused"})
		return
	}

	s.logAction(u.ID, u.Username, "project_load", projectType, "首次加载完成")
	c.JSON(http.StatusOK, gin.H{
		"loaded":        true,
		"first_load":    true,
		"message":       message,
		"session_state": "first_login",
	})
}

func (s *server) handleProjectBatchTemplateGin(c *gin.Context) {
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量模板仅支持AD项目"})
		return
	}

	path := projectBatchTemplatePath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "模板文件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := filepath.Base(path)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename)))
	c.File(path)
}

func (s *server) handleProjectBatchUploadGin(c *gin.Context) {
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量上传仅支持AD项目"})
		return
	}

	if err := os.MkdirAll(projectBatchUploadDir(), 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件不能为空"})
		return
	}
	defer file.Close()

	oldFile := strings.TrimSpace(c.PostForm("old_file"))

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".xlsx" && ext != ".xls" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持上传 xlsx/.xls 文件"})
		return
	}

	storedName := fmt.Sprintf("ad_batch_%d%s", time.Now().UnixNano(), ext)
	outPath := filepath.Join(projectBatchUploadDir(), storedName)
	outFile, err := os.Create(outPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	if oldFile != "" {
		_ = os.Remove(filepath.Join(projectBatchUploadDir(), oldFile))
	}

	c.JSON(http.StatusOK, gin.H{
		"name":          storedName,
		"original_name": header.Filename,
		"path":          outPath,
	})
}

func (s *server) handleProjectBatchFilesGin(c *gin.Context) {
	projectType := c.Param("type")
	if projectType != "ad" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "批量文件仅支持AD项目"})
		return
	}

	files, err := projectBatchExcelFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]map[string]string, 0, len(files))
	for _, name := range files {
		items = append(items, map[string]string{
			"name": name,
			"path": filepath.Join(projectBatchUploadDir(), name),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"dir":   projectBatchUploadDir(),
	})
}

func (s *server) handleProjectOperateGin(c *gin.Context) {
	u := getAuthedUser(c)
	projectType := c.Param("type")

	if !validProjectType(projectType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目类型"})
		return
	}

	var req operateReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	params := req.Params
	if params == nil {
		params = map[string]interface{}{}
	}

	if projectType == "vpn" && req.Action == "delete_users" && toBoolDefault(params["remote_firewall"], false) {
		fwAccount, fwPassword, fwErr := s.getProjectCredential(u.ID, "vpn_firewall")
		if fwErr != nil {
			params["__vpn_fw_configured"] = false
			params["__vpn_fw_error"] = fwErr.Error()
		} else {
			params["__vpn_fw_configured"] = true
			params["__vpn_fw_account"] = fwAccount
			params["__vpn_fw_password"] = fwPassword
		}
	}

	entry, didLogin, _, err := s.ensureProjectSession(u, projectType, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.operateWithProjectSession(entry, req.Action, params)
	if err != nil {
		s.logAction(u.ID, u.Username, "project_operate_failed", projectType, fmt.Sprintf("action=%s, err=%v", req.Action, err))
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if !result.OK {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = result.Message
		}
		s.logAction(u.ID, u.Username, "project_operate_failed", projectType, fmt.Sprintf("action=%s, err=%s", req.Action, errMsg))
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":            false,
			"error":         errMsg,
			"message":       result.Message,
			"data":          result.Data,
			"session_state": projectSessionStateFromDidLogin(didLogin),
		})
		return
	}

	s.logAction(u.ID, u.Username, "project_operate", projectType, fmt.Sprintf("action=%s", req.Action))
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"message":       result.Message,
		"data":          result.Data,
		"session_state": projectSessionStateFromDidLogin(didLogin),
	})
}

func (s *server) handleProjectOperateAsyncStartGin(c *gin.Context) {
	u := getAuthedUser(c)
	var req asyncOperateReqGin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	projectType := strings.TrimSpace(req.ProjectType)
	action := strings.TrimSpace(req.Action)
	if !validProjectType(projectType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目类型"})
		return
	}
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "操作类型不能为空"})
		return
	}

	params := cloneInterfaceMap(req.Params)
	if params == nil {
		params = map[string]interface{}{}
	}

	if projectType == "vpn" && action == "delete_users" && toBoolDefault(params["remote_firewall"], false) {
		fwAccount, fwPassword, fwErr := s.getProjectCredential(u.ID, "vpn_firewall")
		if fwErr != nil {
			params["__vpn_fw_configured"] = false
			params["__vpn_fw_error"] = fwErr.Error()
		} else {
			params["__vpn_fw_configured"] = true
			params["__vpn_fw_account"] = fwAccount
			params["__vpn_fw_password"] = fwPassword
		}
	}

	_, didLogin, _, err := s.ensureProjectSession(u, projectType, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, createErr := s.createAsyncOperateJob(u, projectType, action)
	if createErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建异步任务失败"})
		return
	}

	go s.runAsyncOperate(job.ID, u, projectType, action, params)
	c.JSON(http.StatusOK, gin.H{
		"job_id":        job.ID,
		"status":        job.Status,
		"created_at":    job.CreatedAt.Format(time.RFC3339),
		"project_type":  projectType,
		"action":        action,
		"session_state": projectSessionStateFromDidLogin(didLogin),
	})
}

func (s *server) handleProjectOperateAsyncStatusGin(c *gin.Context) {
	u := getAuthedUser(c)
	jobID := strings.TrimSpace(c.Param("jobId"))

	if jobID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	view, ok := s.getAsyncOperateJobView(jobID, u.ID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在或已过期"})
		return
	}

	c.JSON(http.StatusOK, view)
}

func (s *server) handleLogsGin(c *gin.Context) {
	_ = getAuthedUser(c)
	rows, err := s.db.Query(`SELECT id,user_id,username,action,project_type,detail,created_at FROM operation_logs ORDER BY id DESC LIMIT 200`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
		return
	}
	defer rows.Close()

	items := make([]logRow, 0)
	for rows.Next() {
		var r logRow
		if err = rows.Scan(&r.ID, &r.UserID, &r.Username, &r.Action, &r.ProjectType, &r.Detail, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取日志失败"})
			return
		}
		items = append(items, r)
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
