package reroll

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"pkg/parameter"
	"strconv"
	"time"
	"tsubaki/domain"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel"
)

type KeyRerollWorker struct {
	river.WorkerDefaults[KeyRerollArgs]
	Secret           *domain.RerollStore
	SSMClient        *ssm.Client
	CurrentOperation *parameter.CurrentOperation
}

func (k *KeyRerollWorker) Work(ctx context.Context, job *river.Job[KeyRerollArgs]) error {
	tracer := otel.Tracer("key_reroll")
	rootCtx, rootSpan := tracer.Start(ctx, "tsubaki.key.reroll.start")
	defer rootSpan.End()

	data := k.Secret.Load()
	if data == nil {
		s := make([]byte, 32)
		if _, err := rand.Read(s); err != nil {
			err = failure.Wrap("tsubaki.key.reroll.bytes.generate", err)
			errlog.Resource(rootCtx, nil, "bytes_generate", failure.GenerationError, err)
			return err
		}
		fallbackData := &domain.Reroll{
			PreviousSecretKey: "",
			CurrentSecretKey:  hex.EncodeToString(s),
			Version:           "k0",
			UpdatedAt:         time.Now(),
			ExpiresAt:         time.Now().Unix() - 1,
		}

		if k.Secret.CompareAndSwap(nil, fallbackData) {
			// Data already stored (fallbackData is stored)
			data = fallbackData
		} else {
			// Another goroutine inserted first, we simply give what the goroutine inserted
			data = k.Secret.Load()
		}
	}

	if data.ExpiresAt > time.Now().Unix() {
		return nil
	}

	previousKey := data.CurrentSecretKey
	log := slog.With(slog.String("op", "key_reroll"))
	ctx = logger.SetLogger(ctx, log)

	validateCtx, validateSpan := tracer.Start(rootCtx, "tsubaki.key.reroll.validate")
	defer validateSpan.End()

	if len(data.Version) != 2 {
		err := failure.Invalid("secret_key_version", "invalid secret key version format")
		err = failure.Wrap("tsubaki.key.reroll.secret.key.version.validate", err)
		errlog.Resource(validateCtx, nil, "secret_key_version_check", failure.InvalidError, err)
		return err
	}

	generateCtx, generateSpan := tracer.Start(rootCtx, "tsubaki.key.reroll.generate")
	defer generateSpan.End()

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		err = failure.Wrap("tsubaki.key.reroll.bytes.generate", err)
		errlog.Resource(generateCtx, nil, "bytes_generate", failure.GenerationError, err)
		return err
	}
	newSecretKey := hex.EncodeToString(b)

	upgradeCtx, upgradeSpan := tracer.Start(rootCtx, "tsubaki.key.reroll.upgrade")
	defer upgradeSpan.End()

	upgradeVersion := func(s string) (string, error) {
		b := s[1]
		if b >= '0' && b <= '9' {
			return fmt.Sprintf("k%d", b+1), nil
		}

		i, err := strconv.Atoi(s[1:])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("k%d", i+1), nil
	}

	newVer, err := upgradeVersion(data.Version)
	if err != nil {
		err = failure.Wrap("tsubaki.key.reoll.version.parse", err)
		errlog.Resource(upgradeCtx, nil, "secret_key_version_parse", failure.MalformedError, err)
		return err
	}
	date := time.Now()

	insertCtx, insertSpan := tracer.Start(rootCtx, "tsubaki.key.reroll.ssm.insert")
	defer insertSpan.End()
	err = parameter.PutParameter(k.SSMClient, insertCtx, []parameter.ParamInsert{
		{
			Name:      k.CurrentOperation.FormatParameter("PREVIOUS_KEY"),
			Value:     previousKey,
			Secure:    true,
			Overwrite: true,
		},
		{
			Name:      k.CurrentOperation.FormatParameter("CURRENT_KEY"),
			Value:     newSecretKey,
			Secure:    true,
			Overwrite: true,
		},
		{
			Name:      k.CurrentOperation.FormatParameter("KEY_VERSION"),
			Value:     newVer,
			Secure:    false,
			Overwrite: true,
		},
		{
			Name:      k.CurrentOperation.FormatParameter("KEY_DATE"),
			Value:     date.String(),
			Secure:    false,
			Overwrite: true,
		},
	})
	if err != nil {
		err = failure.Wrap("tsubaki.key.reroll.ssm.secret.key.data.insert", err)
		errlog.Internal(insertCtx, nil, "ssm_secret_key_data_insert", failure.InternalGeneralError, err)
		return err
	}

	k.Secret.Store(&domain.Reroll{
		PreviousSecretKey: data.CurrentSecretKey,
		CurrentSecretKey:  newSecretKey,
		Version:           newVer,
		UpdatedAt:         date,
		ExpiresAt:         time.Now().Add(720 * time.Hour).Unix(),
	})

	return nil
}
