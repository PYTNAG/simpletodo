package gapi

import (
	"context"
	"database/sql"

	db "github.com/PYTNAG/simpletodo/db/sqlc"
	dbtypes "github.com/PYTNAG/simpletodo/db/types"
	"github.com/PYTNAG/simpletodo/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateCreateTask(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if err := s.isListAccessAllowed(ctx, payload, req.GetListId()); err != nil {
		return nil, err
	}

	arg := db.AddTaskParams{
		ListID:     req.GetListId(),
		ParentTask: dbtypes.NewNullInt32(req.GetParentTaskId(), req.GetParentTaskId() > 0),
		Task:       req.TaskText,
	}

	task, err := s.store.AddTask(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %s", err)
	}

	response := &pb.CreateTaskResponse{
		TaskId: task.ID,
	}

	return response, nil
}

func (s *Server) GetTasks(req *pb.GetTasksRequest, stream pb.SimpleTODO_GetTasksServer) error {
	payload, err := s.authorizeUser(stream.Context())
	if err != nil {
		return unauthenticatedError(err)
	}

	violations := validateGetTasks(req)
	if violations != nil {
		return invalidArgumentError(violations)
	}

	if err := s.isListAccessAllowed(stream.Context(), payload, req.GetListId()); err != nil {
		return err
	}

	tasks, err := s.store.GetTasks(stream.Context(), req.GetListId())
	if err != nil && err != sql.ErrNoRows {
		return status.Errorf(codes.Internal, "failed to get tasks: %s", err)
	}

	for _, task := range tasks {
		var taskId *int32 = nil
		if task.ParentTask.Valid {
			taskId = &task.ParentTask.Int32
		}

		msg := &pb.Task{
			TaskId:       task.ID,
			ListId:       task.ListID,
			Text:         task.Task,
			Check:        task.Complete,
			ParentTaskId: taskId,
		}

		if err = stream.Send(msg); err != nil {
			return status.Errorf(codes.Internal, "failed to send task: %s", err)
		}
	}

	return nil
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateDeleteTask(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if err := s.isTaskAccessAllowed(ctx, payload, req.GetTaskId()); err != nil {
		return nil, err
	}

	err = s.store.DeleteTask(ctx, req.GetTaskId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "there is no task %d", req.GetTaskId())
		}

		return nil, status.Errorf(codes.Internal, "failed to delete task: %s", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*emptypb.Empty, error) {
	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateTask(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if err := s.isTaskAccessAllowed(ctx, payload, req.GetTaskId()); err != nil {
		return nil, err
	}

	switch req.Type {
	case pb.UpdateType_CHECK:
		if err := s.store.ToggleTask(ctx, req.TaskId); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to toggle task: %s", err)
		}

	case pb.UpdateType_TEXT:
		if req.NewText == nil {
			return nil, status.Errorf(codes.InvalidArgument, "you must provide text for text update")
		}

		params := db.UpdateTaskTextParams{
			ID:   req.GetTaskId(),
			Task: req.GetNewText(),
		}

		if err := s.store.UpdateTaskText(ctx, params); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update task text: %s", err)
		}

	default:
		return nil, status.Error(codes.InvalidArgument, "update type must be CHECK or TEXT")
	}

	return &emptypb.Empty{}, nil
}
