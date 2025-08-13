package todo

import (
	"fmt"
	"runtime"
	"slices"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

const (
	owned  = "owned"  // Return only user owned lists
	shared = "shared" // Return lists shared with user
	all    = "all"    // Return both of the above
	admin  = "admin"  // Return every single list
)

func (controller *TodoController) ReadLists(ctx *gin.Context) {
	requesterId, requesterUsername, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	show := ctx.DefaultQuery("show", all)
	if !slices.Contains([]string{owned, shared, all, admin}, show) {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			fmt.Sprintf("query param 'show' was: %v", show),
			requesterId,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	reqUser, err := database.GetUserById(controller.db, requesterId, ctx)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			requesterUsername,
			ctx.ClientIP(),
		)
		return
	}

	if show == admin && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			"all lists",
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	lists := &[]db.List{}
	var switchErr error
	switch show {
	case owned:
		*lists, switchErr = controller.db.GetListsByOwnerId(ctx, reqUser.ID)
	case shared:
		*lists, switchErr = controller.db.GetListsBySharedUserId(ctx, reqUser.ID)
	case all:
		*lists, switchErr = controller.db.GetListsAccessibleByUserId(ctx, reqUser.ID)
	case admin:
		*lists, switchErr = controller.db.GetLists(ctx)
	}
	if switchErr != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			fmt.Sprintf("failed to get lists with show: %v", show),
			file,
			line,
			switchErr,
			ctx,
		)
		return
	}
	listIds := make([]string, 0, len(*lists))
	for _, list := range *lists {
		listIds = append(listIds, list.ID)
	}

	todos, err := controller.db.GetTodosByListIds(ctx, listIds)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get todos", file, line, err, ctx)
		return
	}

	todoMap := make(map[string][]db.Todo)
	for _, todo := range todos {
		todoMap[todo.ListID] = append(todoMap[todo.ListID], todo)
	}

	response := make([]map[string]any, 0, len(*lists))
	for _, list := range *lists {
		item := map[string]any{
			"id":          list.ID,
			"user_id":     list.UserID,
			"title":       list.Title,
			"description": list.Description,
			"created_at":  list.CreatedAt,
			"updated_at":  list.UpdatedAt,
			"todos":       todoMap[list.ID],
		}
		response = append(response, item)
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventRead,
		reqUser,
		*lists,
		nil,
		logging.ObjectEventSubList,
	)
	ctx.JSON(200, gin.H{"status": "ok", "lists": response})
}
