package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"sns-server/internal/models"
	"sns-server/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// RegisterInput represents user registration input
type RegisterInput struct {
	Username string
	Email    string
	Password string
	Name     string
	Bio      string
}

// RegisterResult represents user registration result
type RegisterResult struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// DeleteUserInput represents user deletion input
type DeleteUserInput struct {
	UserID uint `json:"userId"`
}

// DeleteUserResponse represents user deletion response
type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateUserInput represents user update input
type UpdateUserInput struct {
	UserID   uint                   `json:"userId"`
	Updates  map[string]interface{} `json:"updates"`
}

// UpdateUserResponse represents user update response
type UpdateUserResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *models.User `json:"user"`
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

// Register creates a new user with validation
func (s *UserService) Register(input RegisterInput) (*RegisterResult, error) {
	// Validate required fields
	if strings.TrimSpace(input.Username) == "" {
		return nil, fmt.Errorf("Username is required")
	}
	if strings.TrimSpace(input.Email) == "" {
		return nil, fmt.Errorf("Email is required")
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, fmt.Errorf("Password is required")
	}

	// パスワード強度チェック
	if len(input.Password) < 6 {
		return nil, fmt.Errorf("Password must be at least 6 characters")
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password, // TODO: ハッシュ化
		Name:     input.Name,
		Bio:      input.Bio,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, fmt.Errorf("Failed to create user: %v", err)
	}

	// Generate token (same as existing implementation)
	token := "temp_token_" + strconv.Itoa(int(user.ID))

	return &RegisterResult{
		Token: token,
		User:  &user,
	}, nil
}

// EnsureDefaultUser creates or gets a default user for testing
func (s *UserService) EnsureDefaultUser() (*models.User, error) {
	user, err := s.userRepo.FindByUsername("defaultuser")
	if err == nil {
		return user, nil
	}

	// Create default user
	defaultUser := models.User{
		Username: "defaultuser",
		Email:    "default@example.com",
		Password: "password",
		Name:     "Default User",
	}

	if err := s.userRepo.Create(&defaultUser); err != nil {
		return nil, err
	}

	return &defaultUser, nil
}

// GetUserByID returns a user by ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// DeleteUser deletes a user with full cascade deletion
func (s *UserService) DeleteUser(input DeleteUserInput) (*DeleteUserResponse, error) {
	// ユーザーIDの検証
	if input.UserID == 0 {
		return nil, fmt.Errorf("User ID is required")
	}

	// ユーザーの存在確認
	exists, err := s.userRepo.Exists(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("Failed to check user existence: %v", err)
	}
	
	if !exists {
		return nil, fmt.Errorf("User not found")
	}

	// ユーザー削除（カスケード削除含む）
	err = s.userRepo.Delete(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("Failed to delete user: %v", err)
	}

	return &DeleteUserResponse{
		Success: true,
		Message: fmt.Sprintf("User %d deleted successfully with all associated data", input.UserID),
	}, nil
}

// UpdateUser updates user with validation
func (s *UserService) UpdateUser(input UpdateUserInput) (*UpdateUserResponse, error) {
	// ユーザーIDの検証
	if input.UserID == 0 {
		return nil, fmt.Errorf("User ID is required")
	}

	// ユーザーの存在確認
	exists, err := s.userRepo.Exists(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("Failed to check user existence: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("User not found")
	}

	// 更新データの検証
	validatedUpdates := make(map[string]interface{})
	
	for field, value := range input.Updates {
		switch field {
		case "username":
			username := strings.TrimSpace(fmt.Sprintf("%v", value))
			if username == "" {
				return nil, fmt.Errorf("Username cannot be empty")
			}
			// ユーザー名の文字数制限チェック（2文字以上50文字以下）
			if len([]rune(username)) < 2 || len([]rune(username)) > 50 {
				return nil, fmt.Errorf("Username must be between 2 and 50 characters")
			}
			// ユーザー名の特殊文字チェック
			if matched, _ := regexp.MatchString(`[^a-zA-Z0-9_\p{L}]`, username); matched {
				return nil, fmt.Errorf("Username contains invalid characters")
			}
			validatedUpdates[field] = username
			
		case "email":
			email := strings.TrimSpace(fmt.Sprintf("%v", value))
			if email == "" {
				return nil, fmt.Errorf("Email cannot be empty")
			}
			// 簡単なメール形式チェック
			if matched, _ := regexp.MatchString(`^[^@]+@[^@]+\.[^@]+$`, email); !matched {
				return nil, fmt.Errorf("Invalid email format")
			}
			validatedUpdates[field] = email
			
		case "name":
			name := strings.TrimSpace(fmt.Sprintf("%v", value))
			validatedUpdates[field] = name
			
		case "bio":
			bio := strings.TrimSpace(fmt.Sprintf("%v", value))
			validatedUpdates[field] = bio
			
		default:
			// 許可されていないフィールドは無視
			continue
		}
	}

	if len(validatedUpdates) == 0 {
		return nil, fmt.Errorf("No valid fields to update")
	}

	// 重複チェック（ユーザー名・メール）
	if username, ok := validatedUpdates["username"]; ok {
		existingUser, err := s.userRepo.FindByUsername(username.(string))
		if err == nil && existingUser.ID != input.UserID {
			return nil, fmt.Errorf("Username already exists")
		}
	}

	if email, ok := validatedUpdates["email"]; ok {
		// メールアドレスの重複チェック
		existingUser, err := s.userRepo.FindByUsernameOrEmail("", email.(string))
		if err == nil && existingUser.ID != input.UserID {
			return nil, fmt.Errorf("Email already exists")
		}
	}

	// 更新実行
	err = s.userRepo.Update(input.UserID, validatedUpdates)
	if err != nil {
		return nil, fmt.Errorf("Failed to update user: %v", err)
	}

	// 更新後のユーザー情報を取得
	updatedUser, err := s.userRepo.GetByID(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve updated user: %v", err)
	}

	return &UpdateUserResponse{
		Success: true,
		Message: fmt.Sprintf("User %d updated successfully", input.UserID),
		User:    updatedUser,
	}, nil
}
