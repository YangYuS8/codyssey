package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/YangYuS8/codyssey/backend/internal/service"
)

func TestJudgeRun_Lifecycle(t *testing.T) {
    repo := repository.NewMemoryJudgeRunRepository()
    svc := service.NewJudgeRunService(repo)
    ctx := context.Background()

    jr, err := svc.Enqueue(ctx, "sub-1", "v1")
    require.NoError(t, err)
    require.Equal(t, domain.JudgeRunStatusQueued, jr.Status)

    jr, err = svc.Start(ctx, jr.ID)
    require.NoError(t, err)
    require.Equal(t, domain.JudgeRunStatusRunning, jr.Status)
    require.NotNil(t, jr.StartedAt)

    jr, err = svc.Finish(ctx, jr.ID, domain.JudgeRunStatusSucceeded, 123, 456, 0, "")
    require.NoError(t, err)
    require.Equal(t, domain.JudgeRunStatusSucceeded, jr.Status)
    require.Equal(t, 123, jr.RuntimeMS)
    require.Equal(t, 456, jr.MemoryKB)
    require.NotNil(t, jr.FinishedAt)
}

func TestJudgeRun_InvalidTransitions(t *testing.T) {
    repo := repository.NewMemoryJudgeRunRepository()
    svc := service.NewJudgeRunService(repo)
    ctx := context.Background()

    jr, err := svc.Enqueue(ctx, "sub-2", "v1")
    require.NoError(t, err)

    // 直接 finish -> 应失败（因为还没 running）
    _, err = svc.Finish(ctx, jr.ID, domain.JudgeRunStatusSucceeded, 10, 10, 0, "")
    require.Error(t, err)

    // 非法终态
    _, err = svc.Finish(ctx, jr.ID, "weird", 0, 0, 0, "")
    require.Error(t, err)

    // 正常 start
    _, err = svc.Start(ctx, jr.ID)
    require.NoError(t, err)

    // 二次 start (已经 running)
    _, err = svc.Start(ctx, jr.ID)
    require.Error(t, err)

    // 正常 finish
    _, err = svc.Finish(ctx, jr.ID, domain.JudgeRunStatusFailed, 1, 2, 42, "boom")
    require.NoError(t, err)

    // 再次 finish (终态后不可再次终结)
    _, err = svc.Finish(ctx, jr.ID, domain.JudgeRunStatusSucceeded, 1, 2, 0, "")
    require.Error(t, err)
}
