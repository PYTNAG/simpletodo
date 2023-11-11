package gapi

import (
	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return
}

func validateDeleteUserRequest(req *pb.DeleteUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUserId(req.GetUserId()); err != nil {
		violations = append(violations, fieldViolation("user_id", err))
	}

	return
}

func validateRehashUserRequest(req *pb.RehashUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUserId(req.GetUserId()); err != nil {
		violations = append(violations, fieldViolation("user_id", err))
	}

	if err := validation.ValidatePassword(req.GetNewPassword()); err != nil {
		violations = append(violations, fieldViolation("new_password", err))
	}

	return
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return
}
