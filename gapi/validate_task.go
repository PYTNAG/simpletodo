package gapi

import (
	"fmt"

	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func validateCreateTask(req *pb.CreateTaskRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateId(req.GetListId()); err != nil {
		violations = append(violations, fieldViolation("list_id", err))
	}

	if err := validation.ValidateStringParam(req.GetTaskText()); err != nil {
		violations = append(violations, fieldViolation("task_text", err))
	}

	if req.ParentTaskId != nil {
		if err := validation.ValidateId(req.GetParentTaskId()); err != nil {
			violations = append(violations, fieldViolation("parent_task_id", err))
		}
	}

	return
}

func validateGetTasks(req *pb.GetTasksRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateId(req.GetListId()); err != nil {
		violations = append(violations, fieldViolation("list_id", err))
	}

	return
}

func validateDeleteTask(req *pb.DeleteTaskRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateId(req.GetTaskId()); err != nil {
		violations = append(violations, fieldViolation("task_id", err))
	}

	return
}

func validateUpdateTask(req *pb.UpdateTaskRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateId(req.GetTaskId()); err != nil {
		violations = append(violations, fieldViolation("task_id", err))
	}

	if req.GetType() == pb.UpdateType_UNSET {
		violations = append(violations, fieldViolation("type", fmt.Errorf("you must specify update type (TEXT or CHECK)")))
	}

	if req.GetType() == pb.UpdateType_TEXT {
		if req.NewText == nil {
			violations = append(violations, fieldViolation("new_text", fmt.Errorf("you must specify new text for text update")))
		}

		if err := validation.ValidateStringParam(req.GetNewText()); err != nil {
			violations = append(violations, fieldViolation("type", err))
		}
	}

	return
}
