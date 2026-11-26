package grpc

import (
	"context"
	"encoding/base64"
	"log/slog"
	"pkg/errlog"
	"pkg/failure"
	"pkg/logger"
	"tsubaki/internal/crypto"
	db "tsubaki/internal/db/sqlc"
	mikov1 "tsubaki/internal/gen/miko/v1"

	"go.opentelemetry.io/otel"
)

func (t *Tsubaki) RiskContext(ctx context.Context, req *mikov1.GetRiskContextRequest) (*mikov1.FraudEnrichmentContextResponse, error) {
	tracer := otel.Tracer("risk_context")

	log := slog.With(slog.String("op", "risk_context"))
	ctx = logger.SetLogger(ctx, log)

	rootCtx, rootSpan := tracer.Start(ctx, "tsubaki.risk.context.start")
	defer rootSpan.End()

	verifyCtx, verifySpan := tracer.Start(rootCtx, "tsubaki.risk.context.verify")
	defer verifySpan.End()
	if req.PayeeHashedIban == "" || req.PayerHashedIban == "" {
		err := failure.EmptyValue("iban_hash")
		err = failure.Wrap("tsubaki.grpc.iban.hash.check", err)
		errlog.Resource(verifyCtx, nil, "iban_hash_check", failure.EmptyValueError, err)
		return nil, err
	}

	normalizeCtx, normalizeSpan := tracer.Start(rootCtx, "tsubaki.risk.context.iban.hash.normalize")
	defer normalizeSpan.End()
	payeeCipherBytes, err := base64.StdEncoding.DecodeString(req.PayeeHashedIban)
	if err != nil {
		err = failure.Wrap("tsubaki.grpc.payee.iban.hash.cipher.decode", err)
		errlog.Resource(normalizeCtx, nil, "payee_iban_hash_cipher_decode", failure.MalformedError, err)
		return nil, err
	}

	payerCipherBytes, err := base64.StdEncoding.DecodeString(req.PayerHashedIban)
	if err != nil {
		err = failure.Wrap("tsubaki.grpc.payer.iban.hash.cipher.decode", err)
		errlog.Resource(normalizeCtx, nil, "payer_iban_hash_cipher_decode", failure.MalformedError, err)
		return nil, err
	}

	payeeRawIban, err := crypto.Decrypt(t.SecretKey, payeeCipherBytes)
	if err != nil {
		err = failure.Wrap("tsubaki.grpc.payee.iban.hash.decrypt", err)
		errlog.Resource(normalizeCtx, nil, "payee_iban_hash_decrypt", failure.MalformedError, err)
		return nil, err
	}
	payerRawIban, err := crypto.Decrypt(t.SecretKey, payerCipherBytes)
	if err != nil {
		err = failure.Wrap("tsubaki.grpc.payer.iban.hash.decrypt", err)
		errlog.Resource(normalizeCtx, nil, "payer_iban_hash_decrypt", failure.MalformedError, err)
		return nil, err
	}

	databaseCtx, databaseSpan := tracer.Start(rootCtx, "tsubaki.risk.context.database.fetch")
	defer databaseSpan.End()
	payeeProfile, err := t.Queries.GetPayeeProfile(databaseCtx, payeeRawIban)
	if err != nil {
		err = failure.Wrap("tsubaki.database.payee.profile.fetch", err)
		errlog.Database(databaseCtx, nil, "payee_profile_fetch", err)
		return nil, err
	}
	payerSecurityContext, err := t.Queries.GetPayerSecurityContext(databaseCtx, payeeRawIban)
	if err != nil {
		err = failure.Wrap("tsubaki.database.payer.security.context.fetch", err)
		errlog.Database(databaseCtx, nil, "payer_security_context_fetch", err)
		return nil, err
	}
	payeeNeighborhood, err := t.Queries.GetPayeeNeighborhood(databaseCtx, db.GetPayeeNeighborhoodParams{
		PayeeIban: payeeRawIban,
		PayerIban: payerRawIban,
	})
	if err != nil {
		err = failure.Wrap("tsubaki.database.payee.neighborhood.fetch", err)
		errlog.Database(databaseCtx, nil, "payee_neighborhood_fetch", err)
		return nil, err
	}

	return &mikov1.FraudEnrichmentContextResponse{
		PayeeProfile: &mikov1.PayeeProfile{
			PayeeIdentifierHash:          req.PayeeHashedIban,
			AccountAgeDays:               payeeProfile.AccountAgeDays,
			IsPep:                        payeeProfile.IsPep,
			UniqueInboundSenders_24H:     int32(payeeProfile.UniqueInboundSenders),
			UniqueOutboundRecipients_24H: int32(payeeProfile.UniqueOutboundReceivers),
			TotalInboundVolume_24HMinor:  payeeProfile.TotalInboundVolume,
			TotalOutboundVolume_24HMinor: payeeProfile.TotalOutboundVolume,
			OutboundToInboundRatio_24H:   float32(payeeProfile.OutboundToInboundRatio),
		},
		PayerSecurity: &mikov1.PayerSecurityContext{
			IsTrustedDevice:        payerSecurityContext.IsTrusted,
			IsProxySession:         payerSecurityContext.IsProxy,
			IpCountryMismatch:      payerSecurityContext.IpCountryMismatch,
			PhoneChangedLast_72H:   payerSecurityContext.PhoneChanged,
			EmailChangedLast_72H:   payerSecurityContext.EmailChanged,
			FailedAuthAttempts_24H: int32(payerSecurityContext.FailedAuthAttempts),
			PasswordResetLast_24H:  payerSecurityContext.PasswordReset,
		},
		PayeeNeighborhood: &mikov1.PayeeNeighborhood{
			CounterpartySummary: &mikov1.CounterpartySummary{
				TotalVolumeTransferredMinor: payeeNeighborhood.TotalVolumeTransferredMinor,
				TransactionCount:            int32(payeeNeighborhood.TransactionCount),
			},
			SharesDeviceWithPayer:   payeeNeighborhood.SharesDeviceWithPayer,
			SharesIpSubnetWithPayer: payeeNeighborhood.SharesIpSubnetWithPayer,
		},
	}, nil
}
