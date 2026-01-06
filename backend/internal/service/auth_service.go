/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:31:10
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 14:30:59
 * @FilePath            : frp-web-testbackendinternalserviceauth_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"errors"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/util"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo      *repository.UserRepository
	eventNotifier *SystemEventNotifier
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

// SetEventNotifier 设置系统事件通知器
func (s *AuthService) SetEventNotifier(notifier *SystemEventNotifier) {
	s.eventNotifier = notifier
}

func (s *AuthService) Login(username, password, clientIP string) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if s.eventNotifier != nil {
			go s.eventNotifier.NotifyLoginFailed(username, clientIP)
		}
		return "", nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if s.eventNotifier != nil {
			go s.eventNotifier.NotifyLoginFailed(username, clientIP)
		}
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := util.GenerateToken(user.ID, user.Username, config.GlobalConfig.JWT.Secret, config.GlobalConfig.JWT.ExpireHours)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) GetProfile(userID uint) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// ChangePassword 修改用户密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 获取用户信息
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}
