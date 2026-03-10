package auth

import (
	"context"
	"errors"
	"go-fiber-stater-kit/pkg/common"
	"go-fiber-stater-kit/pkg/core"
	"time"

	"github.com/redis/go-redis/v9"
)

type UseCase interface {
	Register(dto *RegisterDTO) (*AuthResponse, error)
	Login(dto *LoginDTO) (*AuthResponse, error)
	RefreshToken(dto *RefreshTokenDTO) (*AuthResponse, error)
	Logout(userID string, accessToken string) error
}

type useCase struct {
	Repo        Repository
	RedisClient *redis.Client
}

func NewUseCase(repo Repository, redisClient *redis.Client) UseCase {
	return &useCase{
		Repo:        repo,
		RedisClient: redisClient,
	}
}

func (u *useCase) Register(dto *RegisterDTO) (*AuthResponse, error) {
	// Check if email already exists
	existingUser, _ := u.Repo.FindByEmail(dto.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	existingUser, _ = u.Repo.FindByUsername(dto.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := common.HashPassword(dto.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user, err := u.Repo.CreateUser(dto, hashedPassword)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate token pair
	accessToken, refreshToken, err := core.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Hash and store refresh token
	refreshHash, err := common.HashPassword(refreshToken)
	if err != nil {
		return nil, errors.New("failed to hash refresh token")
	}
	if err := u.Repo.UpdateRefreshTokenHash(user.ID, refreshHash); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (u *useCase) Login(dto *LoginDTO) (*AuthResponse, error) {
	// Try to find user by username (which could also be an email)
	user, err := u.Repo.FindByUsername(dto.Username)
	if err != nil {
		// Fallback: try finding by email
		user, err = u.Repo.FindByEmail(dto.Username)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Compare passwords
	if !common.ComparePasswords(user.Password, dto.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate token pair
	accessToken, refreshToken, err := core.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Hash and store refresh token
	refreshHash, err := common.HashPassword(refreshToken)
	if err != nil {
		return nil, errors.New("failed to hash refresh token")
	}
	if err := u.Repo.UpdateRefreshTokenHash(user.ID, refreshHash); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	// Cache user data in Redis
	ctx := context.Background()
	u.RedisClient.Set(ctx, "user:"+user.ID, user.Email, 15*time.Minute)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (u *useCase) RefreshToken(dto *RefreshTokenDTO) (*AuthResponse, error) {
	// Validate the refresh token
	claims, err := core.ValidateRefreshToken(dto.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Get the stored refresh token hash
	storedHash, err := u.Repo.GetRefreshTokenHash(claims.UserID)
	if err != nil || storedHash == nil {
		return nil, errors.New("refresh token not found")
	}

	// Verify the refresh token matches the stored hash
	if !common.ComparePasswords(*storedHash, dto.RefreshToken) {
		return nil, errors.New("refresh token mismatch")
	}

	// Find the user to get latest data
	user, err := u.Repo.FindByEmail(claims.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate new token pair
	accessToken, refreshToken, err := core.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Hash and store new refresh token
	refreshHash, err := common.HashPassword(refreshToken)
	if err != nil {
		return nil, errors.New("failed to hash refresh token")
	}
	if err := u.Repo.UpdateRefreshTokenHash(user.ID, refreshHash); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (u *useCase) Logout(userID string, accessToken string) error {
	// Blacklist the access token in Redis (expire when token expires)
	ctx := context.Background()
	u.RedisClient.Set(ctx, "blacklist:"+accessToken, "1", core.AccessTokenExpiry)

	// Clear refresh token hash from database
	return u.Repo.ClearRefreshTokenHash(userID)
}
