package grpc

import (
	db "tsubaki/internal/db/sqlc"
	mikov1 "tsubaki/internal/gen/miko/v1"
)

type Tsubaki struct {
	mikov1.UnimplementedRiskContextServiceServer
	Queries   *db.Queries
	SecretKey string
}

func New(queries *db.Queries, secretKey string) *Tsubaki {
	return &Tsubaki{
		Queries:   queries,
		SecretKey: secretKey,
	}
}
