package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	ad "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/admin"
	as "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/asset"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/auth"
	br "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/branch"
	ct "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/contract"
	tgeval "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/evaluation"
	fcm "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/fcmtoken"
	hld "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/holiday"
	jt "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/jobtitle"
	kb "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/kanbanboard"
	kl "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/kanbanlist"
	kt "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/kanbantask"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/leave"
	n "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/notification"
	org "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/organization"
	ot "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/overtime"
	proj "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/projects"
	rc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/recruitment"
	rq "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/registrationrequest"
	tech "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/technology"
	tk "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/timekeeping"
	up "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/userpermission"
	uprj "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/userproject"
	u "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/users"
	ut "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/domains/usertechnology"
	gc "gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/cloud"
	"gitlab.vietnamlab.vn/micro_erp/frontend-api/internal/platform/utils"
)

type AppRouter struct {
	authCtr           *auth.AuthController
	orgCtr            *org.OrgController
	userCtr           *u.UserController
	requestCtr        *rq.RegistRequestController
	projCtr           *proj.ProjectController
	tgevalCtr         *tgeval.TargetEvalController
	tkCtr             *tk.TkController
	leaveCtr          *leave.LvController
	uprjCtr           *uprj.Controller
	branchCtr         *br.Controller
	jobTitleCtr       *jt.Controller
	techCtr           *tech.Controller
	utechCtr          *ut.Controller
	overtimeCtr       *ot.Controller
	holidayCtr        *hld.Controller
	notificationCtr   *n.Controller
	fcmTokenCtr       *fcm.Controller
	recruitmentCtr    *rc.Controller
	kanbanBoardCtr    *kb.Controller
	kanbanListCtr     *kl.Controller
	kanbanTaskCtr     *kt.Controller
	userPermissionCtr *up.Controller
	adminCtr          *ad.Controller
	assetCtr          *as.Controller
	contractCtr       *ct.Controller

	userMw *u.UserMiddleware
	gcs    *gc.GcsStorage
}

func NewAppRouter(logger echo.Logger) (r *AppRouter) {
	userRepo := u.NewPgUserRepository(logger)
	orgRepo := org.NewPgOrgRepository(logger)
	regRepo := auth.NewPgRegCodeRepository(logger)
	requestRepo := rq.NewPgRequestRepository(logger)
	projRepo := proj.NewPgProjRepository(logger)
	tgevalRepo := tgeval.NewPgEvaluationRepository(logger)
	timekeepingRepo := tk.NewPgTimekeepingRepository(logger)
	leaveRepo := leave.NewPgLeaveRepository(logger)
	userProjectRepo := uprj.NewPgUserProjectRepository(logger)
	branchRepo := br.NewPgBranchRepository(logger)
	jobTitleRepo := jt.NewPgJobTitleRepository(logger)
	techRepo := tech.NewPgTechnologyRepository(logger)
	userTechnologyRepo := ut.NewPgUserTechnologyRepository(logger)
	overtimeRepo := ot.NewPgOvertimeRepository(logger)
	notificationRepo := n.NewPgNotificationRepository(logger)
	holidayRepo := hld.NewPgHolidayRepository(logger)
	fcmTokenRepo := fcm.NewPgFcmTokenRepository(logger)
	recruitmentRepo := rc.NewPgRecruitmentRepository(logger)
	kanbanBoardRepo := kb.NewPgKanbanBoardRepository(logger)
	kanbanListRepo := kl.NewPgKanbanListRepository(logger)
	kanbanTaskRepo := kt.NewPgKanbanTaskRepository(logger)
	userPermissionRepo := up.NewPgUserPermissionRepository(logger)
	adminRepo := ad.NewPgAdminRepository(logger)
	assetRepo := as.NewPgAssetRepository(logger)
	contractRepo := ct.NewPgContractRepository(logger)

	gcsStorage := gc.NewGcsStorage(logger)

	r = &AppRouter{
		authCtr: auth.NewAuthController(logger, regRepo, userRepo),
		orgCtr:  org.NewOrgController(logger, userRepo, regRepo, orgRepo, requestRepo, adminRepo),
		userCtr: u.NewUserController(
			logger, userRepo, orgRepo, userProjectRepo, leaveRepo, projRepo,
			tgevalRepo, branchRepo, jobTitleRepo, techRepo, userTechnologyRepo, userPermissionRepo, gcsStorage,
		),
		requestCtr:  rq.NewRegistRequestController(logger, userRepo, regRepo, orgRepo, requestRepo),
		projCtr:     proj.NewProjectController(logger, projRepo, userRepo, userProjectRepo, gcsStorage),
		tgevalCtr:   tgeval.NewTargetEvaluationController(logger, tgevalRepo, userRepo, projRepo, userProjectRepo, branchRepo, gcsStorage),
		tkCtr:       tk.NewTimekeepingController(logger, timekeepingRepo, userRepo, branchRepo),
		leaveCtr:    leave.NewLeaveController(logger, leaveRepo, userRepo, branchRepo, orgRepo, holidayRepo, notificationRepo, userProjectRepo, fcmTokenRepo, gcsStorage),
		uprjCtr:     uprj.NewUserProjectController(logger, userProjectRepo, userRepo, projRepo, branchRepo),
		branchCtr:   br.NewBranchController(logger, branchRepo, userRepo, orgRepo),
		jobTitleCtr: jt.NewJobTitleController(logger, jobTitleRepo, userRepo, orgRepo),
		techCtr:     tech.NewTechnologyController(logger, techRepo, userTechnologyRepo, orgRepo),
		overtimeCtr: ot.NewOvertimeController(logger, overtimeRepo, orgRepo, userRepo,
			projRepo, branchRepo, leaveRepo, holidayRepo, notificationRepo, userProjectRepo, fcmTokenRepo),
		utechCtr:        ut.NewUserTechnologyController(logger, userTechnologyRepo, techRepo, userRepo),
		notificationCtr: n.NewNotificationController(logger, notificationRepo, fcmTokenRepo, userRepo, gcsStorage, orgRepo),
		holidayCtr:      hld.NewHolidayController(logger, holidayRepo, orgRepo),
		fcmTokenCtr:     fcm.NewFcmTokenController(logger, fcmTokenRepo),
		recruitmentCtr:  rc.NewRecruitmentController(logger, gcsStorage, recruitmentRepo, projRepo, branchRepo, userRepo, notificationRepo, fcmTokenRepo),
		kanbanBoardCtr:  kb.NewKanbanBoardController(logger, kanbanBoardRepo, projRepo, userProjectRepo),
		kanbanListCtr:   kl.NewKanbanListController(logger, kanbanListRepo, kanbanBoardRepo, userProjectRepo),
		kanbanTaskCtr: kt.NewKanbanTaskController(
			logger, gcsStorage, kanbanTaskRepo,
			kanbanListRepo, kanbanBoardRepo, userProjectRepo,
			projRepo, notificationRepo, fcmTokenRepo, userRepo,
		),
		userPermissionCtr: up.NewUserPermissionController(logger, userPermissionRepo, userRepo, orgRepo),
		adminCtr:          ad.NewAdminController(logger, adminRepo, userRepo, userPermissionRepo),
		assetCtr:          as.NewAssetController(logger, assetRepo, userRepo, branchRepo, notificationRepo, fcmTokenRepo),
		contractCtr:       ct.NewContractController(logger, contractRepo, userRepo, branchRepo, gcsStorage),

		userMw: u.NewUserMiddleware(logger, userRepo),
	}

	return
}

func (r *AppRouter) AuthRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/login", r.authCtr.Login)
	g.GET("/logout", r.authCtr.Logout, isLoggedIn)
	g.GET("/login-google", r.authCtr.OauthGoogleLogin)
	g.GET("/login-google-callback", r.authCtr.OauthGoogleCallback)
}

func (r *AppRouter) OrgRoute(g *echo.Group) {
	g.POST("/find-organization", r.orgCtr.FindOrganization)
	g.POST("/get-organizations", r.orgCtr.GetOrganizations)
}

func (r *AppRouter) RegRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/checkEmail", r.authCtr.CreateRegRequest)
	g.POST("/checkRegistrationCode", r.authCtr.CheckRegistrationCode)
	g.POST("/checkOrganization", r.orgCtr.CheckOrganization)
	g.POST("/registerOrganization", r.orgCtr.RegisterOrganization)
	g.POST("/registerInviteLink", r.orgCtr.RegisterInviteLink)
	g.POST("/requestRegistration", r.requestCtr.RequestRegistration)
	g.POST("/inviteUser", r.requestCtr.InviteUser, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.GET("/register-org-google", r.authCtr.RegisterOrgGoogle)
	g.GET("/register-org-google-callback", r.authCtr.RegisterOrgGoogleCallBack)
	g.POST("/download-template", r.authCtr.DownloadTemplate, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

func (r *AppRouter) UserRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.GET("/getuser", r.userCtr.GetLoginUser, isLoggedIn)
	g.POST("/forgotPassword", r.userCtr.ForgotPassword)
	g.POST("/checkresetpasswordcode", r.userCtr.CheckResetPasswordCode)
	g.POST("/resetpassword", r.userCtr.ResetPassword)
	g.POST("/changepassword", r.userCtr.ChangePassword, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/insertnewemailforupdate", r.userCtr.InsertNewEmailForUpdate, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/changeemail", r.userCtr.ChangeEmail)
	g.POST("/search-user-profile", r.userCtr.SearchUserProfile, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-user-info", r.userCtr.GetUserInfo, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-list-item-profile", r.userCtr.GetListItemProfile, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/update-profile", r.userCtr.UpdateProfile, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-avatar", r.userCtr.GetAvatarUserLogin, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/display-language-setting", r.userCtr.DisplayLanguageSetting, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/import-profiles", r.userCtr.ImportProfiles, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/export-excel", r.userCtr.ExportExcel, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/download-template", r.userCtr.DownloadTemplate, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

// RequestRoute      : create route for group /request
// Params            : echo.Group
func (r *AppRouter) RequestRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST(
		"/searchListRequest",
		r.requestCtr.SearchListRequest,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckAllManager)

	g.POST(
		"/updateRequestStatus",
		r.requestCtr.UpdateRequestStatus,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckAllManager)

	g.POST(
		"/resendEmailRegister",
		r.requestCtr.ResendEmailRegister,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckAllManager)
}

// ProjectRoute      : create route for projects
// Params            : echo.Group
func (r *AppRouter) ProjectRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/get-project-list", r.projCtr.GetProjectList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-project-details", r.projCtr.GetProjectDetails, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-projects-assigned", r.projCtr.GetProjectOfUser, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/add-project", r.projCtr.AddProject, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/update-project", r.projCtr.UpdateProject, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/delete-project", r.projCtr.DeleteProject, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/create-kanban-board", r.kanbanBoardCtr.CreateKanbanBoard, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-kanban-board", r.kanbanBoardCtr.EditKanbanBoard, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-kanban-boards", r.kanbanBoardCtr.GetKanbanBoards, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-kanban-board", r.kanbanBoardCtr.RemoveKanbanBoard, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-kanban-list", r.kanbanListCtr.CreateKanbanList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-kanban-list", r.kanbanListCtr.EditKanbanList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-kanban-list", r.kanbanListCtr.RemoveKanbanList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-kanban-task", r.kanbanTaskCtr.CreateKanbanTask, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-kanban-task", r.kanbanTaskCtr.EditKanbanTask, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-kanban-tasks", r.kanbanTaskCtr.GetKanbanTasks, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-kanban-task", r.kanbanTaskCtr.GetKanbanTask, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-kanban-task", r.kanbanTaskCtr.RemoveKanbanTask, isLoggedIn, r.userMw.InitUserProfile)
}

// TargetEvalRoute : create route for group /targeteval
// Params          : echo.Group
func (r *AppRouter) TargetEvalRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})
	g.POST("/create-evaluation", r.tgevalCtr.CreateEvaluationForm, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/duplicate-evaluation", r.tgevalCtr.DuplicateEvaluationForm, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-evaluation", r.tgevalCtr.GetEvaluationForm, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/update-evaluation", r.tgevalCtr.UpdateEvaluationForm, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/search-evaluation-list", r.tgevalCtr.SearchEvaluationList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-evaluations-by-id", r.tgevalCtr.EvaluationListByUserID, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/delete-evaluation", r.tgevalCtr.DeleteEvaluation, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/check-evaluation-existed", r.tgevalCtr.CheckEvalExist, isLoggedIn, r.userMw.InitUserProfile)
	g.POST(
		"/get-comment-two-consecutive-quarter",
		r.tgevalCtr.GetCommentTwoConsecutiveQuarter,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckAllManager,
	)
	g.GET("/export-excel", r.tgevalCtr.ExportMultipleExcel, isLoggedIn, r.userMw.InitUserProfile)
	g.GET("/export-evaluation-list", r.tgevalCtr.ExportEvaluationList, isLoggedIn, r.userMw.InitUserProfile)
}

// TimekeepingRoute : create route for group /timekeeping
// Params          : echo.Group
func (r *AppRouter) TimekeepingRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/check-in", r.tkCtr.CheckIn, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/check-out", r.tkCtr.CheckOut, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-timekeeping-today", r.tkCtr.GetTimekeeping, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-all-timekeeping-user", r.tkCtr.GetAllTimekeepingUser, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-all-timekeeping", r.tkCtr.GetAllTimekeeping, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.GET("/export-excel", r.tkCtr.ExportExcel, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

// LeaveRoute : create route for group /leave
// Params     : echo.Group
func (r *AppRouter) LeaveRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-leave", r.leaveCtr.CreateLeaveRequest, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-leave-bonus", r.leaveCtr.CreateLeaveBonus, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-leave-info", r.leaveCtr.GetLeaveInfo, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-leave-info-all-user", r.leaveCtr.GetLeaveInfoAllUser, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-leave-history", r.leaveCtr.GetLeaveHistory, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-leave-requests", r.leaveCtr.GetLeaveRequests, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-leave", r.leaveCtr.RemoveLeaveRequest, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/cron-leave-bonus", r.leaveCtr.CronLeaveBonus, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/cron-leave-start", r.leaveCtr.CronLeaveStart, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/cron-leave-stop", r.leaveCtr.CronLeaveStop, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-entries-leave", r.leaveCtr.GetEntriesLeave, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/remove-cron-leave", r.leaveCtr.RemoveCronLeave, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/import-bonuses", r.leaveCtr.ImportBonuses, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/download-template", r.leaveCtr.DownloadTemplate, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.GET("/export-excel", r.leaveCtr.ExportExcel, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-leave-bonuses", r.leaveCtr.GetLeaveBonuses, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-leave-bonus", r.leaveCtr.GetLeaveBonus, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/edit-leave-bonus", r.leaveCtr.EditLeaveBonuses, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/remove-leave-bonus", r.leaveCtr.RemoveLeaveBonus, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/update-hour-remaining-leave", r.leaveCtr.UpdateHourRemainingLeave, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/add-expire-leave", r.leaveCtr.AddExpireLeave, isLoggedIn, r.userMw.InitUserProfile)
}

// LeaveRoute : create route for group /leave
// Params     : echo.Group
func (r *AppRouter) UserProjectRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-user-project", r.uprjCtr.CreateUserProject, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-user-project", r.uprjCtr.GetUserProject, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-user-from-project", r.uprjCtr.RemoveUserProject, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-user-ids-managed-by-manager", r.uprjCtr.GetUserIdsManagedByManager, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-projects-user-join", r.uprjCtr.GetProjectsUserJoin, isLoggedIn, r.userMw.InitUserProfile)
}

func (r *AppRouter) SettingRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/edit-organization-email", r.orgCtr.EditOrganizationEmail, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/get-organization-setting", r.orgCtr.GetOrganizationSetting, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/edit-expiration-reset-day-off", r.orgCtr.EditExpirationResetDayOff, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)

	g.POST("/branch/create-branch", r.branchCtr.CreateBranch, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/branch/edit-branch", r.branchCtr.EditBranch, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/branch/get-branches", r.branchCtr.GetBranches, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/branch/remove-branch", r.branchCtr.RemoveBranch, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/branch/sort-prioritization", r.branchCtr.SortPrioritization, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)

	g.POST("/job-title/create-job-title", r.jobTitleCtr.CreateJobTitle, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/job-title/edit-job-title", r.jobTitleCtr.EditJobTitle, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/job-title/get-job-titles", r.jobTitleCtr.GetJobTitles, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/job-title/remove-job-title", r.jobTitleCtr.RemoveJobTitle, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/job-title/sort-prioritization", r.jobTitleCtr.SortPrioritization, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)

	g.POST("/technology/create-technology", r.techCtr.CreateTechnology, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/technology/edit-technology", r.techCtr.EditTechnology, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/technology/get-technologies", r.techCtr.GetTechnologies, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/technology/remove-technology", r.techCtr.RemoveTechnology, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/technology/sort-prioritization", r.techCtr.SortPrioritization, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
}

func (r *AppRouter) UserTechnologyRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-user-technologies", r.utechCtr.CreateUserTechnologies, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-technology-of-user", r.utechCtr.GetTechnologiesOfUser, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-technology-of-user", r.utechCtr.RemoveTechnologiesOfUser, isLoggedIn, r.userMw.InitUserProfile)
}

func (r *AppRouter) OvertimeRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-overtime-request", r.overtimeCtr.CreateOvertimeRequest, isLoggedIn, r.userMw.InitUserProfile)
	g.POST(
		"/update-overtime-request-status",
		r.overtimeCtr.UpdateOvertimeRequestStatus,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckGeneralManager,
	)
	g.POST(
		"/get-overtime-requests",
		r.overtimeCtr.GetOvertimeRequests,
		isLoggedIn,
		r.userMw.InitUserProfile,
	)
	g.POST("/get-overtime-request", r.overtimeCtr.GetOvertimeRequestById, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-overtime-request", r.overtimeCtr.EditOvertimeRequest, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-emails-gm-and-pm", r.overtimeCtr.GetEmailsGMAndPM, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/export-excel", r.overtimeCtr.ExportExcel, isLoggedIn, r.userMw.InitUserProfile)
	g.POST(
		"/create-overtime-weight",
		r.overtimeCtr.CreateOvertimeWeight,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckGeneralManager,
	)
	g.POST(
		"/edit-overtime-weight",
		r.overtimeCtr.EditOvertimeWeight,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckGeneralManager,
	)
	g.POST(
		"/get-overtime-weight",
		r.overtimeCtr.GetOvertimeWeight,
		isLoggedIn,
		r.userMw.InitUserProfile,
		r.userMw.CheckGeneralManager,
	)
}

func (r *AppRouter) StatisticRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/general", r.userCtr.Statistic, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/project-statistic-detail", r.uprjCtr.ProjectStatistic, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/technology-statistic-detail", r.utechCtr.TechnologyStatisticDetail, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/job-title-statistic-detail", r.jobTitleCtr.JobTitleStatisticDetail, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/branch-statistic-detail", r.branchCtr.BranchStatisticDetail, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/jp-level-statistic-detail", r.userCtr.JpLevelStatisticDetail, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

func (r *AppRouter) HolidayRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-holiday", r.holidayCtr.CreateHoliday, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/edit-holiday", r.holidayCtr.EditHoliday, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
	g.POST("/get-holidays", r.holidayCtr.GetHolidays, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-holiday", r.holidayCtr.RemoveHoliday, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

func (r *AppRouter) NotificationRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/edit-notification-status-read", r.notificationCtr.EditNotificationStatusRead, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-notification-status", r.notificationCtr.EditNotificationStatus, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-notifications", r.notificationCtr.GetNotifications, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-total-notifications-unread", r.notificationCtr.GetTotalNotificationsUnread, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-notification", r.notificationCtr.RemoveNotification, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-fcm-token", r.fcmTokenCtr.CreateFcmToken, isLoggedIn)
	g.POST("/get-fcm-tokens", r.fcmTokenCtr.GetFcmTokens, isLoggedIn)
	g.POST("/send-noti-request-by-id", r.notificationCtr.SendNotiRequestById, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/notification-event-remind", r.notificationCtr.NotiEventRemind, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/cron-notify-event-remind", r.notificationCtr.CronNotiEventRemind, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/cron-notify-event-remind-start", r.notificationCtr.CronNotifyEventRemindStart, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckAllManager)
}

func (r *AppRouter) RecruitmentRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/create-job", r.recruitmentCtr.CreateJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-detail-job", r.recruitmentCtr.CreateDetailJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/update-detail-job", r.recruitmentCtr.UpdateDetailJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-job", r.recruitmentCtr.EditJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-jobs", r.recruitmentCtr.GetJobs, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-job", r.recruitmentCtr.GetJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-job", r.recruitmentCtr.RemoveJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/upload-cv", r.recruitmentCtr.UploadCv, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-cv", r.recruitmentCtr.CreateCv, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-cv-by-id", r.recruitmentCtr.GetCvById, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/update-cv", r.recruitmentCtr.UpdateCv, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-cvs", r.recruitmentCtr.GetCvs, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-cv", r.recruitmentCtr.RemoveCv, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-cv-comment", r.recruitmentCtr.CreateCvComment, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-cv-comment", r.recruitmentCtr.EditCvComment, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-cv-comments", r.recruitmentCtr.GetCvComments, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-cv-comment", r.recruitmentCtr.RemoveCvComment, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-logs-cv-status", r.recruitmentCtr.GetLogCvStatus, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/upload-detail-job", r.recruitmentCtr.UploadDetailJob, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-detail-job-file", r.recruitmentCtr.GetDetailJobFile, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/remove-detail-job-file", r.recruitmentCtr.RemoveDetailJobFile, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-log-cv-status", r.recruitmentCtr.CreateLogCvStatus, isLoggedIn, r.userMw.InitUserProfile)
}

func (r *AppRouter) UserPermissionRoute(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})

	g.POST("/edit-permission", r.userPermissionCtr.EditPermission, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
	g.POST("/get-permissions", r.userPermissionCtr.GetPermissions, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-user-permissions", r.userPermissionCtr.GetUserPermissions, isLoggedIn, r.userMw.InitUserProfile, r.userMw.CheckGeneralManager)
}

func (r *AppRouter) AdminRouter(g *echo.Group) {
	// keyTokenAuth := utils.GetKeyToken()
	// isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
	// 	SigningKey: []byte(keyTokenAuth),
	// })
	g.POST("/setting-organization-module", r.adminCtr.SettingOrgModule)
	g.POST("/setting-organization-fuctions", r.adminCtr.SettingOrgFunctions)
	g.POST("/get-moudles", r.adminCtr.GetModules)
	g.POST("/setting-user-fuctions", r.adminCtr.SettingUserFunctions)
}

func (r *AppRouter) AssetRouter(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})
	g.POST("/get-assets-list", r.assetCtr.GetAssetList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-asset-type", r.assetCtr.CreateAssetType, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-assets-log", r.assetCtr.GetAssetLog, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-asset-type-list", r.assetCtr.GetAssetTypeList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-user-request-asset", r.assetCtr.CreateRequestAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-asset", r.assetCtr.CreateAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-user-requests-asset", r.assetCtr.GetUserRequestAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/edit-user-requests-asset", r.assetCtr.EditUserRequestAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/update-asset", r.assetCtr.UpdateAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-asset-by-id", r.assetCtr.GetAssetByID, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/delete-asset-by-id", r.assetCtr.DeleteAssetByID, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/delete-asset-type", r.assetCtr.DeleteAssetTypeByID, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-asset-log", r.assetCtr.CreateAssetLog, isLoggedIn, r.userMw.InitUserProfile)
	g.GET("/download-template-import-asset", r.assetCtr.DownloadTemplateImportAsset, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/import-excel-asset", r.assetCtr.ImportExcelAsset, isLoggedIn, r.userMw.InitUserProfile)
}

func (r *AppRouter) ContractRouter(g *echo.Group) {
	keyTokenAuth := utils.GetKeyToken()
	isLoggedIn := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(keyTokenAuth),
	})
	g.POST("/get-contract-current-list", r.contractCtr.GetContractCurrentList, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/get-contract-by-user", r.contractCtr.GetContractByUser, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-contract", r.contractCtr.CreateContract, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/create-contract-type", r.contractCtr.CreateContractType, isLoggedIn)
	g.POST("/get-contract-type-by-id", r.contractCtr.GetContractTypeByID, isLoggedIn)
	g.POST("/get-contract-type-list", r.contractCtr.GetContractTypeList, isLoggedIn)
	g.POST("/edit-contract-type", r.contractCtr.EditContractType, isLoggedIn)
	g.POST("/delete-contract-type", r.contractCtr.DeleteContractTypeByID, isLoggedIn)
	g.POST("/delete-contract", r.contractCtr.DeleteContractByID, isLoggedIn)
	g.POST("/preview-contract", r.contractCtr.PreviewContract, isLoggedIn, r.userMw.InitUserProfile)
	g.POST("/delete-preview-contract", r.contractCtr.DeletePreviewContract, isLoggedIn, r.userMw.InitUserProfile)
}
