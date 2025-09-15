package service

import (
	"fmt"
	"sns-server/internal/models"
	"sns-server/internal/repository"
)

type FollowService struct {
	FollowRepo *repository.FollowRepository
	userRepo   *repository.UserRepository
}

func NewFollowService(followRepo *repository.FollowRepository, userRepo *repository.UserRepository) *FollowService {
	return &FollowService{
		FollowRepo: followRepo,
		userRepo:   userRepo,
	}
}

type FollowUserInput struct {
	FolloweeID uint `json:"followeeId"`
}

type UnfollowUserInput struct {
	FolloweeID uint `json:"followeeId"`
}

type FollowUserResponse struct {
	ID        uint         `json:"id"`
	Follower  *models.User `json:"follower"`
	Followee  *models.User `json:"followee"`
	CreatedAt string       `json:"createdAt"`
}

type UnfollowUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *FollowService) FollowUser(input FollowUserInput) (*FollowUserResponse, error) {
	// デフォルトユーザーを取得（認証システムが未実装のため）
	follower, err := s.getDefaultUser()
	if err != nil {
		return nil, fmt.Errorf("Failed to get current user: %v", err)
	}

	// フォロー対象のユーザーを取得
	followee, err := s.userRepo.GetByID(input.FolloweeID)
	if err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// 自分自身をフォローしようとしていないかチェック
	if follower.ID == followee.ID {
		return nil, fmt.Errorf("Cannot follow yourself")
	}

	// 既にフォローしているかチェック
	existing, checkErr := s.FollowRepo.GetByFollowerAndFollowee(follower.ID, followee.ID)
	if checkErr == nil && existing != nil {
		// レコードが見つかった場合、既にフォローしている
		return nil, fmt.Errorf("Already following this user - DEBUG: Found existing follow ID %d", existing.ID)
	}
	// エラーが出た場合（通常はrecord not found）は続行
	if checkErr != nil {
		// デバッグ情報追加
		fmt.Printf("DEBUG: Follow check error for user %d -> %d: %v\n", follower.ID, followee.ID, checkErr)
	}

	// フォロー関係を作成
	follow := &models.Follow{
		FollowerID: follower.ID,
		FolloweeID: followee.ID,
	}

	err = s.FollowRepo.Create(follow)
	if err != nil {
		return nil, fmt.Errorf("Failed to create follow relationship: %v", err)
	}

	// レスポンス用にユーザー情報を再取得
	follower, _ = s.userRepo.GetByID(follower.ID)
	followee, _ = s.userRepo.GetByID(followee.ID)

	return &FollowUserResponse{
		ID:        follow.ID,
		Follower:  follower,
		Followee:  followee,
		CreatedAt: follow.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *FollowService) UnfollowUser(input UnfollowUserInput) (*UnfollowUserResponse, error) {
	// デフォルトユーザーを取得（認証システムが未実装のため）
	follower, err := s.getDefaultUser()
	if err != nil {
		return nil, fmt.Errorf("Failed to get current user: %v", err)
	}

	// フォロー解除対象のユーザーを取得
	followee, err := s.userRepo.GetByID(input.FolloweeID)
	if err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	// フォロー関係を取得
	follow, err := s.FollowRepo.GetByFollowerAndFollowee(follower.ID, followee.ID)
	if err != nil || follow == nil {
		return nil, fmt.Errorf("Not following this user")
	}

	// フォロー関係を削除
	err = s.FollowRepo.Delete(follow.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to unfollow user: %v", err)
	}

	return &UnfollowUserResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully unfollowed %s", followee.Username),
	}, nil
}

// getDefaultUser returns the first available user for operations (testing compatibility)
func (s *FollowService) getDefaultUser() (*models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("Failed to get users: %v", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("No users found")
	}

	return &users[0], nil
}
