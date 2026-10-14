package app

import (
	"context"
	"strings"
	"time"

	"devenv/internal/domain/site"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type GitService struct {
	sites ports.SiteRepository
	git   ports.GitManager
	bus   ports.EventPublisher
}

func NewGitService(
	sites ports.SiteRepository,
	git ports.GitManager,
	bus ports.EventPublisher,
) *GitService {
	return &GitService{
		sites: sites,
		git:   git,
		bus:   bus,
	}
}

func (s *GitService) Status(ctx context.Context, siteID int64) (*site.GitState, error) {
	cur, err := s.sites.Get(ctx, siteID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	st, err := s.git.Status(ctx, cur.Root)
	if err != nil {
		return nil, err
	}

	if cur.Branch != st.Branch || cur.Commit != st.Commit {
		cur.Branch = st.Branch
		cur.Commit = st.Commit
		if err := s.sites.Update(ctx, cur); err != nil {
			return nil, err
		}
	}

	return st, nil
}

func (s *GitService) Pull(ctx context.Context, siteID int64) error {
	cur, err := s.sites.Get(ctx, siteID)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}
	if !cur.Git {
		return appErrors.New(appErrors.CodeInvalidArgument, "该网站未启用 Git")
	}

	if err := s.git.Pull(ctx, cur.Root); err != nil {
		return err
	}

	st, err := s.git.Status(ctx, cur.Root)
	if err != nil {
		return err
	}
	cur.Branch = st.Branch
	cur.Commit = st.Commit
	cur.PulledAt = time.Now().Format(time.RFC3339)
	if err := s.sites.Update(ctx, cur); err != nil {
		return err
	}

	s.bus.Emit("site:git-pulled", siteID)
	return nil
}

func (s *GitService) Commit(ctx context.Context, siteID int64, message string, files []string) error {
	cur, err := s.sites.Get(ctx, siteID)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}
	msg := strings.TrimSpace(message)
	if msg == "" {
		return appErrors.New(appErrors.CodeInvalidArgument, "提交信息不能为空")
	}

	if err := s.git.Commit(ctx, cur.Root, msg, files); err != nil {
		return err
	}

	st, err := s.git.Status(ctx, cur.Root)
	if err != nil {
		return err
	}
	cur.Commit = st.Commit
	if err := s.sites.Update(ctx, cur); err != nil {
		return err
	}

	s.bus.Emit("site:git-committed", siteID)
	return nil
}

func (s *GitService) Push(ctx context.Context, siteID int64) error {
	cur, err := s.sites.Get(ctx, siteID)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}
	if err := s.git.Push(ctx, cur.Root); err != nil {
		return err
	}
	s.bus.Emit("site:git-pushed", siteID)
	return nil
}
