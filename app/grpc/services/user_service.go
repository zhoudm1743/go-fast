package services

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/zhoudm1743/go-fast/app/grpc/proto/user"
	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// UserServiceServer 实现 userpb.UserServiceServer。
type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
}

// GetUser 根据 ID 查询用户。
func (s *UserServiceServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserReply, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id 不能为空")
	}

	var user models.User
	err := facades.DB().Query().Where("id = ?", req.Id).First(&user)
	if err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
		}
		facades.Log().Errorf("grpc get user: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	return toUserReply(&user), nil
}

// ListUsers 分页查询用户列表。
func (s *UserServiceServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersReply, error) {
	page := int(req.Page)
	size := int(req.Size)
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	q := facades.DB().Query().Model(&models.User{}).Order("created_at DESC")
	if req.Email != "" {
		q = q.Where("email LIKE ?", "%"+req.Email+"%")
	}

	var total int64
	q.Count(&total)

	var users []models.User
	if err := q.Paginate(page, size).Find(&users); err != nil {
		facades.Log().Errorf("grpc list users: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	replies := make([]*userpb.UserReply, 0, len(users))
	for i := range users {
		replies = append(replies, toUserReply(&users[i]))
	}
	return &userpb.ListUsersReply{
		Users: replies,
		Total: total,
		Page:  int32(page),
		Size:  int32(size),
	}, nil
}

// CreateUser 创建用户。
func (s *UserServiceServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserReply, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "name、email、password 均不能为空")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	if err := facades.DB().Query().Create(user); err != nil {
		if errors.Is(err, contracts.ErrDuplicatedKey) {
			return nil, status.Error(codes.AlreadyExists, "邮箱已被注册")
		}
		facades.Log().Errorf("grpc create user: %v", err)
		return nil, status.Error(codes.Internal, "创建失败")
	}
	return toUserReply(user), nil
}

// UpdateUser 更新用户。
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserReply, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id 不能为空")
	}

	var user models.User
	if err := facades.DB().Query().Where("id = ?", req.Id).First(&user); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
		}
		return nil, status.Error(codes.Internal, "查询失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if len(updates) > 0 {
		if err := facades.DB().Query().Model(&user).Updates(updates); err != nil {
			return nil, status.Error(codes.Internal, "更新失败")
		}
	}
	return toUserReply(&user), nil
}

// DeleteUser 删除用户。
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserReply, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id 不能为空")
	}
	if err := facades.DB().Query().Where("id = ?", req.Id).Delete(&models.User{}); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
		}
		return nil, status.Error(codes.Internal, "删除失败")
	}
	return &userpb.DeleteUserReply{Success: true}, nil
}

// ── 辅助函数 ──────────────────────────────────────────────────────────

func toUserReply(u *models.User) *userpb.UserReply {
	return &userpb.UserReply{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: strconv.FormatInt(u.CreatedAt, 10),
		UpdatedAt: strconv.FormatInt(u.UpdatedAt, 10),
	}
}
