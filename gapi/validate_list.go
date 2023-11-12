package gapi

import (
	"github.com/PYTNAG/simpletodo/pb"
	"github.com/PYTNAG/simpletodo/validation"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func validateCreateList(req *pb.CreateListRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validation.ValidateUserId(req.GetUserId()); err != nil {
		violations = append(violations, fieldViolation("user_id", err))
	}

	if err := validation.ValidateHeader(req.GetHeader()); err != nil {
		violations = append(violations, fieldViolation("header", err))
	}

	return
}
