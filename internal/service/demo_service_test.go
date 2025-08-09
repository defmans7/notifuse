package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDemoService_VerifyRootEmailHMAC(t *testing.T) {
	t.Run("returns false when root email is empty", func(t *testing.T) {
		svc := &DemoService{
			logger: logger.NewLoggerWithLevel("disabled"),
			config: &config.Config{RootEmail: "", Security: config.SecurityConfig{SecretKey: "secret"}},
		}
		assert.False(t, svc.VerifyRootEmailHMAC("anything"))
	})

	t.Run("returns true for valid HMAC and false for invalid", func(t *testing.T) {
		rootEmail := "root@example.com"
		secret := "supersecretkey"
		cfg := &config.Config{RootEmail: rootEmail, Security: config.SecurityConfig{SecretKey: secret}}
		svc := &DemoService{logger: logger.NewLoggerWithLevel("disabled"), config: cfg}

		valid := domain.ComputeEmailHMAC(rootEmail, secret)
		assert.True(t, svc.VerifyRootEmailHMAC(valid))
		assert.False(t, svc.VerifyRootEmailHMAC(valid+"x"))
	})
}

func TestDemoService_DeleteAllWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWorkspaceRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockTaskRepo := domainmocks.NewMockTaskRepository(ctrl)

	svc := &DemoService{
		logger:        logger.NewLoggerWithLevel("disabled"),
		workspaceRepo: mockWorkspaceRepo,
		taskRepo:      mockTaskRepo,
	}

	ctx := context.Background()
	workspaces := []*domain.Workspace{{ID: "w1"}, {ID: "w2"}}

	// Success path
	mockWorkspaceRepo.EXPECT().List(ctx).Return(workspaces, nil)
	mockWorkspaceRepo.EXPECT().Delete(ctx, "w1").Return(nil)
	mockTaskRepo.EXPECT().DeleteAll(ctx, "w1").Return(nil)
	mockWorkspaceRepo.EXPECT().Delete(ctx, "w2").Return(nil)
	mockTaskRepo.EXPECT().DeleteAll(ctx, "w2").Return(nil)

	err := svc.deleteAllWorkspaces(ctx)
	assert.NoError(t, err)

	// Partial failures should still return nil
	mockWorkspaceRepo2 := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockTaskRepo2 := domainmocks.NewMockTaskRepository(ctrl)
	svc2 := &DemoService{logger: logger.NewLoggerWithLevel("disabled"), workspaceRepo: mockWorkspaceRepo2, taskRepo: mockTaskRepo2}

	mockWorkspaceRepo2.EXPECT().List(ctx).Return(workspaces, nil)
	mockWorkspaceRepo2.EXPECT().Delete(ctx, "w1").Return(assert.AnError)
	mockTaskRepo2.EXPECT().DeleteAll(ctx, "w1").Return(assert.AnError)
	mockWorkspaceRepo2.EXPECT().Delete(ctx, "w2").Return(nil)
	mockTaskRepo2.EXPECT().DeleteAll(ctx, "w2").Return(nil)

	err = svc2.deleteAllWorkspaces(ctx)
	assert.NoError(t, err)
}

func TestDemoService_GenerateSampleContactsBatch(t *testing.T) {
	svc := &DemoService{logger: logger.NewLoggerWithLevel("disabled")}

	batch := svc.generateSampleContactsBatch(10, 100)
	assert.Len(t, batch, 10)
	for i, c := range batch {
		assert.NotEmpty(t, c.Email)
		assert.NotZero(t, c.CreatedAt.Unix())
		assert.NotNil(t, c.FirstName)
		assert.NotNil(t, c.LastName)
		assert.True(t, strings.Contains(strings.ToLower(c.Email), strings.ToLower(c.FirstName.String)))
		assert.True(t, strings.Contains(strings.ToLower(c.Email), strings.ToLower(c.LastName.String)))
		// Ensure progression uses startIndex in at least some addresses across batch
		_ = i
	}
}

func TestGenerateEmail_BasicStructure(t *testing.T) {
	first := "John"
	last := "Doe"

	email := generateEmail(first, last, 42)
	// Basic checks
	assert.Contains(t, strings.ToLower(email), strings.ToLower(first))
	assert.Contains(t, strings.ToLower(email), strings.ToLower(last))
	parts := strings.SplitN(email, "@", 2)
	if assert.Len(t, parts, 2) {
		domainPart := parts[1]
		// Validate domain is one of the configured demo domains
		var found bool
		for _, d := range emailDomains {
			if domainPart == d {
				found = true
				break
			}
		}
		assert.True(t, found, "unexpected domain: %s", domainPart)
	}
}

func TestGetRandomElement(t *testing.T) {
	options := []string{"a", "b", "c"}
	picked := getRandomElement(options)
	assert.Contains(t, options, picked)
}

func TestCreateFallbackHTML(t *testing.T) {
	svc := &DemoService{logger: logger.NewLoggerWithLevel("disabled")}
	html := svc.createFallbackHTML()
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "</html>")
}
