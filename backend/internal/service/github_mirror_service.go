/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-21 14:01:53
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-21 14:02:08
 * @FilePath            : frp-web-testbackendinternalservicegithub_mirror_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"strings"
)

type GithubMirrorService struct {
	repo *repository.GithubMirrorRepository
}

func NewGithubMirrorService() *GithubMirrorService {
	return &GithubMirrorService{
		repo: repository.NewGithubMirrorRepository(),
	}
}

func (s *GithubMirrorService) GetAll() ([]model.GithubMirror, error) {
	return s.repo.GetAll()
}

func (s *GithubMirrorService) GetByID(id uint) (*model.GithubMirror, error) {
	return s.repo.GetByID(id)
}

func (s *GithubMirrorService) GetDefault() (*model.GithubMirror, error) {
	return s.repo.GetDefault()
}

func (s *GithubMirrorService) Create(mirror *model.GithubMirror) error {
	return s.repo.Create(mirror)
}

func (s *GithubMirrorService) Update(mirror *model.GithubMirror) error {
	return s.repo.Update(mirror)
}

func (s *GithubMirrorService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *GithubMirrorService) SetDefault(id uint) error {
	return s.repo.SetDefault(id)
}

func (s *GithubMirrorService) ConvertGithubURL(originalURL string, mirrorID *uint) (string, error) {
	if mirrorID == nil {
		return originalURL, nil
	}

	var mirror *model.GithubMirror
	var err error

	if *mirrorID == 0 {
		mirror, err = s.repo.GetDefault()
		if err != nil {
			return originalURL, nil
		}
	} else {
		mirror, err = s.repo.GetByID(*mirrorID)
		if err != nil {
			return "", fmt.Errorf("加速源不存在")
		}
	}

	if !mirror.Enabled {
		return originalURL, nil
	}

	if !strings.HasPrefix(originalURL, "https://github.com") {
		return originalURL, nil
	}

	convertedURL := strings.Replace(originalURL, "https://github.com", mirror.BaseURL, 1)
	return convertedURL, nil
}
