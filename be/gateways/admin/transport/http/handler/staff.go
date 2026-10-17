package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	exception "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/exception"
	staffRequest "github.com/kainguyen/badminton-hub/be/gateways/admin/transport/http/request"
	httpMiddleware "github.com/kainguyen/badminton-hub/be/platform/http/middleware"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
	platformService "github.com/kainguyen/badminton-hub/be/platform/service"
	identityV1 "github.com/kainguyen/badminton-hub/contracts/gen/identity/v1"
)

func RegisterStaff(engine *gin.Engine, client identityV1.IdentityAdminServiceClient) {
	// Build the handler with the owning Identity client.
	handler := NewStaffHandler(client)

	// Register authenticated staff routes.
	group := engine.Group("/api/v1/admin/staff")
	group.Use(httpMiddleware.RequireBearer)
	group.GET("/", handler.List)
	group.POST("/", handler.Create)
	group.GET("/:staff_id", handler.Get)
	group.PUT("/:staff_id", handler.Update)
	group.DELETE("/:staff_id", handler.Delete)
	group.PUT("/:staff_id/access", handler.UpdateAccess)
}

type StaffHandler struct {
	client identityV1.IdentityAdminServiceClient
}

func NewStaffHandler(client identityV1.IdentityAdminServiceClient) *StaffHandler {
	return &StaffHandler{client: client}
}

func (h *StaffHandler) List(c *gin.Context) {
	params, err := request.ListQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	result, err := h.client.ListStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.ListStaffRequest{
			AccessToken: request.BearerToken(c),
			Limit:       params.Limit,
			Offset:      params.Offset,
			Search:      params.Search,
			CreatedFrom: params.CreatedFrom,
			CreatedTo:   params.CreatedTo,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *StaffHandler) Create(c *gin.Context) {
	var input staffRequest.CreateStaffRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.CreateStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.CreateStaffRequest{
			AccessToken: request.BearerToken(c), Email: input.Email, Password: input.Password,
			FullName: input.FullName, IsAdmin: input.IsAdmin, RoleId: input.RoleID,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusCreated)
}

func (h *StaffHandler) Get(c *gin.Context) {
	result, err := h.client.GetStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.GetStaffRequest{AccessToken: request.BearerToken(c), StaffId: request.Param(c, "staff_id")},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *StaffHandler) Update(c *gin.Context) {
	var input staffRequest.UpdateStaffRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.UpdateStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.UpdateStaffRequest{
			AccessToken: request.BearerToken(c), StaffId: request.Param(c, "staff_id"),
			Email: input.Email, FullName: input.FullName, Status: input.Status,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}

func (h *StaffHandler) Delete(c *gin.Context) {
	result, err := h.client.DeleteStaff(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.DeleteStaffRequest{AccessToken: request.BearerToken(c), StaffId: request.Param(c, "staff_id")},
	)
	exception.WriteGRPC(c, result, err, http.StatusNoContent)
}

func (h *StaffHandler) UpdateAccess(c *gin.Context) {
	var input staffRequest.UpdateStaffAccessRequest
	if request.BindJSON(c, &input) {
		return
	}
	result, err := h.client.UpdateStaffAccess(
		request.RPCContext(c, platformService.AdminGateway),
		&identityV1.UpdateStaffAccessRequest{
			AccessToken: request.BearerToken(c), StaffId: request.Param(c, "staff_id"),
			IsAdmin: input.IsAdmin, RoleId: input.RoleID,
		},
	)
	exception.WriteGRPC(c, result, err, http.StatusOK)
}
